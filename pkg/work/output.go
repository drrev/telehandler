package work

import (
	"io"
	"sync/atomic"

	"github.com/drrev/telehandler/internal/safe"
)

// OutputReader implements [io.Reader] for a [Job].
//
// This OutputReader presents STDIO/STDERR in a single
// reader.
//
// This reader starts reading from the process epoch.
type OutputReader struct {
	// seq is the current sync.CondChan sequence number
	seq int64
	off int64
	max int64
	// ctx is copied as a field to catch ctx.Done()
	// ctx is not used for any other reason
	out    *safe.Buffer
	closed atomic.Bool
}

// newOutputReader wraps [safe.Buffer] to implement [io.Reader].
func newOutputReader(out *safe.Buffer) *OutputReader {
	return &OutputReader{
		out: out,
	}
}

// Read implements io.Reader.
func (o *OutputReader) Read(p []byte) (n int, err error) {
	if o.closed.Load() {
		err = io.EOF
		return
	}

	n, err = o.out.ReadAt(p, o.off)
	o.off += int64(n)
	return
}

// Close implements io.Closer.
// This wakes any current `Read`.
func (o *OutputReader) Close() error {
	o.closed.Store(true)
	o.out.Wake()
	return nil
}
