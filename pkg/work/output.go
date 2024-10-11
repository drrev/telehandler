package work

import (
	"github.com/drrev/telehandler/internal/safe"
)

// ErrTooEarly is a retryable error when the underlying buffer has no more data, but
// more data may be available in the future.
var ErrTooEarly = safe.ErrTooEarly

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
	out *safe.Buffer
}

// newOutputReader wraps [safe.Buffer] to implement [io.Reader].
func newOutputReader(out *safe.Buffer) *OutputReader {
	return &OutputReader{
		out: out,
	}
}

// Read implements io.Reader.
func (o *OutputReader) Read(p []byte) (n int, err error) {
	n, err = o.out.ReadAt(p, o.off)
	o.off += int64(n)
	return
}
