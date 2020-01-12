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

type SessionOption struct {
	DebugColor     bool
	DebugAnsiColor bool
}

type Session struct {
	wg     sync.WaitGroup
	cache  bytes.Buffer
	conn   net.Conn
	outBuf chan []byte
	out    io.Writer
	host   string
	Option *SessionOption
}

func NewSession(host string, out io.Writer) (*Session, error) {
	conn, err := net.DialTimeout("tcp", host, 30*time.Second)
	if err != nil {
		return nil, err
	}

	recvBuf := make(chan byte, 4096)
	inQ := make(chan []byte, 80)
	outQ := make(chan []byte, 80)

	sess := &Session{
		conn:   conn,
		outBuf: outQ,
		out:    out,
		host:   host,
		Option: &SessionOption{},
	}

	sess.wg.Add(3)
	go sess.preprocessor(recvBuf, inQ)
	go sess.receiver(recvBuf)
	go sess.sender(outQ)

	go func(s *Session) {
		writer := tview.ANSIWriter(s.out)
	DONE:
		for {
			select {
			case line, ok := <-inQ:
				if !ok {
					break DONE
				}
				s.cache.Write(line)
			default:
				if s.cache.Len() > 0 {
					var w io.Writer
					if s.Option.DebugAnsiColor {
						w = s.out
					} else {
						w = writer
					}
					fmt.Fprint(w, s.cache.String())
					s.cache.Reset()
				} else {
					time.Sleep(300 * time.Millisecond)
				}
			}
		}

		s.wg.Wait()
		fmt.Fprintf(s.out, "\nsession to %s closed ...\n", s.host)
		UserShell.SetSession(nil)
	}(sess)

	return sess, nil
}

func (s *Session) Close() {
	close(s.outBuf)
}
func (s *Session) Send(data []byte) {
	s.outBuf <- data
}

func (s *Session) sender(in <-chan []byte) {
	writer := bufio.NewWriter(s.conn)

DONE:
	for data := range in {
		if data == nil {
			break DONE
		}
		data = EncodeTo("GB18030", data)
		_, err := writer.Write(data)
		if err != nil {
			break DONE
		}
		writer.Flush()
	}

	s.conn.Close()
	s.wg.Done()
}

func (s *Session) receiver(out chan<- byte) {
	buf := bufio.NewReaderSize(s.conn, 2048)

	var b byte
	var err error
DONE:
	for b, err = buf.ReadByte(); err == nil; b, err = buf.ReadByte() {
		if b == byte(255) {
			cmd, err := readIAC(buf)
			if err != nil {
				break DONE
			}
			writeBytes(out, []byte(fmt.Sprintf("IAC %v\n", cmd)))
			continue
		}
		out <- b
	}

	writeBytes(out, handleConnError(err))

	close(out)
	s.wg.Done()
}

func (s *Session) preprocessor(in <-chan byte, out chan<- []byte) {
	var b bytes.Buffer
	buffer := new(bytes.Buffer)
	buf := bufio.NewReader(buffer)
DONE:
	for {
		select {
		case b, ok := <-in:
			if !ok {
				break DONE
			}
			buffer.WriteByte(b)
			if b == byte(0x1b) {
			ESCAPE_END:
				for {
					select {
					case eb, ok := <-in:
						if !ok {
							break DONE
						}
						buffer.WriteByte(eb)
						if eb == byte('m') {
							break ESCAPE_END
						}
					default:
						time.Sleep(300 * time.Millisecond)
					}
				}
			}
		default:
			line, err := buf.ReadBytes('\n')
			if err != nil && err != io.EOF {
				break
			}
			if len(line) > 0 {
				msg := DecodeFrom("GB18030", append(b.Bytes(), line...))
				_, err := utf8.DecodeLastRune(msg)
				if err == utf8.RuneError {
					b.Write(line)
				} else {
					out <- msg
					b.Reset()
				}
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	close(out)
	s.wg.Done()
}

func writeBytes(out chan<- byte, p []byte) {
	for _, v := range p {
		out <- v
	}
}

func handleConnError(err error) []byte {
	switch err {
	case io.EOF:
		return []byte("connection was closed by server\n")
	default:
		return []byte(err.Error())
	}
}

func readIAC(reader *bufio.Reader) ([]byte, error) {
	iac := new(bytes.Buffer)
	isSubCmd := false

	var b byte
	var err error
DONE:
	for b, err = reader.ReadByte(); err == nil; b, err = reader.ReadByte() {
		// drop IAC byte
		if b == byte(255) {
			continue
		}

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
