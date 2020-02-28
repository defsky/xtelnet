package session

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"github.com/defsky/xtelnet/telnet"

	"github.com/takama/daemon"
)

const baseDir string = "/var/run/xtelnet"

var outCh = make(chan []byte, 100)
var closeCh = make(chan struct{})

var nvtConfig = &telnet.SessionOption{
	NVTOptionCfg: telnet.NewNVTOptionConfig(),
}
var nvt *telnet.NVT

func create(name string, fname string) {
	s, err := daemon.New(name, name)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.Start()
}
func Create(name string, fname string) {
	fname, err := socketFileName(name)
	if err != nil {
		return
	}

	unixAddr, err := net.ResolveUnixAddr("unix", fname)
	if err != nil {
		fmt.Println(err)
	}
	l, err := net.ListenUnix("unix", unixAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	l.SetUnlinkOnClose(true)

	term := NewTerminal()
	term.Start()
	go func() {
		<-closeCh
		term.Stop()
		l.Close()
	}()

	for {
		conn, err := l.AcceptUnix()
		if err != nil {
			break
		}
		term.HandleIncoming(conn)
	}
}

func mkdirIfNotExist(path string, mode os.FileMode) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, mode)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

func socketFileName(name string) (string, error) {
	homedir, err := SocketHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%d.%s", homedir, os.Getpid(), name), nil
}

func SocketHomeDir() (string, error) {
	err := mkdirIfNotExist(baseDir, os.ModeDir|os.FileMode(0775))
	if err != nil {
		return "", err
	}

	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	homedirName := fmt.Sprintf("S-%s", currentUser.Username)
	dirname := filepath.Join(baseDir, homedirName)

	err = mkdirIfNotExist(dirname, os.ModeDir|os.FileMode(0700))
	if err != nil {
		return "", err
	}
	return dirname, nil
}

func GetSessionList(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}
	for i, v := range files {
		_, n := filepath.Split(v)
		files[i] = n
	}
	return files, nil
}

func GetRootCmd() *Command {
	return rootCMD
}
