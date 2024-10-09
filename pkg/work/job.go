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

func NewJob(owner string, cmd string, args []string) *Job {
	return &Job{
		ID:    uuid.New(),
		Owner: owner,
		Cmd:   cmd,
		Args:  args,
	}
}

func (j *Job) Identity() uuid.UUID {
	return j.ID
}

func (j *Job) Parent() string {
	return j.Owner
}

func (j *Job) Running() bool {
	return j.State == Running
}
