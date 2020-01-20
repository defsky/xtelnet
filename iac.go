package main

import (
	"bytes"
	"fmt"
)

type IACParseStatus int

type NVTOption interface {
	String() string
	Byte() byte
}

type NVTCommand interface {
	String() string
	Byte() byte
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

func (c nvtCmd) Byte() byte {
	return byte(c)
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
func (o nvtOpt) Byte() byte {
	return byte(o)
}

const (
	WANT_CMD IACParseStatus = iota
	WANT_OPT
	WANT_SUBOPT
	WANT_DATA
	WANT_NOTHING
)

type IACPacket struct {
	data   bytes.Buffer
	cmd    NVTCommand
	opt    NVTOption
	status IACParseStatus
}

func (c *IACPacket) Bytes() []byte {
	b := make([]byte, 0)
	if c.cmd == nil {
		return nil
	}
	b = append(b, c.cmd.Byte())

	if c.opt == nil {
		return b
	}
	b = append(b, c.opt.Byte())

	if c.data.Len() <= 0 {
		return b
	}
	b = append(b, c.data.Bytes()...)

	return b
}

// Scan will put b in packet, return false indicate not need any more byte
func (c *IACPacket) Scan(b byte) bool {
	// drop IAC
	if b == byte(IAC) {
		return true
	}
	switch c.status {
	case WANT_CMD:
		//c.WriteByte(b)
		c.cmd = nvtCmd(b)
		switch c.cmd {
		case WILL, WONT, DO, DONT, SB:
			c.status = WANT_OPT
		default:
			c.status = WANT_NOTHING
		}
		return c.status != WANT_NOTHING
	case WANT_OPT:
		//c.WriteByte(b)
		c.opt = nvtOpt(b)
		switch c.cmd {
		case SB:
			c.status = WANT_DATA
		default:
			c.status = WANT_NOTHING
		}
		return c.status != WANT_NOTHING
	case WANT_DATA:
		if SE == nvtCmd(b) {
			return false
		}
		c.data.WriteByte(b)
	}

	return true
}

func (c *IACPacket) String() string {
	s := "IAC "
	if c.cmd == nil {
		return ""
	}
	s = s + c.cmd.String()

	if c.opt == nil {
		return s
	}
	s = s + " " + c.opt.String()

	if c.data.Len() <= 0 {
		return s
	}
	s = s + " " + fmt.Sprintf("%v", c.data.Bytes())

	return s
}

type NVTOptionConfig struct {
	options   map[NVTOption]bool
	serverOpt map[NVTOption]bool
}

func NewNVTOptionConfig() *NVTOptionConfig {
	cfg := &NVTOptionConfig{
		options: map[NVTOption]bool{
			O_ECHO:  true,
			O_TTYPE: true,
		},
		serverOpt: map[NVTOption]bool{},
	}

	return cfg
}

func (c *NVTOptionConfig) Get(o NVTOption) bool {
	return c.options[o]
}
func (c *NVTOptionConfig) Set(o NVTOption, v bool) {
	c.options[o] = v
}

func (c *NVTOptionConfig) GetRemote(o NVTOption) bool {
	return c.serverOpt[o]
}

type NVTCommandHandler func(cfg *NVTOptionConfig, data *IACPacket) *IACPacket
type NVTCommandHandlerMap map[NVTCommand]NVTCommandHandler
type IACReactor struct {
	config  *NVTOptionConfig
	handler NVTCommandHandlerMap
}

func NewIACReactor(cfg *NVTOptionConfig) *IACReactor {
	return &IACReactor{
		config: cfg,
		handler: NVTCommandHandlerMap{
			WILL: handleNVTWill,
			WONT: handleNVTWont,
			DO:   handleNVTDo,
			DONT: handleNVTDont,
			SB:   handleNVTSb,
		},
	}
}

func (r *IACReactor) React(m *IACPacket) *IACPacket {
	handler, ok := r.handler[m.cmd]
	if !ok {
		return nil
	}
	return handler(r.config, m)
}

func handleNVTWill(cfg *NVTOptionConfig, p *IACPacket) *IACPacket {
	if cfg.Get(p.opt) {
		p.cmd = DO
		cfg.serverOpt[p.opt] = true
	} else {
		p.cmd = DONT
		cfg.serverOpt[p.opt] = false
	}
	return p
}
func handleNVTWont(cfg *NVTOptionConfig, p *IACPacket) *IACPacket {
	p.cmd = DONT
	cfg.serverOpt[p.opt] = false
	return p
}
func handleNVTDo(cfg *NVTOptionConfig, p *IACPacket) *IACPacket {
	if cfg.Get(p.opt) {
		p.cmd = WILL
	} else {
		p.cmd = DONT
	}
	return p
}
func handleNVTDont(cfg *NVTOptionConfig, p *IACPacket) *IACPacket {
	p.cmd = WONT
	return p
}
func handleNVTSb(cfg *NVTOptionConfig, p *IACPacket) *IACPacket {
	switch p.opt {
	case O_TTYPE:
		subopt, err := p.data.ReadByte()
		if err != nil || subopt != 1 {
			return nil
		}
		p.data.Reset()

		buf := []byte{0}
		buf = append(buf, []byte("xtelnet")...)
		buf = append(buf, []byte{byte(IAC), byte(SE)}...)

		p.data.Write(buf)

		return p
	}

	return nil
}
