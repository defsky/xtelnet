package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gdamore/tcell"
	"io"
	"sort"
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

var rootCMD = &Command{
	name:       "",
	handler:    nil,
	subCommand: commands,
	desc:       "Available commands",
}

var debugSubCommands = CommandMap{
	"color": &Command{
		name:       "color",
		handler:    handleCmdDebugColor,
		subCommand: nil,
		desc:       "switch color debug",
		help:       "\tUsage /debug color",
	},
	"ansicolor": &Command{
		name:       "ansicolor",
		handler:    handleCmdDebugAnsiColor,
		subCommand: nil,
		desc:       "switch ansi color debug",
		help:       "\tUsage /debug ansicolor",
	},
	"iac": &Command{
		name:       "iac",
		handler:    handleCmdDebugIAC,
		subCommand: nil,
		desc:       "switch iac debug",
		help:       "\tUsage /debug iac",
	},
}
var commands = CommandMap{
	"open": &Command{
		name:       "/open",
		handler:    handleCmdOpen,
		subCommand: nil,
		desc:       "Open a session",
		help:       "\tUsage: /open <host> <port>",
	},
	"close": &Command{
		name:       "/close",
		handler:    handleCmdClose,
		subCommand: nil,
		desc:       "Close a session, equivalent to Ctrl-d",
		help:       "\tUsage: /close",
	},
	"debug": &Command{
		name:       "/debug",
		handler:    nil,
		subCommand: debugSubCommands,
		desc:       "Switches for debug",
		help:       "\tUsage: /debug",
	},
	"quit": &Command{
		name:       "/quit",
		handler:    handleCmdQuit,
		subCommand: nil,
		desc:       "Quit this terminal, equivalent to Ctrl-c",
		help:       "\tUsage: /quit",
	},
}

func handleCmdQuit(c *Command, p *bufio.Reader) (string, []byte, error) {
	app.QueueEvent(tcell.NewEventKey(tcell.KeyCtrlC, rune('c'), tcell.ModCtrl))
	return "", nil, nil
}
func handleCmdDebugIAC(c *Command, p *bufio.Reader) (string, []byte, error) {
	if sess := UserShell.GetSession(); sess != nil {

		iacDebug := !sess.Option.DebugIAC
		sess.Option.DebugIAC = iacDebug
		if iacDebug {
			return "IAC debug opened", nil, nil
		} else {
			return "IAC debug closed", nil, nil
		}
	}
	return "No active session", nil, nil
}

func handleCmdDebugColor(c *Command, p *bufio.Reader) (string, []byte, error) {
	if sess := UserShell.GetSession(); sess != nil {

		colorDebug := !sess.Option.DebugColor
		screen.SetDynamicColors(!colorDebug)
		sess.Option.DebugColor = colorDebug
		if colorDebug {
			return "Color debug opened", nil, nil
		} else {
			return "Color debug closed", nil, nil
		}
	}
	return "No active session", nil, nil
}
func handleCmdDebugAnsiColor(c *Command, p *bufio.Reader) (string, []byte, error) {
	if sess := UserShell.GetSession(); sess != nil {

		sess.Option.DebugAnsiColor = !sess.Option.DebugAnsiColor

		if sess.Option.DebugAnsiColor {
			return "Ansi Color debug opened", nil, nil
		} else {
			return "Ansi Color debug closed", nil, nil
		}
	}
	return "No active session", nil, nil
}
func handleCmdClose(c *Command, p *bufio.Reader) (string, []byte, error) {
	if sess := UserShell.GetSession(); sess != nil {
		sess.Close()
	}
	return "No active session", nil, nil
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

	// open session to remote host
	fmt.Fprintf(screen, "connecting to %s:%s ...\n", host, port)
	sess, err := NewSession(fmt.Sprintf("%s:%s", host, port), screen)
	if err != nil {
		fmt.Fprintln(screen, err)
		return "", nil, err
	}

	UserShell.SetSession(sess)

	return "connection established", nil, nil
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
	cmdNames := []string{}
	for n, _ := range c.subCommand {
		cmdNames = append(cmdNames, n)
	}
	sort.Strings(cmdNames)
	for _, name := range cmdNames {
		msg = msg + fmt.Sprintf("\t%-10s%-50s\n", c.subCommand[name].name, c.subCommand[name].desc)
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
