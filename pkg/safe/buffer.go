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
	// TODO: In a real production-level application, storing in memory is
	// a bad idea. Commands can--and will--OOM this process. In fact,
	// when running this version, there are numerous problems with memory
	// management. Namely, the go GC does not release allocated memory quickly
	// unless there is pressure to do so from the OS. This can easily result in the following:
	//
	// # snippet from /proc/<pid>/stats:
	// ...
	// VmPeak: 18124988 kB
	// VmSize: 18123964 kB
	// VmLck:         0 kB
	// VmPin:         0 kB
	// VmHWM:  12953640 kB
	// VmRSS:   9937416 kB
	// ...
	//
	// The above memory allocations apply to the below runtime:
	//
	// (pprof) top
	// Showing nodes accounting for 5630.53MB, 99.86% of 5638.62MB total
	// Dropped 57 nodes (cum <= 28.19MB)
	// Showing top 10 nodes out of 33
	// 	  flat  flat%   sum%        cum   cum%
	//  4664.79MB 82.73% 82.73%  4667.79MB 82.78%  github.com/drrev/telehandler/pkg/safe.(*NotifyingBuffer).Write
	//   518.88MB  9.20% 91.93%   754.32MB 13.38%  github.com/drrev/telehandler/internal/foreman.(*Service).WatchJobOutput
	//   227.34MB  4.03% 95.96%   227.34MB  4.03%  google.golang.org/grpc/mem.(*simpleBufferPool).Get
	//   140.65MB  2.49% 98.46%   140.65MB  2.49%  google.golang.org/grpc/mem.NewTieredBufferPool.newSizedBufferPool.func1
	//    78.87MB  1.40% 99.86%    78.87MB  1.40%  crypto/tls.(*halfConn).encrypt
	// 		 0     0% 99.86%    78.87MB  1.40%  crypto/tls.(*Conn).Write
	// 		 0     0% 99.86%    78.87MB  1.40%  crypto/tls.(*Conn).writeRecordLocked
	// 		 0     0% 99.86%   754.32MB 13.38%  github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1._ForemanService_WatchJobOutput_Handler
	// 		 0     0% 99.86%    78.87MB  1.40%  golang.org/x/net/http2.(*Framer).WriteData
	// 		 0     0% 99.86%    78.87MB  1.40%  golang.org/x/net/http2.(*Framer).WriteDataPadded
	//
	//
	// This of course does not outline everything; however, no matter how aggressively the below is optimized, it cannot support a large number of concurrent Jobs
	// that generate a lot of output.
	//
	// To resolve that problem, it would likely be sufficient enough to write the subprocess output to a tmpfile, which can then be streamed. In the real world,
	// this is what we would want to do anyway, so we can store jobs and their results in persistent storage, then users can check the result of those jobs later.
	//
	// In that case, Telehandler would need to be run in some kind of... sandbox... with namespaces and limits for disk utilization, etc. I wonder where we could
	// find one of those?
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
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0, ErrClosedWriter
	}

	b.buff = append(b.buff, p...)
	b.broadcast()

	return len(p), nil
}

// Close implements io.Closer.
func (b *NotifyingBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.closed = true
	close(b.notify)

	return nil
}

// broadcast to notify any listeners of a change within the buffer.
// This method is NOT thread-safe.
func (b *NotifyingBuffer) broadcast() {
	if b.closed {
		return
	}

	close(b.notify)
	b.notify = make(chan struct{})
}

// Wait returns a channel that is closed if the buffer changed.
func (b *NotifyingBuffer) Wait() (notify <-chan struct{}) {
	b.mu.RLock()
	notify = b.notify
	b.mu.RUnlock()
	return
}

// Status returns the size of the current buffer and closed is true
// if this buffer is closed and no more data will be written to it.
func (b *NotifyingBuffer) Status() (size int, closed bool) {
	b.mu.RLock()
	size = len(b.buff)
	closed = b.closed
	b.mu.RUnlock()
	return
}

// NotifyingBufferReader is a utility type for reading from [NotifyingBuffer].
// The reader always begins at the start of the buffer and automatically
// handles cursor movement and blocking.
type NotifyingBufferReader struct {
	offs  int
	nb    *NotifyingBuffer
	close chan struct{}
	once  sync.Once
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
	n = copy(p, r.nb.buff[r.offs:])
	r.nb.mu.RUnlock()

	r.offs += n
	return
}

// Close implements io.Closer.
func (r *NotifyingBufferReader) Close() error {
	r.once.Do(func() {
		close(r.close)
	})
	return nil
}
