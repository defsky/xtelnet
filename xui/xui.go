package xui

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/defsky/xtelnet/proto"
	"github.com/defsky/xtelnet/session"

	"github.com/rivo/tview"
)

// UI is the interface that wraps basic methods for user interface
type UI interface {
	io.Reader
	io.Writer
}

// XUI is an extensible UI object
type XUI struct {
	conn        *net.UnixConn
	widgets     []tview.Primitive
	stopMsg     string
	sessionName string
}

// Read data into p
func (ui *XUI) Read(p []byte) (int, error) {
	return 0, errors.New("method not defined")
}

// Write p into XUI
func (ui *XUI) Write(p []byte) (int, error) {
	return 0, errors.New("method not defined")
}

// NewXUI create new XUI
func NewXUI() *XUI {
	return &XUI{}
}

func getSessionName(name string) (string, error) {
	homedir, err := session.SocketHomeDir()
	if err != nil {
		return "", err
	}
	sessions, err := session.GetSessionList(homedir)
	if err != nil {
		return "", err
	}

	matchedSession := make([]string, 0)
	for _, s := range sessions {
		if strings.HasSuffix(s, name) {
			matchedSession = append(matchedSession, s)
			continue
		}
		if strings.HasPrefix(s, name) {
			matchedSession = append(matchedSession, s)
		}
	}
	if len(matchedSession) == 0 {
		return "", errors.New(fmt.Sprintf("There is no matched session for name: %s", name))
	}
	if len(matchedSession) > 1 {
		msg := "There are more than one session matched:\n"
		for _, s := range matchedSession {
			msg = msg + fmt.Sprintf("\t%s\n", s)
		}
		return "", errors.New(msg)
	}

	return matchedSession[0], nil
}

// Attach will attach to specified session
//  name  : string, session name
//  detach: bool, if detach other
func (ui *XUI) Attach(name string, detach bool) {
	defer func() {
		if len(ui.stopMsg) > 0 {
			fmt.Printf("  %s\n", ui.stopMsg)
		}
	}()

	s, err := getSessionName(name)
	if err != nil {
		ui.stopMsg = fmt.Sprintf("%s", err.Error())
		return
	}
	ui.sessionName = s
	homedir, err := session.SocketHomeDir()
	if err != nil {
		ui.stopMsg = fmt.Sprintf("%s", err.Error())
		return
	}
	fpath := filepath.Join(homedir, s)

	sessionAddr, err := net.ResolveUnixAddr("unix", fpath)
	if err != nil {
		ui.stopMsg = fmt.Sprintf("%s", err.Error())
		return
	}

	conn, err := net.DialUnix("unix", nil, sessionAddr)
	if err != nil {
		os.Remove(fpath)
		ui.stopMsg = fmt.Sprintf("%s", err.Error())
		return
	}
	defer conn.Close()

	p := &proto.Packet{}
	p.Opcode = proto.CM_ATTACH_REQ
	if detach {
		p.WriteByte(byte(1))
	} else {
		p.WriteByte(byte(0))
	}
	if err := proto.WritePacket(conn, p); err != nil {
		ui.stopMsg = fmt.Sprintf("Attach error: %s", err.Error())
		return
	}
	p2, err := proto.ReadPacket(conn)
	if err != nil {
		ui.stopMsg = fmt.Sprintf("Attach error: %s", err.Error())
		return
	}

	switch p2.Opcode {
	case proto.SM_ATTACH_ACK:
		b, err := p2.ReadByte()
		if err != nil {
			ui.stopMsg = fmt.Sprintf("Attach error: %s", err.Error())
			return
		}
		ret := uint8(b)
		if ret != 1 {
			ui.stopMsg = fmt.Sprintf("Attaching denied: %s", p2.String())
			return
		}
	default:
		ui.stopMsg = fmt.Sprintf("Attach error: %s", "unknown ack code")
		return
	}
	ui.conn = conn

	go ui.receiver()
	go ui.sender()

	ui.run()
}

func (ui *XUI) sender() {
	defer ui.conn.Close()

DONE:
	for {
		select {
		case cmd, ok := <-inputCh:
			if !ok {
				break DONE
			}
			p := &proto.Packet{}
			p.Opcode = proto.CM_USER_INPUT
			p.Write(cmd)
			err := proto.WritePacket(ui.conn, p)
			// _, err := ui.conn.Write(cmd)
			if err != nil {
				fmt.Fprintln(screen, err)
				break DONE
			}
		}
	}
}

func (ui *XUI) receiver() {
	defer ui.conn.Close()

	ansiW := tview.ANSIWriter(screen)

	// r := bufio.NewReader(ui.conn)
DONE:
	for {
		// b, err := r.ReadString('\n')
		// if err != nil {
		// 	break
		// }
		p, err := proto.ReadPacket(ui.conn)
		if err != nil {
			if err == proto.EInvalidPacket {
				fmt.Fprintf(ansiW, "[red]%s\n[-]", err.Error())
				continue
			}
			if err == io.EOF {
				ui.stopMsg = fmt.Sprintf("Remotely detached from session: %s", ui.sessionName)
			} else {
				ui.stopMsg = fmt.Sprintf("Detached from session: %s", ui.sessionName)
			}
			break
		}

		switch p.Opcode {
		case proto.SM_DETACH_STATUS:
			status, err := p.ReadByte()
			if err == nil {
				if uint8(status) == 0 {
					ui.stopMsg = fmt.Sprintf("Session already attached: %s", ui.sessionName)
					break DONE
				}
			}
		default:
			fmt.Fprint(ansiW, p.String())
		}

	}

	app.Stop()
	// close(inputCh)
	// app.QueueEvent(tcell.NewEventKey(tcell.KeyCtrlC, rune('c'), tcell.ModCtrl))
}

func (ui *XUI) run() error {
	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}

	return nil
}
