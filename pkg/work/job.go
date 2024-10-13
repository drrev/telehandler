package work

import (
	"time"

	"github.com/google/uuid"
)

// JobState is used to demarcate where a [Job]
// is in its lifecycle.
type JobState string

const (
	// The job is currently running and active.
	Running JobState = "JOB_STATE_RUNNING"
	// The job failed during execution.
	// If a job is in this state, request the job output with TailJobOutput to discern the cause of failure.
	Failed JobState = "JOB_STATE_FAILED"
	// The job ran to completion and exited successfully.
	Completed JobState = "JOB_STATE_COMPLETED"
	// The job was stopped by a user before completing execution.
	Stopped JobState = "JOB_STATE_STOPPED"
)

// Job represents a command context.
type Job struct {
	ID uuid.UUID
	// Owner that created this Job.
	Owner string
	// Cmd path to an executable to run for this Job.
	Cmd string
	// Args passed to the subprocess.
	Args []string
	// StartTime of when the subprocess began execution.
	StartTime time.Time
	// EndTime is the time that the job terminated.
	// This field is only valid if State != Running.
	EndTime time.Time
	State   JobState
	// ExitCode captures the exit_code of the subprocess.
	// This field is only valid if State != Running.
	ExitCode int
}

// NewJob creates a [Job] with a randomly generated UUID and the given
// owner, cmd, and args.
func NewJob(owner string, cmd string, args []string) *Job {
	return &Job{
		ID:    uuid.New(),
		Owner: owner,
		Cmd:   cmd,
		Args:  args,
	}
}

// Identity returns the unique [Job] identifier.
func (j *Job) Identity() uuid.UUID {
	return j.ID
}

// Parent returns the owner of the [Job].
func (j *Job) Parent() string {
	return j.Owner
}

// Running is a convenience function to check
// if the [Job] is [Running].
func (j *Job) Running() bool {
	return j.State == Running
}

// setTerminate sets the correct State, ExitCode,
// and EndTime all at once based on code and stopped.
func (j *Job) setTerminate(code int, stopped bool) {
	j.ExitCode = code
	j.EndTime = time.Now()

	if stopped {
		j.State = Stopped
		return
	}

	if code == 0 {
		j.State = Completed
	} else {
		j.State = Failed
	}
}

// setRunning is a simple helper for
// syncronizing StartTime and State.
func (j *Job) setRunning() {
	j.StartTime = time.Now()
	j.State = Running
}
