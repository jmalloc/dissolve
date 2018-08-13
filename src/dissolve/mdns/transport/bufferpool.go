package transport

import (
	"sync"
)

const bufferSize = 65536

var buffers = sync.Pool{
	New: func() interface{} {
		return make([]byte, bufferSize)
	},
}

// getBuffer fetches a buffer from the buffer pool.
func getBuffer() []byte {
	return buffers.Get().([]byte)
}

// putBuffer returns a buffer to the buffer pool.
func putBuffer(buf []byte) {
	if cap(buf) >= bufferSize {
		buffers.Put(buf[:bufferSize])
	}
}
