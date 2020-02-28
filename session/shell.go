package session

import (
	"bufio"
	"strings"

	"github.com/defsky/xtelnet/telnet"
)

type Shell struct {
	nvt *telnet.NVT
}

func NewShell() *Shell {
	s := &Shell{}
	return s
}

func (s *Shell) Exec(cmd string) (string, []byte, error) {
	if len(cmd) <= 0 || cmd[0] != '/' {
		return "", []byte(cmd + "\r\n"), nil
	}
	rd := bufio.NewReader(strings.NewReader(cmd[1:]))
	rd.Peek(1)

	return rootCMD.Exec(rd)
}
