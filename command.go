package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// CommandHandler handle command
type CommandHandler func(*Command, *bufio.Reader) (string, []byte, error)

// Command
type CommandMap map[string]*Command

// Command user command struct
type Command struct {
	name       string
	handler    CommandHandler
	subCommand CommandMap
	desc       string
	help       string
}

var Commands = CommandMap{
	"open": &Command{
		name:       "open",
		handler:    handleCmdOpen,
		subCommand: nil,
		desc:       "Open a connection",
		help:       "Usage: /open <host> <port>",
	},
}

var Shell = &Command{
	name:       "",
	handler:    nil,
	subCommand: Commands,
	desc:       "Client Command",
}

func handleCmdOpen(c *Command, p *bufio.Reader) (string, []byte, error) {
	if p.Buffered() <= 0 {
		return c.help, nil, nil
	}

	var host, port string
	var err error
	host, err = p.ReadString(' ')
	if err != nil && err != io.EOF {
		return "", nil, err
	}
	host = strings.TrimRight(host, " ")

	if err != io.EOF {
		port, err = p.ReadString(' ')
		if err != nil && err != io.EOF {
			return "", nil, err
		}
		port = strings.TrimRight(port, " ")
	}

	if len(port) == 0 {
		return "", nil, errors.New("need port param")
	}

	// Todo: add operations to open connection to remote host
	return fmt.Sprintf("connect to %s:%s ...", host, port), nil, nil
}

func (c *Command) Exec(p *bufio.Reader) (string, []byte, error) {
	if c.handler != nil {
		return c.handler(c, p)
	}
	if c.subCommand == nil {
		return "", nil, errors.New("Unhandled command: " + c.name)
	}

	if p.Buffered() > 0 {
		cmdName, err := p.ReadString(' ')
		if err != nil && err != io.EOF {
			return "", nil, err
		}
		cmdName = strings.TrimRight(cmdName, " ")
		subCmd, ok := c.subCommand[cmdName]
		if !ok {
			return subCmdDesc(c), nil, errors.New(fmt.Sprintf("command not found: %s", cmdName))
		}
		subCmd.Exec(p)
	}
	return fmt.Sprintf("bufferd size:%d", p.Buffered()), nil, nil
}

func subCmdDesc(c *Command) string {
	msg := c.desc + ":\n"
	for k, v := range c.subCommand {
		msg = msg + fmt.Sprintf("%-10s%-50s\n", k, v.desc)
	}
	strings.TrimRight(msg, "\r\n ")
	return msg
}
