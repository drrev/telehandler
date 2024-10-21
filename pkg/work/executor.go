package work

import (
	"log/slog"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/drrev/telehandler/pkg/safe"
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
	contexts map[string]*execContext
	startCmd commandStarter
}

// NewExecutor creates an initialized [Executor] ready for use.
func NewExecutor(cgroupRoot string) *Executor {
	return &Executor{
		mu:       sync.RWMutex{},
		cgroot:   cgroupRoot,
		contexts: make(map[string]*execContext),
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

	ec, err := m.lookupContext(j.Name)
	if err == nil {
		if !ec.Running() {
			return ec.jobSafe(), invalidJobState(ec.State)
		}
		return ec.jobSafe(), nil
	}

	ec = &execContext{
		Job: j,
		m:   sync.Mutex{},
		buf: safe.NewNotifyingBuffer(),
	}
	m.contexts[j.Name] = ec

	// make a new cgroup for the job specifically
	cgroupJob := filepath.Join(m.cgroot, filepath.Base(j.Name))
	cmd, cancel := makeCommand(ec.buf, cgroupJob, j.Cmd, j.Args...)
	ec.stop = cancel
	ec.StartTime = time.Now()
	ec.State = Running

	if err := m.startCmd(cmd, ec.exit); err != nil {
		return ec.jobSafe(), err
	}

	slog.Info("Job started", slog.Any("job", ec.LogValue()))

	return ec.jobSafe(), nil
}

// Stop a Job using the provided jobID. A non-nil error is returned
// if the Job failed to exit cleanly, or [ErrJobNotFound] if no job
// with the given jobID exists.
//
// Calling Stop on a non-running Job is a no-op.
func (m *Executor) Stop(name string) error {
	m.mu.Lock()
	ec, err := m.lookupContext(name)
	m.mu.Unlock()

	if err != nil {
		return err
	}

	return ec.interrupt()
}

// Lookup returns a copy of any [Job] found. If no Job is found, a [ErrJobNotFound]
// is returned and the Job value is zero.
func (m *Executor) Lookup(name string) (job Job, err error) {
	m.mu.RLock()
	ec, err := m.lookupContext(name)
	m.mu.RUnlock()

	if err == nil {
		job = ec.jobSafe()
	}
	return
}

// OpenReader returns a [safe.NotifyingBufferReader] for reading STDOUT and STDERR from a [Job]. This method
// may be used to get output from Jobs in any state.
func (m *Executor) OpenReader(name string) (*safe.NotifyingBufferReader, error) {
	m.mu.RLock()
	ec, err := m.lookupContext(name)
	m.mu.RUnlock()

	if err != nil {
		return nil, err
	}

	return ec.buffer().Reader(), nil
}

// Wait for a [Job] to terminate.
func (m *Executor) Wait(name string) error {
	m.mu.RLock()
	ec, err := m.lookupContext(name)
	m.mu.RUnlock()

	if err != nil {
		return err
	}

	for {
		if !ec.Running() {
			return nil
		}
	}
}

// lookupContext is a thread-safe method for finding execContext by Job ID.
func (m *Executor) lookupContext(name string) (*execContext, error) {
	ec, ok := m.contexts[name]
	if !ok {
		return nil, jobNotFound(name)
	}

	return ec, nil
}
