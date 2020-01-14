package main

type NVTCommand byte
type NVTOptionCode byte

const (
	IAC  NVTCommand = 255 // 0xFF	Interpret as Command
	DONT NVTCommand = 254 // 0xFE	Don't do something
	DO   NVTCommand = 253 // 0xFD	Do something
	WONT NVTCommand = 252 // 0xFC	Won't do something
	WILL NVTCommand = 251 // 0xFB	Will do something
	SB   NVTCommand = 250 // 0xFA	Subnegotiation Begin
	GA   NVTCommand = 249 // 0xF9	Go Ahead
	EL   NVTCommand = 248 // 0xF8	Erase Line
	EC   NVTCommand = 247 // 0xF7	Erase Character
	AYT  NVTCommand = 246 // 0xF6	Are You Here?
	NOP  NVTCommand = 241 // 0xF1	No operation
	SE   NVTCommand = 240 // 0xF0	Subnegotiation End
)

const (
	GMCP  NVTOptionCode = 201 // 0xC9	Generic MUD Communication Protocol
	ZMP   NVTOptionCode = 93  // 0x5D	Zenith MUD Protocol
	MXP   NVTOptionCode = 91  // 0x5B	MUD eXtension Protocol
	MSSP  NVTOptionCode = 70  // 0x46	MUD Server Status Protocol
	NENV  NVTOptionCode = 39  // 0x27	New Environment
	NAWS  NVTOptionCode = 31  // 0x1F	Negotiate About Window Size
	TTYPE NVTOptionCode = 24  // 0x18	Terminal Type
	ECHO  NVTOptionCode = 1   // 0x01	Echo
)

func (c NVTCommand) String() string {
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

	return cmdName[c]
}

func (c NVTOptionCode) String() string {
	optName := map[NVTOptionCode]string{
		TTYPE: "TTYPE",
		NAWS:  "NAWS",
		NENV:  "NENV",
		MXP:   "MXP",
		MSSP:  "MSSP",
		ZMP:   "ZMP",
		GMCP:  "GMCP",
		ECHO:  "ECHO",
	}

	return optName[c]
}
