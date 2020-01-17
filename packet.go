package main

type DataPacket interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}
