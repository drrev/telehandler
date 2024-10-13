package safe

import (
	"errors"
	"io"
	"sync"
)

// ErrClosedWriter is returned if [NotifyingBuffer.Write] is called
// after [NotifyingBuffer] is closed.
var ErrClosedWriter = errors.New("io: write on closed writer")

// NotifyingBuffer is a utility type that implements a thread-safe
// API. This type is useful for writing asynchronously into a buffer
// that grows automatically. Do note, everything is buffered in memory.
//
// The buffer can be read by any number of readers by requesting a reader instance
// with [NotifyingBuffer.Reader].
//
// See [NewNotifyingBuffer].
type NotifyingBuffer struct {
	closed bool
	mu     sync.RWMutex
	buff   []byte
	notify chan struct{}
}

// NewNotifyingBuffer creates a [NotifyingBuffer] that is initialized
// and ready to use.
func NewNotifyingBuffer() *NotifyingBuffer {
	return &NotifyingBuffer{
		closed: false,
		mu:     sync.RWMutex{},
		buff:   []byte{},
		notify: make(chan struct{}),
	}
}

// Reader returns a new [NotifyingBufferReader] to read this buffer.
func (b *NotifyingBuffer) Reader() *NotifyingBufferReader {
	return &NotifyingBufferReader{
		offs:  0,
		nb:    b,
		close: make(chan struct{}),
	}
}

// Write implements io.Writer.
func (b *NotifyingBuffer) Write(p []byte) (n int, err error) {
	defer b.broadcast()

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0, ErrClosedWriter
	}

	b.buff = append(b.buff, p...)
	return len(p), nil
}

// Close implements io.Closer.
func (b *NotifyingBuffer) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	b.mu.Unlock()

	close(b.notify) // broadcast, but do not add a new notify chan

	return nil
}

// broadcast to notify any listeners of a change within the buffer.
func (b *NotifyingBuffer) broadcast() {
	b.mu.Lock()

	if b.closed {
		b.mu.Unlock()
		return
	}

	notify := b.notify
	b.notify = make(chan struct{})
	b.mu.Unlock()

	// broadcast while unlocked
	close(notify)
}

// Wait returns a channel that is closed if the buffer changed.
func (b *NotifyingBuffer) Wait() <-chan struct{} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.notify
}

// Status returns the size of the current buffer and closed is true
// if this buffer is closed and no more data will be written to it.
func (b *NotifyingBuffer) Status() (size int, closed bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.buff), b.closed
}

// NotifyingBufferReader is a utility type for reading from [NotifyingBuffer].
// The reader always begins at the start of the buffer and automatically
// handles cursor movement and blocking.
type NotifyingBufferReader struct {
	offs  int
	nb    *NotifyingBuffer
	close chan struct{}
}

// Read implements io.Reader.
func (r *NotifyingBufferReader) Read(p []byte) (n int, err error) {
	for blen, closed := r.nb.Status(); r.offs >= blen && !closed; blen, closed = r.nb.Status() {
		select {
		case <-r.close:
			return 0, io.EOF
		case <-r.nb.Wait():
		}
	}

	if blen, closed := r.nb.Status(); r.offs >= blen && closed {
		return 0, io.EOF
	}

	r.nb.mu.RLock()
	defer r.nb.mu.RUnlock()

	n = copy(p, r.nb.buff[r.offs:])
	r.offs += n
	return
}

// Close implements io.Closer.
func (r *NotifyingBufferReader) Close() error {
	close(r.close)
	return nil
}
