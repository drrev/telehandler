package safe

import (
	"bytes"
	"io"
	"sync"
)

// A Buffer is a variable-sized Buffer of bytes with [Buffer.Read] and [Buffer.Write] methods.
// The zero value MUST NEVER be used--use [NewBuffer].
type Buffer struct {
	cond   *sync.Cond
	buf    *bytes.Buffer
	m      *sync.RWMutex
	closed bool
}

// Len returns the total number of bytes in the CondBuff.
func (b *Buffer) Len() int {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.buf.Len()
}

// Write appends the contents of p to the CondBuff, growing the CondBuff as
// needed. The return value n is the length of p; err is always nil. If the
// CondBuff becomes too large, Write will panic with [ErrTooLarge].
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	defer b.cond.Broadcast()
	return b.buf.Write(p)
}

// ReadAt implements io.ReaderAt.
func (b *Buffer) ReadAt(p []byte, off int64) (n int, err error) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	if off >= int64(b.buf.Len()) && b.closed {
		err = io.EOF
		return
	}

	for !b.closed && off >= int64(b.buf.Len()) {
		b.cond.Wait()
	}

	raw := b.buf.Bytes()
	n = copy(p, raw[off:])

	// ensure error is set if on short read
	if n < len(p) {
		if b.closed {
			err = io.EOF
		}
	}
	return
}

// Closed returns true if the CondBuff is closed.
func (b *Buffer) Closed() bool {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.closed
}

// Close implements io.Closer.
// Close should only be called if no more data will be written.
func (b *Buffer) Close() error {
	b.m.Lock()
	defer b.m.Unlock()
	b.closed = true
	b.cond.Broadcast()
	return nil
}

// Wake forces a broadcast to wake up all Readers.
func (b *Buffer) Wake() {
	b.cond.Broadcast()
}

// NewBuffer creates and initializes a new [Buffer] using buf as its
// initial contents. The new [Buffer] takes ownership of buf, and the
// caller should not use buf after this call. NewBuffer is intended to
// prepare a [Buffer] to read existing data. It can also be used to set
// the initial size of the internal CondBuff for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new([Buffer]) (or just declaring a [Buffer] variable) is
// sufficient to initialize a [Buffer].
func NewBuffer(buf []byte) *Buffer {
	m := &sync.RWMutex{}
	return &Buffer{cond: sync.NewCond(m.RLocker()), buf: &bytes.Buffer{}, m: m}
}
