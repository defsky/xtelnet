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
	conn       net.Conn
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
	outCh <- []byte("[green]Presss Ctrl-C to detach\n\n")
	outCh <- []byte("[yellow]Type /<Enter> for help[-]\n\n")
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

				proto.WritePacket(t.conn, p)
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
func (t *Terminal) sendFirstScreenData(conn net.Conn) error {
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

	return proto.WritePacket(conn, p)
}

func (t *Terminal) sendDetachStatus(c net.Conn) error {
	p := &proto.Packet{}
	p.Opcode = proto.SM_DETACH_STATUS

	status := uint8(1)
	if t.conn != nil {
		status = uint8(0)
	}
	p.WriteByte(byte(status))
	return proto.WritePacket(c, p)
}

func (t *Terminal) HandleIncoming(conn net.Conn) {
	for {
		p, err := proto.ReadPacket(conn)
		if err != nil {
			conn.Close()
			return
		}

		switch p.Opcode {
		case proto.CM_QUERY_DETACH_STATUS:
			t.sendDetachStatus(conn)

		case proto.CM_SCREEN_SIZE:

		case proto.CM_ATTACH_REQ:
			b, _ := p.ReadByte()
			detach := false
			if uint8(b) == 1 {
				detach = true
			}
			retp := &proto.Packet{}
			retp.Opcode = proto.SM_ATTACH_ACK

			if t.conn != nil {
				if detach {
					t.conn.Close()
				} else {
					retp.WriteByte(byte(0))
					retp.WriteString("already attached")
					proto.WritePacket(conn, retp)
					return
				}
			}
			retp.WriteByte(byte(1))

			proto.WritePacket(conn, retp)
			t.handleAttaching(conn)
		}
	}
}

func (t *Terminal) handleAttaching(conn net.Conn) {
	defer conn.Close()

	err := t.sendFirstScreenData(conn)
	if err != nil {
		return
	}
	t.conn = conn
	defer func() {
		t.conn = nil
	}()

	// r := bufio.NewReader(conn)
DONE:
	for {
		// b, err := r.ReadBytes('\n')
		p, err := proto.ReadPacket(conn)
		if err != nil {
			break DONE
		}

		switch p.Opcode {
		case proto.CM_USER_INPUT:
			b := p.Bytes()
			t.Input(b)
		}
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
