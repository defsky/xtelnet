package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

const (
	headSize   = 4
	opcodeSize = 2
)

var EInvalidPacket = errors.New("invalid data packet format")

// Packet wraps business data, byte order is Big-Endian.
//
// Packet struct:
// +-----------------------------------------------------------------------+
// | 0 byte | 1 byte | 2 byte | 3 byte | 4 byte |  5 byte |     .....      |
// +-----------------------------------+-----------------------------------+
// |            Packet Head            | Packet Data                       |
// +-----------------------------------+-----------------------------------+
// |    Data Length (4 bytes)          | Opcode (2 bytes) | Other Data     |
// +-----------------------------------------------------------------------+
type Packet struct {
	bytes.Buffer
	Opcode uint16
}

// Size return data size in packet
func (p *Packet) Size() int {
	return 2 + p.Len()
}

func makeHeadData(p *Packet) []byte {
	len := make([]byte, headSize)
	binary.BigEndian.PutUint32(len, uint32(p.Size()))

	return len
}

func Marshal(p *Packet) []byte {
	b := make([]byte, opcodeSize)
	binary.BigEndian.PutUint16(b, p.Opcode)

	return append(b, p.Bytes()...)
}

func Unmarshal(data []byte) (*Packet, error) {
	n := len(data)
	if n < opcodeSize {
		return nil, EInvalidPacket
	}
	p := &Packet{}

	cmd := data[0:opcodeSize]
	p.Opcode = binary.BigEndian.Uint16(cmd)

	if n > opcodeSize {
		p.Write(data[opcodeSize:])
	}

	return p, nil
}

func WritePacket(c net.Conn, p *Packet) error {
	b := makeHeadData(p)

	data := Marshal(p)
	b = append(b, data...)

	_, err := c.Write(b)

	return err
}

func ReadPacket(c net.Conn) (*Packet, error) {
	head := make([]byte, 4)
	_, err := c.Read(head)
	if err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(head)
	if dataLen > 0 && dataLen < 100*1024*1024 {
		data := make([]byte, dataLen)

		_, err = c.Read(data)
		if err != nil {
			return nil, err
		}

		return Unmarshal(data)
	}
	return nil, EInvalidPacket
}
