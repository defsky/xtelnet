package session

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"time"
	"github.com/defsky/xtelnet/proto"
)

type TaskType int

const historyCmdLength = 1000

const (
	Timer TaskType = iota
	Ticker
)

type TaskHandler func()
type ScheduleTask struct {
	ttype   TaskType
	handler TaskHandler
}
type TickTaskMap map[time.Duration][]ScheduleTask

// Terminal is the interface wraps basic methods for terminal
type Terminal struct {
	wg         sync.WaitGroup
	history    *HistoryCmd
	shell      *Shell
	conn       *net.UnixConn
	buffer     *OutBuffer
	netWriter  *bufio.Writer
	close      chan struct{}
	closeTimer chan struct{}
}

func NewTerminal() *Terminal {
	return &Terminal{
		history:    NewHistoryCmd(historyCmdLength),
		shell:      NewShell(),
		buffer:     NewBuffer(500),
		close:      make(chan struct{}),
		closeTimer: make(chan struct{}),
	}
}

func (t *Terminal) Start() {
	go t.terminal()

	outCh <- []byte("[green]Welcome to xtelnet!\n\n")
	outCh <- []byte("[yellow]Type /<Enter> for help\n[-]")
}

func (t *Terminal) Stop() {
	if nvt != nil {
		nvt.Close()
	}
	if t.conn != nil {
		t.conn.Close()
	}
	close(t.close)
}

func (t *Terminal) terminal() {
DONE:
	for {
		select {
		case <-t.close:
			break DONE
		case msg, ok := <-outCh:
			if !ok {
				break DONE
			}

			t.buffer.Put(msg)

			if t.conn != nil {
				p := &proto.Packet{}
				p.Write(msg)

				proto.Send(t.conn, p)
			}
		}
	}
}

func (t *Terminal) GetBufferdLines(count int) [][]byte {
	if count <= 0 {
		return nil
	}

	return t.buffer.Get(count)
}
func (t *Terminal) SetConn(c *net.UnixConn) {
	t.conn = c
}
func (t *Terminal) sendFirstScreenData(conn *net.UnixConn) error {
	p := &proto.Packet{}

	lines := t.GetBufferdLines(25)
	if lines != nil && len(lines) > 0 {
		for _, l := range lines {
			if len(l) > 0 {
				_, err := p.Write(l)
				if err != nil {
					p.WriteString(err.Error())
					break
				}
			}
		}
	} else {
		_, err := p.WriteString("No buffered message\n")
		if err != nil {
			p.WriteString(err.Error())
		}
	}

	return proto.Send(conn, p)
}

func (t *Terminal) HandleIncoming(conn *net.UnixConn) {
	defer conn.Close()

	err := t.sendFirstScreenData(conn)
	if err != nil {
		return
	}
	t.conn = conn
	defer func() {
		t.conn = nil
	}()

	r := bufio.NewReader(conn)
DONE:
	for {
		b, err := r.ReadBytes('\n')
		if err != nil {
			break DONE
		}
		t.Input(b)
	}
}

func (t *Terminal) Input(cmd []byte) {
	msg, data, err := t.shell.Exec(strings.TrimRight(string(cmd), "\r\n"))
	if len(msg) > 0 {
		outCh <- []byte(msg + "\n")
	}
	if err != nil {
		outCh <- []byte(err.Error() + "\n")
	}
	if len(data) > 0 {
		if nvt == nil || false == nvt.Send(data) {
			outCh <- []byte("no active conncetion\n")
		}
	}
}

// RunAfter wil call f only once when d duration elapsed
func (s *Terminal) RunAfter(d time.Duration, f func()) {
	timer := time.NewTimer(d)

	s.wg.Add(1)
	go func(f func()) {
		defer s.wg.Done()
		defer timer.Stop()

		select {
		case <-timer.C:
			f()
		case <-s.closeTimer:
			// close timer
		}

	}(f)
}

// RunEvery will call f periodly when every d duaration elapsed
func (s *Terminal) RunEvery(d time.Duration, f func()) {
	ticker := time.NewTicker(d)

	s.wg.Add(1)
	go func(f func()) {
		defer s.wg.Done()
		defer ticker.Stop()
	DONE:
		for {
			select {
			case <-ticker.C:
				f()
			case <-s.closeTimer:
				break DONE
			}
		}
	}(f)
}
