package proto

import (
	"bytes"
	"encoding/binary"
)

type Packet struct {
	cmd  uint16
	data bytes.Buffer
}

func (p *Packet) Size() int {
	return 2 + p.data.Len()
}

func (p *Packet) HeadData() []byte {
	len := make([]byte, 4)
	binary.BigEndian.PutUint32(len, uint32(p.Size()))

	return len
}

func Marshal(p *Packet) []byte {
	cmd := make([]byte, 2)
	binary.BigEndian.PutUint16(cmd, p.cmd)

	return append(cmd, p.data.Bytes()...)
}

func Unmarshal(data []byte) *Packet {
	p := &Packet{}

	cmd := data[0:2]
	p.cmd = binary.BigEndian.Uint16(cmd)
	p.data.Write(data[2:])

	return p
}
