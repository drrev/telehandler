package work

import (
	"sync"
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
// Jobs should only be instantiated with [NewJob].
type Job struct {
	ID        uuid.UUID
	owner     string
	cmd       string
	args      []string
	startTime time.Time
	endTime   time.Time
	state     JobState
	exitCode  int
	mu        *sync.Mutex
}

// NewJob creates a job with a random UUID using the given args.
func NewJob(owner string, cmd string, args []string, opts ...Option) *Job {
	j := &Job{
		ID:    uuid.New(),
		owner: owner,
		cmd:   cmd,
		args:  args,
		mu:    &sync.Mutex{},
	}

	for _, fn := range opts {
		fn(j)
	}

	return j
}

func (j *Job) Identity() uuid.UUID {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.ID
}

func (j *Job) Parent() string {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.owner
}

func (j *Job) LoadCmd() string {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.cmd
}

func (j *Job) LoadArgs() []string {
	j.mu.Lock()
	defer j.mu.Unlock()
	// create a new slice to avoid leaking a ref to args
	return append([]string{}, j.args...)
}

func (j *Job) LoadStartTime() time.Time {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.startTime
}

func (j *Job) StoreStartTime(v time.Time) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.startTime = v
}

func (j *Job) LoadEndTime() time.Time {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.endTime
}

func (j *Job) StoreEndTime(v time.Time) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.endTime = v
}

func (j *Job) LoadState() JobState {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.state
}

func (j *Job) StoreState(v JobState) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.state = v
}

func (j *Job) Running() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.state == Running
}

func (j *Job) LoadExitCode() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.exitCode
}

func (j *Job) StoreExitCode(v int) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.exitCode = v
}

type Option func(j *Job)

func WithStartTime(startTime time.Time) Option {
	return func(j *Job) {
		j.startTime = startTime
	}
}

func WithendTime(endTime time.Time) Option {
	return func(j *Job) {
		j.endTime = endTime
	}
}

func Withstate(state JobState) Option {
	return func(j *Job) {
		j.state = state
	}
}

func WithexitCode(exitCode int) Option {
	return func(j *Job) {
		j.exitCode = exitCode
	}
}
