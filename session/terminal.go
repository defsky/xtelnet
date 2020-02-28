package session

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"time"
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

func NewTerminal(s *Shell) *Terminal {
	return &Terminal{
		history:    NewHistoryCmd(historyCmdLength),
		shell:      s,
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
	close(t.close)

	if nvt != nil {
		nvt.Close()
	}

	if t.conn != nil {
		t.conn.Close()
	}
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

			if t.netWriter != nil {
				t.netWriter.Write(msg)
				t.netWriter.Flush()
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
	if c != nil {
		t.conn = c
		t.netWriter = bufio.NewWriter(c)
	} else {
		t.netWriter = nil
		t.conn.Close()
		t.conn = nil
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
	if len(data) > 0 && nvt != nil && false == nvt.Send(data) {
		outCh <- []byte("no active conncetion\n")
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
