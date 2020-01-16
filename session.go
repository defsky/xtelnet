package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/rivo/tview"
)

type TaskType int

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

// SessionOption contains some options of session
type SessionOption struct {
	DebugColor     bool
	DebugAnsiColor bool
}

// Session is a telnet session based on net.Conn
type Session struct {
	wg      sync.WaitGroup
	Option  *SessionOption
	host    string
	out     io.Writer
	conn    net.Conn
	closing bool

	inBuffer    chan byte
	iacInBuffer chan *IACMessage
	outBuffer   chan []byte

	closeTimer chan struct{}
}

// NewSession will return a new session with host and output message to out
func NewSession(host string, out io.Writer) (*Session, error) {
	conn, err := net.DialTimeout("tcp", host, 30*time.Second)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		Option:      &SessionOption{},
		host:        host,
		out:         out,
		conn:        conn,
		inBuffer:    make(chan byte, 4096),
		iacInBuffer: make(chan *IACMessage, 20),
		outBuffer:   make(chan []byte, 80),
		closeTimer:  make(chan struct{}),
	}

	sess.RunEvery(time.Minute, func() {
		sess.Send([]byte("look\r\n"))
	})

	sess.wg.Add(1)
	go sess.receiver()

	return sess, nil
}

// IsAlive
func (s *Session) IsAlive() bool {
	return !s.closing
}

// Close will close session
func (s *Session) Close() {
	if s.closing {
		return
	}
	s.closing = true
	close(s.closeTimer)
	close(s.outBuffer)
}

// Send wil send data to session
func (s *Session) Send(data []byte) bool {
	if s.closing {
		return false
	}
	fmt.Fprint(s.out, string(data))
	s.outBuffer <- data

	return true
}

// RunAfter wil call f only once when d duration elapsed
func (s *Session) RunAfter(d time.Duration, f func()) {
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
func (s *Session) RunEvery(d time.Duration, f func()) {
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

func (s *Session) preprocessor() {
	defer func() {
		s.Close()
		s.wg.Done()
		s.wg.Wait()
		fmt.Fprintln(s.out, "Session closed")
	}()

	var w io.Writer
	ansiWriter := tview.ANSIWriter(s.out)
	buffer := new(bytes.Buffer)

	s.wg.Add(1)
	go s.sender()

DONE:
	for {
		select {
		case b, ok := <-s.inBuffer:
			if !ok {
				break DONE
			}
			buffer.WriteByte(b)
		case _, ok := <-s.iacInBuffer:
			if !ok {
				break DONE
			}

		default:
			if buffer.Len() > 0 {
				msg := DecodeFrom("GB18030", buffer.Bytes())
				r, _ := utf8.DecodeLastRune(msg)
				if r != utf8.RuneError {
					buffer.Reset()

					if s.Option.DebugAnsiColor {
						w = s.out
					} else {
						w = ansiWriter
					}
					fmt.Fprint(w, string(msg))
					break
				}
			}
			b2, ok := <-s.inBuffer
			if !ok {
				break DONE
			}
			buffer.WriteByte(b2)
		}
	}

	if buffer.Len() > 0 {
		msg := DecodeFrom("GB18030", buffer.Bytes())
		if s.Option.DebugAnsiColor {
			w = s.out
		} else {
			w = ansiWriter
		}
		fmt.Fprint(w, string(msg))
	}
}

func (s *Session) iacprocessor() {

}
func (s *Session) receiver() {
	defer func() {
		close(s.inBuffer)
		s.wg.Done()
	}()

	buf := bufio.NewReaderSize(s.conn, 2048)

	var b byte
	var err error

	s.wg.Add(1)
	go s.preprocessor()

DONE:
	for {
		b, err = buf.ReadByte()
		if err != nil {
			break DONE
		}

		// IAC
		if b == byte(IAC) {
			iac, e := readIACMessage(buf)
			if e != nil {
				break DONE
			}
			s.iacInBuffer <- iac
			writeBytes(s.inBuffer, []byte(iac.String()+"\n"))
			continue
		}

		// ansi escape sequence
		if b == byte(0x1b) {
			data, e := readEscSeq(buf)
			if e != nil {
				break DONE
			}
			s.inBuffer <- b
			writeBytes(s.inBuffer, data)

			continue
		}

		s.inBuffer <- b
	}

	writeBytes(s.inBuffer, handleConnError(err))
}

func (s *Session) sender() {
	defer s.wg.Done()

	writer := bufio.NewWriter(s.conn)

DONE:
	for {
		data, ok := <-s.outBuffer
		if !ok {
			s.conn.Close()
			break DONE
		}
		data = EncodeTo("GB18030", data)

		_, err := writer.Write(data)
		if err != nil {
			writeBytes(s.inBuffer, []byte(err.Error()+"\n"))
			continue
		}

		err = writer.Flush()
		if err != nil {
			writeBytes(s.inBuffer, []byte(err.Error()+"\n"))
		}
	}
}

// writeBytes will write p into channel out byte by byte
func writeBytes(out chan<- byte, p []byte) {
	for _, v := range p {
		out <- v
	}
}

func handleConnError(err error) []byte {
	switch err {
	case io.EOF:
		return []byte("\nconnection was closed by server ...\n")
	default:
		return []byte(fmt.Sprintf("\n%s\n", err.Error()))
	}
}

// readEscSeq will read a complete ansi escape sequence from inbuffer
func readEscSeq(r *bufio.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)

	var b byte
	var err error
DONE:
	for b, err = r.ReadByte(); err == nil; b, err = r.ReadByte() {
		buf.WriteByte(b)

		if b == byte('m') || b == byte('\n') {
			break DONE
		}
	}

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// readIAC will read a complete NVT command from inbuffer
func readIACMessage(r *bufio.Reader) (*IACMessage, error) {
	iac := &IACMessage{}
	var err error
	var b byte
	for {
		b, err = r.ReadByte()
		if err != nil {
			break
		}
		if false == iac.Scan(b) {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return iac, nil
}

// readIAC will read a complete NVT command from inbuffer
func readIAC(r *bufio.Reader) ([]byte, error) {
	iac := new(bytes.Buffer)
	isSubCmd := false

	var b byte
	var err error
DONE:
	for b, err = r.ReadByte(); err == nil; b, err = r.ReadByte() {
		// drop IAC byte when readIAC
		if b == byte(IAC) {
			continue
		}

		iac.WriteByte(b)

		switch b {
		case byte(WILL), byte(WONT), byte(DO), byte(DONT):

			// read next byte
		case byte(SB): // IAC SB
			isSubCmd = true

		default:
			if isSubCmd {
				// IAC SE
				if b == byte(SE) {
					break DONE
				}

				// read next byte
				break
			}

			break DONE
		}
	}

	if err != nil {
		return nil, err
	} else {
		return iac.Bytes(), nil
	}
}
