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
	return len(p.cmd) + p.data.Len()
}

func (p *Packet) HeadData() []byte {
	len := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(p.Size()))

	return len
}

func Marshal(p *Packet) []byte {
	b := bytes.NewBuffer()

	cmd := make([]byte, 2)
	binary.BigEndian.PutUint16(cmd, p.cmd)
	b.Write(cmd)

	b.Write(p.data.Bytes())

	return b.Bytes()
}

func Unmarshal(data []byte) *Packet {
	p := &Packet{}

	cmd := data[0:2]
	p.cmd = binary.BigEndian.Uint16(cmd)
	p.data.Write(data[2:])

	return p
}
