package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/defsky/telnet"
	"github.com/rivo/tview"
)

type Session struct {
	nvt    telnet.NVT
	conn   net.Conn
	wg     sync.WaitGroup
	netInQ chan byte
	outQ   chan []byte
	inQ    chan []byte
	msgQ   chan int
	done   chan struct{}
	cache  bytes.Buffer
	out    io.Writer
	host   string
	Option *SessionOption
}

type SessionOption struct {
	DebugColor bool
}

func NewSession(host string, out io.Writer) (*Session, error) {
	conn, err := net.DialTimeout("tcp", host, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return &Session{
		conn:   conn,
		netInQ: make(chan byte, 4096),
		outQ:   make(chan []byte, 80),
		inQ:    make(chan []byte, 80),
		done:   make(chan struct{}),
		msgQ:   make(chan int, 1),
		out:    out,
		host:   host,
		Option: &SessionOption{},
	}, nil
}
func (s *Session) Start() {
	s.wg.Add(3)
	go receiver(s.conn, s.netInQ, s.done, s.msgQ, &s.wg)
	go sender(s.conn, s.outQ, s.done, s.msgQ, &s.wg)
	go messageProcessor(s.netInQ, s.inQ, s.done, &s.wg)

	writer := tview.ANSIWriter(s.out)
DONE:
	for {
		select {
		case <-s.done:
			break DONE
		case <-s.msgQ:
			close(s.done)
			break DONE
		case line := <-s.inQ:
			s.cache.Write(line)
		default:
			if s.cache.Len() > 0 {
				//msg := tview.TranslateANSI(s.cache.String())
				fmt.Fprint(writer, s.cache.String())
				s.cache.Reset()
			}
		}
	}
	s.wg.Wait()
	fmt.Fprintf(s.out, "\nsession to %s closed ...\n", s.host)
	UserShell.SetSession(nil)
}
func (s *Session) Close() {
	s.conn.Close()
	//close(s.done)
}
func (s *Session) Send(data []byte) {
	s.outQ <- data
}

func messageProcessor(in <-chan byte, out chan<- []byte, done <-chan struct{}, wg *sync.WaitGroup) {
	var buffer bytes.Buffer
	var b bytes.Buffer
	buf := bufio.NewReadWriter(bufio.NewReader(&buffer), bufio.NewWriter(&buffer))
DONE:
	for {
		select {
		case <-done:
			break DONE
		case b := <-in:
			buffer.WriteByte(b)
			if b == byte(0x1b) {
			ESCAPE_END:
				for {
					select {
					case eb := <-in:
						buffer.WriteByte(eb)
						if eb == byte('m') {
							break ESCAPE_END
						}
					default:
						time.Sleep(100 * time.Millisecond)
					}
				}
			}
		default:
			line, err := buf.ReadBytes('\n')
			if err != nil && err != io.EOF {
				break
			}
			msg := DecodeFrom("GB18030", append(b.Bytes(), line...))
			if msg == nil {
				b.Write(line)
			} else {
				out <- msg
				b.Reset()
			}
		}
	}
	wg.Done()
}
func sender(c net.Conn, out <-chan []byte, done <-chan struct{}, msg chan<- int, wg *sync.WaitGroup) {
	buf := bufio.NewWriter(c)
DONE:
	for {
		select {
		case <-done:
			break DONE
		case data := <-out:
			data = EncodeTo("GB18030", data)
			_, err := buf.Write(data)
			if err != nil || err == io.EOF {
				msg <- 1
				break DONE
			}
			buf.Flush()
		}
	}

	wg.Done()
}
func receiver(c net.Conn, in chan<- byte, done <-chan struct{}, msg chan<- int, wg *sync.WaitGroup) {
	buf := bufio.NewReaderSize(c, 2048)

DONE:
	for {
		select {
		case <-done:
			break DONE
		default:
			b, err := buf.ReadByte()
			if err != nil {
				if err == io.EOF {

				}
				msg <- 1
				break DONE
			}
			if b == byte(255) {
				// TODO: process IAC sequence
				data, err := parseIAC(buf)
				if err != nil {
					msg <- 1
					break DONE
				}
				fmt.Fprintf(screen, "IAC %v\n", data)
			} else {
				in <- b
			}
		}
	}

	wg.Done()
}

func parseIAC(buf *bufio.Reader) ([]byte, error) {
	ret := make([]byte, 0)
	var b byte
	var err error
	for b, err = buf.ReadByte(); err == nil; {
		ret = append(ret, b)
		switch uint8(b) {
		case 254, 253, 252, 251:

		default:
			return ret, nil
		}

		b, err = buf.ReadByte()
	}

	return nil, err
}
