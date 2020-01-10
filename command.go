package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
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

func handleCmdOpen(c *Command, p *bufio.Reader) (string, []byte, error) {
	if p.Buffered() <= 0 {
		return c.help, nil, errors.New("need params: <host> <port>")
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
		return c.help, nil, errors.New("need param: <port>")
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return "", nil, errors.New("port param must be a number")
	}
	if portNumber < 0 || portNumber > 65535 {
		return "", nil, errors.New("port number must in range 1-65535")
	}

	// Todo: add operations to open connection to remote host
	sess, err := NewSession(fmt.Sprintf("%s:%s", host, port), screen)
	if err != nil {
		fmt.Fprintln(screen, err)
	}
	UserShell.SetSession(sess)
	go sess.Start()

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
		return subCmd.Exec(p)
	}
	return subCmdDesc(c), nil, nil
}

func subCmdDesc(c *Command) string {
	msg := c.desc + ":\n"
	for k, v := range c.subCommand {
		msg = msg + fmt.Sprintf("\t%-10s%-50s\n", k, v.desc)
	}
	strings.TrimRight(msg, "\r\n ")
	return msg
}

func doCommand(cmd string) (string, []byte, error) {
	if len(cmd) <= 0 || cmd[0] != '/' {
		return "", []byte(cmd + "\r\n"), nil
	}
	rd := bufio.NewReader(strings.NewReader(cmd[1:]))
	rd.Peek(1)
	return rootCMD.Exec(rd)
}

var rootCMD = &Command{
	name:       "",
	handler:    nil,
	subCommand: commands,
	desc:       "Available commands",
}

var commands = CommandMap{
	"open": &Command{
		name:       "open",
		handler:    handleCmdOpen,
		subCommand: nil,
		desc:       "Open a connection",
		help:       "\tUsage: /open <host> <port>",
	},
}
