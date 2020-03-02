package session

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/defsky/xtelnet/telnet"
)

const socketRunDir string = "run"

var outCh = make(chan []byte, 100)
var closeCh = make(chan struct{})

var nvtConfig = &telnet.SessionOption{
	NVTOptionCfg: telnet.NewNVTOptionConfig(),
}
var nvt *telnet.NVT

type Session struct {
	name  string
	term  *Terminal
	ln    net.Listener
	fd    *os.File
	fname string
}

func NewSession(name, fname string) *Session {
	return &Session{
		name: name,
		term: NewTerminal(),
	}
}
func (s *Session) Start() {
	s.term.Start()

	fname, err := socketFileName(s.name)
	if err != nil {
		return
	}
	s.fname = fname
	addr, err := net.ResolveUnixAddr("unix", fname)
	if err != nil {
		return
	}
	l, err := net.ListenUnix("unix", addr)
	if err != nil {
		return
	}
	l.SetUnlinkOnClose(false)
	fd, err := l.File()
	if err != nil {
		return
	}
	l.Close()

	s.fd = fd

	go s.listenUnixSocket()

	<-closeCh
	s.term.Stop()
	if s.ln != nil {
		s.ln.Close()
	}
	if s.fd != nil {
		s.fd.Close()
	}
	if s.fname != "" {
		os.Remove(s.fname)
	}
}

func (s *Session) listenUnixSocket() {
	l, err := net.FileListener(s.fd)
	if err != nil {
		// fmt.Println(err)
		return
	}
	s.ln = l

	conn, err := l.Accept()
	if err != nil {
		return
	}
	l.Close()

	go s.listenUnixSocket()

	s.term.HandleIncoming(conn)
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
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	baseDir := filepath.Join(userHome, cacheDir)
	err = mkdirIfNotExist(baseDir, os.ModeDir|os.FileMode(0775))
	if err != nil {
		return "", err
	}

	baseDir = filepath.Join(baseDir, socketRunDir)
	err = mkdirIfNotExist(baseDir, os.ModeDir|os.FileMode(0775))
	if err != nil {
		return "", err
	}

	// currentUser, err := user.Current()
	// if err != nil {
	// 	return "", err
	// }
	// homedirName := fmt.Sprintf("S-%s", currentUser.Username)
	// dirname := filepath.Join(baseDir, homedirName)

	// err = mkdirIfNotExist(dirname, os.ModeDir|os.FileMode(0700))
	// if err != nil {
	// 	return "", err
	// }
	return baseDir, nil
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
