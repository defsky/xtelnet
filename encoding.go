package main

import "golang.org/x/text/encoding/simplifiedchinese"

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

func DecodeFrom(charset Charset, data []byte) []byte {

	var str []byte
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(data)
		str = decodeBytes
	case UTF8:
		fallthrough
	default:
		str = data
	}

	return str
}

func EncodeTo(charset Charset, s []byte) []byte {
	var str []byte

	switch charset {
	case GB18030:
		str, _ = simplifiedchinese.GB18030.NewEncoder().Bytes(s)
	case UTF8:
		fallthrough
	default:

	}

	return str
}
