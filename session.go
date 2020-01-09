package main

import "github.com/defsky/telnet"

import "fmt"

type Session struct {
	nvt telnet.NVT
}

func (s *Session) Send(data []byte)  {
	fmt.Fprint(s.nvt,data)
}