package main

import (
	"bufio"
	"fmt"
	"strings"
)

type Shell struct {
	session *Session
}

var UserShell = NewShell()

func NewShell() *Shell {
	s := &Shell{}
	s.SetSession(nil)
	return s
}

func (s *Shell) Exec(cmd string) (string, error) {
	if len(cmd) <= 0 || cmd[0] != '/' {
		s.SendData([]byte(cmd + "\r\n"))
		return "", nil
	}
	rd := bufio.NewReader(strings.NewReader(cmd[1:]))
	rd.Peek(1)
	msg, data, err := rootCMD.Exec(rd)
	if len(data) > 0 {
		s.SendData(data)
	}
	return msg, err
}

func (s *Shell) SetSession(sess *Session) {
	s.session = sess
	if sess == nil {
		statusBar.GetCell(0, 0).SetText("No active session")
	} else {
		statusBar.GetCell(0, 0).SetText(s.session.host)
	}
}

func (s *Shell) SendData(data []byte) {
	if s.session != nil {
		ok := s.session.Send(data)
		if !ok {
			fmt.Fprintln(screen, "send failed ...")
			UserShell.SetSession(nil)
		}
	} else {
		fmt.Fprintln(screen, "No session. Use /open <host> <port> to open one")
	}
}

func (s *Shell) GetSession() *Session {
	return s.session
}
