package work

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/drrev/telehandler/pkg/safe"
	"github.com/google/uuid"
)

// execContext is a full execution context for a Job.
// Once a Job is started, this is the source of truth for Job state.
type execContext struct {
	job    *Job
	cmd    *exec.Cmd
	output *safe.NotifyingBuffer
}

// kill the underlying job and ensure bookkeeping.
func (e *execContext) kill() error {
	// if the job is already terminated, do nothing
	if !e.job.Running() {
		return nil
	}

	e.job.State = Stopped
	_ = e.output.Close()
	return e.cmd.Process.Kill()
}

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
	cmds       map[uuid.UUID]*execContext
	cgroupPath string
	mu         *sync.Mutex
}

// NewExecutor creates an initialized [Executor] ready for use.
func NewExecutor(cgroupPath string) *Executor {
	return &Executor{make(map[uuid.UUID]*execContext), cgroupPath, &sync.Mutex{}}
}

// updateJob replaces the Job in cache with the given Job.
func (e *Executor) updateJob(j Job) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if v, ok := e.cmds[j.ID]; ok {
		v.job = &j
	}
}

// Find returns a copy of any Job found. If no Job is found, a [ErrJobNotFound]
// is returned and the Job value is zero.
func (e *Executor) Find(jobID uuid.UUID) (job Job, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ec, ok := e.cmds[jobID]
	if !ok || ec.job == nil {
		err = jobNotFound(jobID)
		return
	}

	job = *ec.job
	return
}

// Running returns v=true, ok=true if the given jobID belongs to a running job.
func (e *Executor) Running(jobID uuid.UUID) (v bool, ok bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	ec, ok := e.cmds[jobID]
	if !ok {
		return false, false
	}
	return ec.job.Running(), true
}

// Start the given [Job] and return an updated Job copy. An error is returned
// if the Job could not be started.
//
// [ErrInvalidState] is returned if the Job already exists with a non-running status.
//
// Calling Start on a Job that is already running is a no-op.
//
// This operation is stateful. If this call is successful, a copy of Job is
// maintained internally. Use [Executor.Find] to lookup any existing Jobs for the
// latest state.
func (e *Executor) Start(j Job) (job Job, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	logger := slog.With(
		slog.Group("job",
			slog.String("owner", j.Owner),
			slog.String("id", j.ID.String()),
			slog.String("cmd", j.Cmd),
			slog.Any("args", j.Args),
		),
	)

	refj := &j

	if v, ok := e.cmds[refj.ID]; ok {
		if v.job.Running() {
			logger.Debug("Job is already running")
			return *v.job, nil
		}
		logger.Warn("Start called on terminated job")
		return *v.job, invalidJobState(refj.State)
	}

	// buffer up to 65535 by default, this will grow up to 128TiB automatically
	buf := safe.NewNotifyingBuffer()
	cmd := makeCommand(buf, filepath.Join(e.cgroupPath, refj.ID.String()), *refj)

	// save job regardless of the outcome
	ec := &execContext{
		job:    refj,
		cmd:    cmd,
		output: buf,
	}
	e.cmds[refj.ID] = ec

	// TODO: make start/stop/etc. into tight methods on execContext to prevent
	// bookkeeping errors for job <-> cmd state
	if err := cmd.Start(); err != nil {
		refj.setTerminate(cmd.ProcessState.ExitCode(), false)
		buf.Close()
		logger.Error("Job failed to start", slog.Int("exit_code", refj.ExitCode), slog.Any("error", err))
		return j, err
	}
	refj.setRunning()

	// wait so process doesn't become a zombie
	go e.wait(logger, ec)

	job = *refj
	return
}

// wait is a helper for waiting on an execContext to finish and
// perform any necessary bookkeeping.
func (e *Executor) wait(logger *slog.Logger, ec *execContext) {
	// wait so process doesn't become a zombie
	defer ec.output.Close()

	// *ec.job MUST NOT be updated in this func, it is not locked
	job := *ec.job

	err := ec.cmd.Wait()
	if err != nil {
		logger.Error("Wait failed on process", slog.Any("error", err))

		var exiterr *exec.ExitError
		if errors.As(err, &exiterr) {
			job.setTerminate(exiterr.ExitCode(), false)
			e.updateJob(job)
			return
		}
	}

	job.setTerminate(ec.cmd.ProcessState.ExitCode(), false)
	logger.Info(
		"Job finished",
		slog.Int("exit_code", job.ExitCode),
		slog.Duration("duration", job.EndTime.Sub(job.StartTime)),
	)
	e.updateJob(job)
}

// Stop a Job using the provided jobID. A non-nil error is returned
// if the Job failed to exit cleanly, or [ErrJobNotFound] if no job
// with the given jobID exists.
//
// Calling Stop on a non-running Job is a no-op.
func (e *Executor) Stop(jobID uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	v, ok := e.cmds[jobID]
	if !ok {
		return jobNotFound(jobID)
	}

	return v.kill()
}

// Output returns a [OutputReader] for reading STDOUT and STDERR from a Job. This method
// may be used to get output from Jobs in any state.
func (e *Executor) Output(jobID uuid.UUID) (*safe.NotifyingBufferReader, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	v, ok := e.cmds[jobID]
	if !ok {
		return nil, fmt.Errorf("unknown job '%v'", jobID)
	}

	return v.output.Reader(), nil
}

// Teardown sends the given signal to all running Jobs and cleans up. This should only
// be used when the calling process is exiting.
func (e *Executor) Teardown() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	errs := make([]error, 0, len(e.cmds))
	for _, v := range e.cmds {
		if !v.cmd.ProcessState.Exited() {
			errs = append(errs, v.cmd.Process.Kill())
		}
	}

	return errors.Join(errs...)
}
