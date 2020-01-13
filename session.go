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

type SessionOption struct {
	DebugColor     bool
	DebugAnsiColor bool
}

type Session struct {
	wg     sync.WaitGroup
	conn   net.Conn
	out    io.Writer
	host   string
	Option *SessionOption

	inBuffer  chan byte
	outBuffer chan []byte

	closeTimer chan struct{}
}

func NewSession(host string, out io.Writer) (*Session, error) {
	conn, err := net.DialTimeout("tcp", host, 30*time.Second)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		conn:       conn,
		out:        out,
		host:       host,
		Option:     &SessionOption{},
		inBuffer:   make(chan byte, 4096),
		outBuffer:  make(chan []byte, 80),
		closeTimer: make(chan struct{}),
	}

	sess.wg.Add(1)
	go sess.receiver()

	sess.RunEvery(time.Minute, func() {
		sess.Send([]byte("look\r\n"))
	})

	return sess, nil
}

func (s *Session) Close() {
	close(s.closeTimer)
	close(s.outBuffer)
}
func (s *Session) Send(data []byte) {
	s.outBuffer <- data
}
func (s *Session) RunAfter(d time.Duration, f func()) {

}
func (s *Session) RunEvery(d time.Duration, f func()) {
	ticker := time.NewTicker(d)
	go func(f func()) {
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
func (s *Session) sender() {
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
			break DONE
		}
		writer.Flush()
	}

	s.wg.Done()
}

func (s *Session) preprocessor() {
	defer s.wg.Done()

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
	for b, err = buf.ReadByte(); err == nil; b, err = buf.ReadByte() {
		// IAC
		if b == byte(0xff) {
			cmd, e := readIAC(buf)
			if e != nil {
				break DONE
			}
			writeBytes(s.inBuffer, []byte(fmt.Sprintf("IAC %v\r\n", cmd)))
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

		if b == byte('m') {
			break DONE
		}
	}

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// readIAC will read a complete NVT command from inbuffer
func readIAC(r *bufio.Reader) ([]byte, error) {
	iac := new(bytes.Buffer)
	isSubCmd := false

	var b byte
	var err error
DONE:
	for b, err = r.ReadByte(); err == nil; b, err = r.ReadByte() {
		// embeded IAC
		// if b == byte(255) {
		// 	cmd, e := readIAC(r)
		// 	if e != nil {
		// 		break DONE
		// 	}
		// 	iac.Write(cmd)

		// }

		iac.WriteByte(b)

		switch uint8(b) {
		case 254, // DONT
			253, // DO
			252, // WONT
			251: // WILL

			// read next byte
		case 250: // IAC SB
			isSubCmd = true
		default:
			if isSubCmd {
				// IAC SE
				if b == byte(240) {
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
