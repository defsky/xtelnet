package proto

import "testing"

func TestPacket(t *testing.T) {
	p := &Packet{}
	p.WriteString("abcdef")

	data := Marshal(p)
	t.Logf("packet: %v", data)
	p2, err := Unmarshal(data)
	if err != nil {
		t.Log(err)
	}
	t.Logf("Opcode: %d data:%s\n", p2.Opcode, p2.String())
}
