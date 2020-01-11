package main

import (
	"bufio"
	"fmt"
	"strings"
)

type Shell struct {
	session *Session
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
}

func (s *Shell) SendData(data []byte) {
	if s.session != nil {
		s.session.Send(data)
		fmt.Fprintln(screen, string(data))
	} else {
		fmt.Fprintln(screen, "No session. Use /open <host> <port> to open one")
	}
}

func (s *Shell) GetSession() *Session {
	return s.session
}
