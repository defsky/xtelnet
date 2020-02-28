package session

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"xtelnet/telnet"
)

const baseDir string = "/var/run/xtelnet"

var outCh = make(chan []byte, 100)
var closeCh = make(chan struct{})

var nvt *telnet.NVT
var nvtConfig = &telnet.SessionOption{
	NVTOptionCfg: telnet.NewNVTOptionConfig(),
}

func Create(name string, fname string) {
	shell := NewShell()
	term := NewTerminal(shell)

	term.Start()

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
	defer l.Close()

	l.SetUnlinkOnClose(true)

	//os.Chmod(fname, os.ModeSocket|os.FileMode(0600))
	go func() {
		<-closeCh
		l.Close()
	}()

DONE:
	for {
		select {
		case <-closeCh:
			break DONE
		default:
			conn, err := l.AcceptUnix()
			if err != nil {
				break
			}
			go handleIncoming(conn, term)
		}
	}
}

func handleIncoming(conn *net.UnixConn, term *Terminal) {
	err := sendBufferedMessage(conn, term)
	if err != nil {
		return
	}
	term.SetConn(conn)
	defer term.SetConn(nil)

	r := bufio.NewReader(conn)
DONE:
	for {
		select {
		case <-closeCh:
			break DONE
		default:
			b, err := r.ReadBytes('\n')
			if err != nil {
				break DONE
			}
			term.Input(b)
		}
	}
}

func sendBufferedMessage(conn *net.UnixConn, term *Terminal) error {
	lines := term.GetBufferdLines(25)

	w := bufio.NewWriter(conn)
	if lines != nil && len(lines) > 0 {
		for _, l := range lines {
			if len(l) > 0 {
				_, err := w.Write(l)
				if err != nil {
					return err
				}
			}
		}
	} else {
		_, err := w.Write([]byte("No buffered message\n"))
		if err != nil {
			return err
		}
		w.Flush()
	}

	err := w.Flush()
	if err != nil {
		return err
	}
	return nil
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
