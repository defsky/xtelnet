package main

import "bytes"

import "fmt"

type IACParseStatus int

type NVTOption interface {
	Parse(NVTCommand) IACParseStatus
	String() string
}

type NVTCommand interface {
	Parse() IACParseStatus
	String() string
}

type nvtCmd byte
type nvtOpt byte

const (
	IAC  nvtCmd = 255 // 0xFF	Interpret as Command
	DONT nvtCmd = 254 // 0xFE	Don't do something
	DO   nvtCmd = 253 // 0xFD	Do something
	WONT nvtCmd = 252 // 0xFC	Won't do something
	WILL nvtCmd = 251 // 0xFB	Will do something
	SB   nvtCmd = 250 // 0xFA	Subnegotiation Begin
	GA   nvtCmd = 249 // 0xF9	Go Ahead
	EL   nvtCmd = 248 // 0xF8	Erase Line
	EC   nvtCmd = 247 // 0xF7	Erase Character
	AYT  nvtCmd = 246 // 0xF6	Are You Here?
	NOP  nvtCmd = 241 // 0xF1	No operation
	SE   nvtCmd = 240 // 0xF0	Subnegotiation End
)

const (
	O_GMCP   nvtOpt = 201 // 0xC9	Generic MUD Communication Protocol
	O_ZMP    nvtOpt = 93  // 0x5D	Zenith MUD Protocol
	O_MXP    nvtOpt = 91  // 0x5B	MUD eXtension Protocol
	O_MSSP   nvtOpt = 70  // 0x46	MUD Server Status Protocol
	O_NENV   nvtOpt = 39  // 0x27	[RFC1572] New Environment
	O_NAWS   nvtOpt = 31  // 0x1F	[RFC1073] Negotiate About Window Size
	O_TTYPE  nvtOpt = 24  // 0x18	[RFC1091] Terminal Type
	O_ECHO   nvtOpt = 1   // 0x01	[RFC857]  Echo
	O_BINARY nvtOpt = 0   // 0x00	[RFC856]  Binary Transmission
)

const (
	O_TTYPE_REQ nvtOpt = 1
	O_TTYPE_ACK nvtOpt = 0
)

func (c nvtCmd) String() string {
	cmdName := map[NVTCommand]string{
		IAC:  "IAC",
		WILL: "WILL",
		WONT: "WONT",
		DO:   "DO",
		DONT: "DONT",
		SB:   "SB",
		SE:   "SE",
		GA:   "GA",
		EL:   "EL",
		EC:   "EC",
		AYT:  "AYT",
		NOP:  "NOP",
	}

	name, ok := cmdName[c]
	if ok {
		return name
	}

	return string(c)
}

func (c nvtCmd) Parse() IACParseStatus {
	switch c {
	case WILL, WONT, DO, DONT, SB:
		return WANT_OPT

	}

	return WANT_NOTHING
}

func (o nvtOpt) String() string {
	optName := map[NVTOption]string{
		O_TTYPE: "TTYPE",
		O_NAWS:  "NAWS",
		O_NENV:  "NENV",
		O_MXP:   "MXP",
		O_MSSP:  "MSSP",
		O_ZMP:   "ZMP",
		O_GMCP:  "GMCP",
		O_ECHO:  "ECHO",
	}

	name, ok := optName[o]
	if ok {
		return name
	}

	return string(o)
}
func (o nvtOpt) Parse(c NVTCommand) IACParseStatus {
	switch o {
	case O_ECHO:
		return WANT_NOTHING
	case O_TTYPE:
		switch c {
		case WILL, WONT, DO, DONT:
			return WANT_NOTHING
		case SB:
			return WANT_SUBOPT
		}
	}
	return WANT_NOTHING
}

const (
	WANT_CMD IACParseStatus = iota
	WANT_OPT
	WANT_SUBOPT
	WANT_DATA
	WANT_NOTHING
)

type IACMessage struct {
	status IACParseStatus
	cmd    NVTCommand
	opt    NVTOption
	subopt NVTOption
	data   bytes.Buffer
}

func (c *IACMessage) Scan(b byte) bool {
	switch c.status {
	case WANT_CMD:
		c.cmd = nvtCmd(b)
		c.status = c.cmd.Parse()
		return c.status != WANT_NOTHING
	case WANT_OPT:
		c.opt = nvtOpt(b)
		c.status = c.opt.Parse(c.cmd)
		return c.status != WANT_NOTHING
	case WANT_SUBOPT:
		c.subopt = nvtOpt(b)
		c.status = WANT_DATA
		return c.status != WANT_NOTHING
	case WANT_DATA:
		if SB == c.cmd && SE == nvtCmd(b) {
			return false
		}
		c.data.WriteByte(b)
	}

	return true
}

func (c *IACMessage) String() string {
	s := ""
	if c.cmd == nil {
		return s
	}
	s = "IAC " + c.cmd.String()

	if c.opt == nil {
		return s
	}
	s = s + " " + c.opt.String()

	if c.subopt == nil {
		return s
	}
	s = s + " " + c.subopt.String()

	if c.data.Len() <= 0 {
		return s
	}
	s = s + " " + fmt.Sprintf("%v", c.data)

	return s
}
