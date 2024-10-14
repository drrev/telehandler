package work

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrCannotStop = errors.New("cannot stop process")

func invalidJobState(s JobState) *ErrInvalidJobState {
	return &ErrInvalidJobState{s}
}

// ErrInvalidJobState is returned when an operation is attempted against any Job
// that is in the incorrect state.
type ErrInvalidJobState struct {
	state JobState
}

func (e *ErrInvalidJobState) String() string {
	return e.Error()
}

// Error implements error.
func (e *ErrInvalidJobState) Error() string {
	return fmt.Sprintf("invalid state '%s'", e.state)
}

func jobNotFound(id uuid.UUID) *ErrJobNotFound {
	return &ErrJobNotFound{id}
}

// ErrJobNotFound is returned if no job is found for a given ID.
type ErrJobNotFound struct {
	id uuid.UUID
}

// Error implements error.
func (e *ErrJobNotFound) Error() string {
	return fmt.Sprintf("no job found with id='%v'", e.id)
}
