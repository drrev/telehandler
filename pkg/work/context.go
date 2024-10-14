package work

import (
	"sync"
	"time"

	"github.com/drrev/telehandler/pkg/safe"
)

type execContext struct {
	Job
	m    sync.Mutex
	buf  *safe.NotifyingBuffer
	stop func()
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
	} else if exitCode < 0 {
		// interrupted can only happen if
		// Telehandler interrupted the process
		// as a result of a Stop() request
		e.State = Stopped
	} else {
		e.State = Failed
	}
}
