package proto

import (
	"bytes"
	"encoding/binary"
	"net"
)

const (
	headSize   = 4
	opcodeSize = 2
)

type Packet struct {
	Opcode uint16
	bytes.Buffer
}

func (p *Packet) Size() int {
	return 2 + p.Len()
}

func (p *Packet) HeadData() []byte {
	len := make([]byte, headSize)
	binary.BigEndian.PutUint32(len, uint32(p.Size()))

	return len
}

func Marshal(p *Packet) []byte {
	b := make([]byte, opcodeSize)
	binary.BigEndian.PutUint16(b, p.Opcode)

	return append(b, p.Bytes()...)
}

func Unmarshal(data []byte) *Packet {
	p := &Packet{}

	cmd := data[0:opcodeSize]
	p.Opcode = binary.BigEndian.Uint16(cmd)

	p.Write(data[opcodeSize:])

	return p
}

func Send(c *net.UnixConn, p *Packet) error {
	b := p.HeadData()

	data := Marshal(p)
	b = append(b, data...)

	_, err := c.Write(b)

	return err
}

func ReadPacket(c *net.UnixConn) (*Packet, error) {
	head := make([]byte, 4)
	_, err := c.Read(head)
	if err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(head)
	data := make([]byte, dataLen)

	_, err = c.Read(data)
	if err != nil {
		return nil, err
	}

	return Unmarshal(data), nil
}
