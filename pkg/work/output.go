package work

import (
	"context"
	"io"

	"github.com/drrev/telehandler/internal/safe"
)

var _ io.Reader = (*OutputReader)(nil)

// OutputReader implements [io.Reader] for a [Job].
//
// This OutputReader presents STDIO/STDERR in a single
// reader.
type OutputReader struct {
	// seq is the current sync.CondChan sequence number
	seq int64
	off int64
	max int64
	// ctx is copied as a field to catch ctx.Done()
	// ctx is not used for any other reason
	ctx context.Context
	out *safe.Buffer
}

// newOutputReader wraps [safe.Buffer] to implement [io.Reader].
func newOutputReader(ctx context.Context, out *safe.Buffer) *OutputReader {
	return &OutputReader{
		ctx: ctx,
		out: out,
		max: int64(out.Len()),
	}
}

// Read implements io.Reader.
// If the underlying reader is closed, io.EOF is returned.
//
// All jobs have STDIO and STDERR muxed into a single stream;
// therefore, a read from this reader will return all
// interpersed output data.
func (o *OutputReader) Read(p []byte) (n int, err error) {
	if o.off >= o.max {
		o.seq = o.out.Wait(o.ctx, o.seq)
		o.max = int64(o.out.Len())
	}

	n, err = o.out.ReadAt(p, o.off)
	o.off += int64(n)

	return
}
