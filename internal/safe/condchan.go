package safe

import "context"

// CondChan is a simplified channel-based implementation of
// sync.Cond for broadcasting changes with context cancellation support.
//
// CondChan should only be instantiated with [NewCond].
type CondChan struct {
	ch chan state
}

type state struct {
	seq    int64
	listen chan struct{}
}

// Wait suspends the calling goroutine until a change occurs.
//
// It is required that seq is 0 on first wait. On subsequent Wait calls,
// the newSeq returned by the previous wait should be used for the next Wait.
//
// The last seq seen is returned immediately if ctx is cancelled.
func (c *CondChan) Wait(ctx context.Context, seq int64) (newSeq int64) {
	st := <-c.ch

	// if state changed before this Wait call, eagerly return
	if st.seq != seq {
		c.ch <- st
		return st.seq
	}

	if st.listen == nil {
		st.listen = make(chan struct{})
	}
	c.ch <- st

	// race between ctx and change to allow cancellation
	select {
	case <-ctx.Done():
		return seq
	case <-st.listen:
		return seq + 1
	}
}

// Broadcast wakes all goroutines waiting on c.
func (c *CondChan) Broadcast() {
	st := <-c.ch
	if st.listen != nil {
		// closing chan broadcasts a change
		close(st.listen)
	}
	c.ch <- state{st.seq + 1, nil}
}

// NewCond creates a new [CondChan] that is ready for use.
func NewCond() *CondChan {
	st := make(chan state, 1)
	st <- state{seq: 0}
	return &CondChan{st}
}
