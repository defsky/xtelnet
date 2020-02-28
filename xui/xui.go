package xui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"xtelnet/session"

	"github.com/rivo/tview"
)

// UI is the interface that wraps basic methods for user interface
type UI interface {
	io.Reader
	io.Writer
}

// XUI is an extensible UI object
type XUI struct {
	conn    *net.UnixConn
	widgets []tview.Primitive
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

func (ui *XUI) Attach(name string) {
	homedir, err := session.SocketHomeDir()
	if err != nil {
		fmt.Println(err)
		return
	}
	sessions, err := session.GetSessionList(homedir)
	if err != nil {
		fmt.Println(err)
		return
	}

	matchedSession := make([]string, 0)
	for _, s := range sessions {
		if strings.HasSuffix(s, name) {
			matchedSession = append(matchedSession, s)
		}
		if strings.HasPrefix(s, name) {
			matchedSession = append(matchedSession, s)
		}
	}
	if len(matchedSession) == 0 {
		fmt.Printf("There is no matched session for name: %s\n", name)
		return
	}
	if len(matchedSession) > 1 {
		fmt.Println("There are more than one session matched:")
		for _, s := range matchedSession {
			fmt.Printf("  %s\n", s)
		}
		return
	}

	fpath := filepath.Join(homedir, matchedSession[0])

	sessionAddr, err := net.ResolveUnixAddr("unix", fpath)
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.DialUnix("unix", nil, sessionAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	ui.conn = conn

	go ui.receiver(conn)
	go ui.sender(conn)

	ui.run()
}

func (ui *XUI) sender(conn *net.UnixConn) {
	w := bufio.NewWriter(conn)
DONE:
	for {
		select {
		case cmd, ok := <-inputCh:
			if !ok {
				ui.conn.Close()
				break DONE
			}
			_, err := w.Write(cmd)
			if err != nil {
				fmt.Fprintln(screen, err)
			}
			err = w.Flush()
			if err != nil {
				fmt.Fprintln(screen, err)
			}
		}
	}
}

func (ui *XUI) receiver(conn *net.UnixConn) {
	ansiW := tview.ANSIWriter(screen)

	r := bufio.NewReader(conn)
	for {
		b, err := r.ReadString('\n')
		if err != nil {
			break
		}

		fmt.Fprint(ansiW, b)
	}

	app.Stop()
}

func (ui *XUI) run() error {

	if err := app.SetRoot(layout, true).Run(); err != nil {
		panic(err)
	}

	return nil
}
