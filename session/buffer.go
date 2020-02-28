package session

import (
	"sync"
)

type OutBuffer struct {
	mu     sync.Mutex
	maxLen int
	buffer [][]byte
}

func NewBuffer(len int) *OutBuffer {
	return &OutBuffer{
		maxLen: len,
		buffer: make([][]byte, 0, 100),
	}
}

func (b *OutBuffer) Put(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, data)
	if len(b.buffer) > b.maxLen {
		b.buffer = b.buffer[1:]
	}
}

func (b *OutBuffer) Get(n int) [][]byte {
	if n <= 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	size := len(b.buffer)

	startIdx := 0
	// retSize := size

	if size > n {
		startIdx = size - n
		// retSize = n
	}

	ret := make([][]byte, 0)
	ret = append(ret, b.buffer[startIdx:]...)
	// src := b.buffer[startIdx:]
	// copy(ret, src)

	return ret
}
