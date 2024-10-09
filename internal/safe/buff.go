package safe

import (
	"bytes"
	"context"
	"io"
	"sync"
)

// A Buffer is a variable-sized buffer of bytes with [Buffer.Read] and [Buffer.Write] methods.
// The zero value MUST NEVER be used--use [NewBuffer].
type Buffer struct {
	cond   *CondChan
	buf    *bytes.Buffer
	m      *sync.RWMutex
	closed bool
}

// Len returns the total number of bytes in the buffer.
func (b *Buffer) Len() int {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.buf.Len()
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with [ErrTooLarge].
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	defer b.cond.Broadcast()
	return b.buf.Write(p)
}

// ReadAt implements io.ReaderAt.
func (b *Buffer) ReadAt(p []byte, off int64) (n int, err error) {
	b.m.RLock()
	defer b.m.RUnlock()
	if off >= int64(b.buf.Len()) {
		if b.closed {
			return 0, io.EOF
		}
		return 0, nil
	}

	raw := b.buf.Bytes()
	n = copy(p, raw[off:])

	if b.closed && off >= int64(b.buf.Len()-n) {
		err = io.EOF
	}

	return
}

// Wait for Buffer to change.
func (b *Buffer) Wait(ctx context.Context, seq int64) (newSeq int64) {
	return b.cond.Wait(ctx, seq)
}

// Close implements io.Closer.
// Close should only be called if no more data will be written.
func (b *Buffer) Close() error {
	b.m.Lock()
	defer b.m.Unlock()
	defer b.cond.Broadcast()
	b.closed = true
	return nil
}

// NewBuffer creates and initializes a new [Buffer] using buf as its
// initial contents. The new [Buffer] takes ownership of buf, and the
// caller should not use buf after this call. NewBuffer is intended to
// prepare a [Buffer] to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new([Buffer]) (or just declaring a [Buffer] variable) is
// sufficient to initialize a [Buffer].
func NewBuffer(buf []byte) *Buffer {
	var b bytes.Buffer
	return &Buffer{cond: NewCond(), buf: &b, m: &sync.RWMutex{}}
}

// NewBufferString creates and initializes a new [Buffer] using string s as its
// initial contents. It is intended to prepare a buffer to read an existing
// string.
func NewBufferString(s string) *Buffer {
	return NewBuffer([]byte(s))
}
