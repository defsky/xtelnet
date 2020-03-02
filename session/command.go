package session

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/defsky/xtelnet/telnet"
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

func (c *Command) Name() string {
	return c.name
}

func (c *Command) GetCommandMap() CommandMap {
	return c.subCommand
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
		help:       "\tUsage: /debug color",
	},
	"ansicolor": &Command{
		name:       "ansicolor",
		handler:    handleCmdDebugAnsiColor,
		subCommand: nil,
		desc:       "switch ansi color debug",
		help:       "\tUsage: /debug ansicolor",
	},
	"iac": &Command{
		name:       "iac",
		handler:    handleCmdDebugIAC,
		subCommand: nil,
		desc:       "switch iac debug",
		help:       "\tUsage: /debug iac",
	},
}
var setSubCommands = CommandMap{
	"GA": &Command{
		name:       "GA",
		handler:    handleCmdSetGA,
		subCommand: nil,
		desc:       "switch GA visibility",
		help:       "\t Usage: /set GA",
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
		desc:       "debug switches",
		help:       "\tUsage: /debug",
	},
	"set": &Command{
		name:       "/set",
		handler:    nil,
		subCommand: setSubCommands,
		desc:       "subcommands for setting",
		help:       "\tUsage: /set",
	},
	"exit": &Command{
		name:       "/exit",
		handler:    handleCmdExit,
		subCommand: nil,
		desc:       "exit daemon of this session",
		help:       "\tUsage: /exit",
	},
	"detach": &Command{
		name:       "/detach",
		handler:    handleCmdDetach,
		subCommand: nil,
		desc:       "Detach from this session, equivalent to Ctrl-c",
		help:       "\tUsage: /detach",
	},
}

func handleCmdExit(c *Command, p *bufio.Reader) (string, []byte, error) {
	close(closeCh)

	return "", nil, nil
}
func handleCmdDetach(c *Command, p *bufio.Reader) (string, []byte, error) {
	// app.QueueEvent(tcell.NewEventKey(tcell.KeyCtrlC, rune('c'), tcell.ModCtrl))

	return "", nil, errors.New("\nPress CTRL-C to Detach\n")
}
func handleCmdSetGA(c *Command, p *bufio.Reader) (string, []byte, error) {
	gaVisible := !nvtConfig.GAVisible
	nvtConfig.GAVisible = gaVisible
	if gaVisible {
		return "GA visible on", nil, nil
	} else {
		return "GA visible off", nil, nil
	}
}

func handleCmdDebugIAC(c *Command, p *bufio.Reader) (string, []byte, error) {

	iacDebug := !nvtConfig.DebugIAC
	nvtConfig.DebugIAC = iacDebug
	if iacDebug {
		return "IAC debug opened", nil, nil
	} else {
		return "IAC debug closed", nil, nil
	}

}

func handleCmdDebugColor(c *Command, p *bufio.Reader) (string, []byte, error) {

	colorDebug := !nvtConfig.DebugColor
	// screen.SetDynamicColors(!colorDebug)
	nvtConfig.DebugColor = colorDebug
	if colorDebug {
		return "Color debug opened", nil, nil
	} else {
		return "Color debug closed", nil, nil
	}

}
func handleCmdDebugAnsiColor(c *Command, p *bufio.Reader) (string, []byte, error) {
	nvtConfig.DebugAnsiColor = !nvtConfig.DebugAnsiColor

	if nvtConfig.DebugAnsiColor {
		return "Ansi Color debug opened", nil, nil
	} else {
		return "Ansi Color debug closed", nil, nil
	}
}
func handleCmdClose(c *Command, p *bufio.Reader) (string, []byte, error) {

	if nvt != nil {
		nvt.Close()
		return "", nil, nil
	}
	return "No active connection", nil, nil
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

	go func() {
		nvt = telnet.NewNVT(outCh, host, port, nvtConfig)
	}()

	return fmt.Sprintf("connecting to %s:%s ...", host, port), nil, nil
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
