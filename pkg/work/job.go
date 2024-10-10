package work

import (
	"time"

	"github.com/google/uuid"
)

type JobState string

const (
	Running   JobState = "JOB_STATE_RUNNING"
	Failed    JobState = "JOB_STATE_FAILED"
	Completed JobState = "JOB_STATE_COMPLETED"
	Stopped   JobState = "JOB_STATE_STOPPED"
)

// Job represents a command context.
type Job struct {
	ID        uuid.UUID
	Owner     string
	Cmd       string
	Args      []string
	StartTime time.Time
	EndTime   time.Time
	State     JobState
	ExitCode  int
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
