package bidirpc

import "bytes"

type bufferPool struct {
	bufC chan *bytes.Buffer
}

func newBufferPool(size int) (bp *bufferPool) {
	return &bufferPool{
		bufC: make(chan *bytes.Buffer, size),
	}
}

func (bp *bufferPool) Get() (b *bytes.Buffer) {
	select {
	case b = <-bp.bufC:
		// Reuse existing buffer
	default:
		// Create new buffer
		b = bytes.NewBuffer([]byte{})
	}
	return
}

func (bp *bufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	select {
	case bp.bufC <- b:
		// OK
	default:
		// Discard the buffer
	}
}
