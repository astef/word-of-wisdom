package main

import (
	"bytes"
	"sync"
)

type bufferPool struct {
	readBufferPool  *sync.Pool
	writeBufferPool *sync.Pool
}

type buffers struct {
	pool        *bufferPool
	ReadBuffer  []byte
	WriteBuffer *bytes.Buffer
}

func NewBufferPool(readBufferSize int) *bufferPool {
	return &bufferPool{
		readBufferPool: &sync.Pool{
			New: func() any {
				buffer := make([]byte, readBufferSize)
				return &buffer
			},
		},
		writeBufferPool: &sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

func (bp *bufferPool) Borrow() buffers {
	return buffers{
		ReadBuffer:  *bp.readBufferPool.Get().(*[]byte),
		WriteBuffer: bp.writeBufferPool.Get().(*bytes.Buffer),
		pool:        bp,
	}
}

func (b buffers) Release() {
	// clear memory
	clear(b.ReadBuffer)
	b.WriteBuffer.Reset()

	// return back to pool
	b.pool.readBufferPool.Put(&b.ReadBuffer)
	b.pool.writeBufferPool.Put(b.WriteBuffer)
}
