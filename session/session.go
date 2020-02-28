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

var nvt = telnet.NewNVT()
var outCh = make(chan []byte, 100)

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

	os.Chmod(fname, os.ModeSocket|os.FileMode(0600))

	for {
		conn, err := l.AcceptUnix()
		if err != nil {
			conn.Close()
			continue
		}

		go handleIncoming(conn, term)
	}
}

func handleIncoming(conn *net.UnixConn, term *Terminal) {
	defer conn.Close()

	err := sendBufferedMessage(conn, term)
	if err != nil {
		return
	}
	term.SetConn(conn)

	r := bufio.NewReader(conn)
	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			term.SetConn(nil)
			break
		}
		term.Input(b)
	}
}

func sendBufferedMessage(conn *net.UnixConn, term *Terminal) error {
	lines := term.GetBufferdLines(80)
	w := bufio.NewWriter(conn)

	for _, l := range lines {
		_, err := w.Write(l)
		if err != nil {
			return err
		}
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
