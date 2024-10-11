package safe

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
)

// ErrTooEarly is a retryable error when the underlying buffer has no more data, but
// more data may be available in the future.
var ErrTooEarly = fmt.Errorf("read occurred too early, data is not ready yet")

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

// CopyAt reads len(p) bytes into p starting at offset off in the underlying input source.
// It returns the number of bytes read (0 <= n <= len(p)) and any error encountered.
//
// When CopyAt returns n < len(p), it returns a non-nil error explaining why more bytes were not returned.
// [ErrTooEarly] is returned if reads should be retried.
// [io.EOF] is returned if the reader will not receive any more data.
//
// Even if CopyAt returns n < len(p), it may use all of p as scratch space during the call.
//
// Clients of CopyAt can execute parallel CopyAt calls on the same input source.
func (b *Buffer) CopyAt(p []byte, off int64) (n int, err error) {
	b.m.RLock()
	defer b.m.RUnlock()
	if off >= int64(b.buf.Len()) {
		if b.closed {
			err = io.EOF
			return
		}
		err = ErrTooEarly
		return
	}

	raw := b.buf.Bytes()
	n = copy(p, raw[off:])

	// ensure error is set if on short read
	if n < len(p) {
		if b.closed {
			err = io.EOF
		} else {
			err = ErrTooEarly
		}
	}

	return
}

// Wait for Buffer to change.
func (b *Buffer) Wait(ctx context.Context, seq int64) (newSeq int64) {
	// If this buffer is closed, do not allow waiting
	if b.Closed() {
		return seq
	}

	return b.cond.Wait(ctx, seq)
}

// Closed returns true if the Buffer is closed.
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
