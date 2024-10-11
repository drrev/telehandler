package work

import (
	"context"

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
	out *safe.Buffer
}

// newOutputReader wraps [safe.Buffer] to implement [io.Reader].
func newOutputReader(out *safe.Buffer) *OutputReader {
	return &OutputReader{
		out: out,
		max: int64(out.Len()),
	}
}

// If the underlying reader is closed, io.EOF is returned.
//
// All jobs have STDIO and STDERR muxed into a single stream;
// therefore, a read from this reader will return all
// interpersed output data.
func (o *OutputReader) Read(ctx context.Context, p []byte) (n int, err error) {
	if o.off >= o.max {
		o.seq = o.out.Wait(ctx, o.seq)
		o.max = int64(o.out.Len())
	}

	n, err = o.out.ReadAt(p, o.off)
	o.off += int64(n)

	return
}
