package work

import (
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/drrev/telehandler/pkg/safe"
	"github.com/google/uuid"
)

// commandStarter starts and waits for execution of commands,
// then calls done with the process exit_code.
type commandStarter func(c *exec.Cmd, done func(exitCode int)) error

// Executor is a thread-safe [Job] manager.
// Each Job runs in a separate subprocess.
// All Jobs are resource limited using cgroup v2 and
// placed in separate PID, User, and Network Linux namespaces.
//
// Despite namespace isolation, this does not create a full
// sandbox. Mounts are not fully isolated from the host and Jobs
// can still interact with any mounted partitions on the host system.
//
// Network is fully isolated. Jobs have no non-loopback network interfaces,
// thus no network connectivity. The hostname is forced to "sandbox".
//
// See [NewExecutor].
type Executor struct {
	mu       sync.RWMutex
	cgroot   string
	contexts map[uuid.UUID]*execContext
	startCmd commandStarter
}

// NewExecutor creates an initialized [Executor] ready for use.
func NewExecutor(cgroupRoot string) *Executor {
	return &Executor{
		mu:       sync.RWMutex{},
		cgroot:   cgroupRoot,
		contexts: make(map[uuid.UUID]*execContext),
		startCmd: startCmd,
	}
}

// Start the given [Job]. An error is returned if the Job could not be started.
//
// [ErrInvalidState] is returned if the Job already exists with a non-running status.
//
// Calling Start on a Job that is already running is a no-op.
//
// This operation is stateful. If this call is successful, a copy of Job is
// maintained internally. Use [Executor.Find] to lookup any existing Jobs for the
// latest state.
func (m *Executor) Start(j Job) (Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ec, err := m.lookupContext(j.ID)
	if err == nil {
		if !ec.Running() {
			return ec.Job, invalidJobState(ec.State)
		}
		return ec.Job, nil
	}

	ec = &execContext{
		Job: j,
		m:   sync.Mutex{},
		buf: safe.NewNotifyingBuffer(),
	}
	m.contexts[j.ID] = ec

	// make a new cgroup for the job specifically
	cgroupJob := filepath.Join(m.cgroot, j.ID.String())
	cmd, cancel := makeCommand(ec.buf, cgroupJob, j.Cmd, j.Args...)
	ec.stop = cancel

	if err := m.startCmd(cmd, ec.exit); err != nil {
		return ec.Job, err
	}

	ec.StartTime = time.Now()
	ec.State = Running

	return ec.Job, nil
}

// Stop a Job using the provided jobID. A non-nil error is returned
// if the Job failed to exit cleanly, or [ErrJobNotFound] if no job
// with the given jobID exists.
//
// Calling Stop on a non-running Job is a no-op.
func (m *Executor) Stop(id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ec, err := m.lookupContext(id)
	if err != nil {
		return err
	}
	return ec.interrupt()
}

// Lookup returns a copy of any [Job] found. If no Job is found, a [ErrJobNotFound]
// is returned and the Job value is zero.
func (m *Executor) Lookup(id uuid.UUID) (Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ec, err := m.lookupContext(id)
	if err != nil {
		return Job{}, err
	}
	return ec.Job, nil
}

// OpenReader returns a [safe.NotifyingBufferReader] for reading STDOUT and STDERR from a [Job]. This method
// may be used to get output from Jobs in any state.
func (m *Executor) OpenReader(id uuid.UUID) (*safe.NotifyingBufferReader, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ec, err := m.lookupContext(id)
	if err != nil {
		return nil, err
	}

	return ec.buffer().Reader(), nil
}

// Wait for a [Job] to terminate.
func (m *Executor) Wait(id uuid.UUID) error {
	m.mu.RLock()

	ec, err := m.lookupContext(id)
	if err != nil {
		return err
	}

	if !ec.Running() {
		return nil
	}

	buf := ec.buffer()

	m.mu.RUnlock()

	for _, closed := buf.Status(); !closed; _, closed = buf.Status() {
		buf.Wait()
	}

	return nil
}

// lookupContext is a thread-safe method for finding execContext by Job ID.
func (m *Executor) lookupContext(id uuid.UUID) (*execContext, error) {
	ec, ok := m.contexts[id]
	if !ok {
		return nil, jobNotFound(id)
	}

	return ec, nil
}
