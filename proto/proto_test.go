package proto

import "testing"

func TestHeadData(t *testing.T) {
	p := &Packet{}
	p.WriteString("This is a test data packet!")

	t.Logf("HeadData: %d => %v\n", p.Size(), p.HeadData())
}

func TestPacket(t *testing.T) {
	p := &Packet{}
	p.WriteString("abcdef")

	data := Marshal(p)
	t.Logf("packet: %v", data)
	p2 := Unmarshal(data)
	t.Logf("Opcode: %d data:%s\n", p2.Opcode, p2.String())
}
