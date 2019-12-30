package main

import "golang.org/x/text/encoding/simplifiedchinese"

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

func DecodeFrom(charset Charset, data []byte) string {

	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(data)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(data)
	}

	return str
}

func EncodeTo(charset Charset, s string) []byte {
	var str string

	switch charset {
	case GB18030:
		str, _ = simplifiedchinese.GB18030.NewEncoder().String(s)
	case UTF8:
		fallthrough
	default:
		str = ""
	}

	return []byte(str)
}
