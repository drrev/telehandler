package work

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/drrev/telehandler/pkg/safe"
)

type execContext struct {
	Job
	m       sync.Mutex
	buf     *safe.NotifyingBuffer
	stop    func()
	stopped atomic.Bool
}

// buffer is a thread-safe method for getting the [safe.NotifyingBuffer].
func (e *execContext) buffer() *safe.NotifyingBuffer {
	e.m.Lock()
	defer e.m.Unlock()
	return e.buf
}

// interrupt the [Job].
// This operation is thread-safe.
//
// An error is returned if the Job is not running.
// If the job cannot be stopped, [ErrCannotStop] is returned.
func (e *execContext) interrupt() error {
	e.m.Lock()
	defer e.m.Unlock()
	if !e.Running() {
		return invalidJobState(e.State)
	}

	if e.stop == nil {
		return ErrCannotStop
	}

	e.stop()
	e.stop = nil
	e.stopped.Store(true)
	return nil
}

// exit performs all bookkeeping required when the [Job] exits.
// This operation is thread-safe.
func (e *execContext) exit(exitCode int) {
	e.m.Lock()
	defer e.m.Unlock()

	_ = e.buf.Close()
	e.EndTime = time.Now()
	e.ExitCode = exitCode

	if exitCode == 0 {
		e.State = Completed
	} else {
		e.State = Failed
	}
	if e.stopped.Load() {
		e.State = Stopped
	}

	slog.Info("Job terminated", slog.Any("job", e.LogValue()))
}
