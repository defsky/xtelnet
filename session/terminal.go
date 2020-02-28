package session

import (
	"bufio"
	"net"
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
	wg      sync.WaitGroup
	history *HistoryCmd
	shell   *Shell
	in      chan []byte
	out     chan []byte

	buffer     [][]byte
	netWriter  *bufio.Writer
	closeTimer chan struct{}
}

func NewTerminal(s *Shell) *Terminal {
	return &Terminal{
		history:    NewHistoryCmd(historyCmdLength),
		shell:      s,
		buffer:     make([][]byte, 0, 100),
		out:        outCh,
		closeTimer: make(chan struct{}),
	}
}

func (t *Terminal) Start() {
	go t.terminal()

	t.out <- []byte("[green]Welcome to xtelnet!\n\n")
	t.out <- []byte("[yellow]Type /<Enter> for help\n[-]")
}

func (t *Terminal) terminal() {
DONE:
	for {
		select {
		case msg, ok := <-t.out:
			if !ok {
				break DONE
			}
			t.buffer = append(t.buffer, msg)

			if len(t.buffer) > 10000 {
				t.buffer = t.buffer[1:]
			}
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

	size := len(t.buffer)
	ret := make([][]byte, 0, count)

	startIdx := 0
	if size > count {
		startIdx = size - count
	}
	copy(ret, t.buffer[startIdx:])

	return ret
}
func (t *Terminal) SetConn(c *net.UnixConn) {
	if c != nil {
		t.netWriter = bufio.NewWriter(c)
	} else {
		t.netWriter = nil
	}
}

func (t *Terminal) Input(cmd []byte) {
	msg, data, err := t.shell.Exec(string(cmd))
	if len(msg) > 0 {
		t.out <- []byte(msg)
	}
	if err != nil {
		t.out <- []byte(err.Error())
	}
	if len(data) > 0 && false == nvt.Send(data) {
		t.out <- []byte("no active conncetion\n")
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
