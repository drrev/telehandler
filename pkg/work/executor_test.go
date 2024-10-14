package work

import (
	"errors"
	"os/exec"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/drrev/telehandler/pkg/safe"
	"github.com/drrev/telehandler/tests/utils"
	"github.com/google/uuid"
)

func TestExecutor_Start(t *testing.T) {
	noopStart := func(v *int, err error) commandStarter {
		return func(c *exec.Cmd, done func(exitCode int)) error {
			return nil
		}
	}
	mockStart := func(v *int, err error) commandStarter {
		return func(c *exec.Cmd, done func(exitCode int)) error {
			*v++
			return err
		}
	}

	type fields struct {
		cgroot   string
		contexts map[uuid.UUID]*execContext
	}
	type args struct {
		j Job
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   func(error) bool
		startFn   func(*int, error) commandStarter
		wantCalls int
		injectErr error
	}{
		{
			name:    "existing non-running job",
			fields:  fields{cgroot: "/tmp", contexts: map[uuid.UUID]*execContext{uuid.Max: {}}},
			args:    args{j: Job{ID: uuid.Max}},
			wantErr: utils.ErrorTextContains(t, "invalid state"),
			startFn: noopStart,
		},
		{
			name:    "existing running job",
			fields:  fields{cgroot: "/tmp", contexts: map[uuid.UUID]*execContext{uuid.Max: {Job: Job{State: Running}}}},
			args:    args{j: Job{ID: uuid.Max}},
			wantErr: utils.NoError(t),
			startFn: noopStart,
		},
		{
			name:      "start new job",
			fields:    fields{cgroot: "/tmp", contexts: make(map[uuid.UUID]*execContext)},
			args:      args{j: Job{ID: uuid.Max}},
			wantErr:   utils.NoError(t),
			startFn:   mockStart,
			wantCalls: 1,
		},
		{
			name:      "start new job with error",
			fields:    fields{cgroot: "/tmp", contexts: make(map[uuid.UUID]*execContext)},
			args:      args{j: Job{ID: uuid.Max}},
			wantErr:   utils.ErrorTextContains(t, "testing error"),
			startFn:   mockStart,
			wantCalls: 1,
			injectErr: errors.New("testing error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startCalls := 0
			m := &Executor{
				mu:       sync.RWMutex{},
				cgroot:   tt.fields.cgroot,
				contexts: tt.fields.contexts,
				startCmd: tt.startFn(&startCalls, tt.injectErr),
			}
			if err := m.Start(tt.args.j); !tt.wantErr(err) {
				t.Errorf("Executor.Start() error = %v", err)
			}
			if startCalls != tt.wantCalls {
				t.Errorf("Executor.Start() got %v calls, expected %v", startCalls, tt.wantCalls)
			}
		})
	}
}

func TestExecutor_Stop(t *testing.T) {
	type args struct {
		id uuid.UUID
	}
	tests := []struct {
		name     string
		contexts map[uuid.UUID]*execContext
		args     args
		wantErr  func(error) bool
	}{
		{
			name:    "missing job",
			args:    args{},
			wantErr: utils.ErrorTextContains(t, "no job found"),
		},
		{
			name:     "stop job",
			args:     args{},
			contexts: map[uuid.UUID]*execContext{uuid.Nil: {}},
			wantErr:  utils.ErrorTextContains(t, "invalid state"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Executor{
				mu:       sync.RWMutex{},
				cgroot:   "",
				contexts: tt.contexts,
			}
			if err := m.Stop(tt.args.id); !tt.wantErr(err) {
				t.Errorf("Executor.Stop() error = %v", err)
			}
		})
	}
}

func TestExecutor_Lookup(t *testing.T) {
	type args struct {
		id uuid.UUID
	}
	tests := []struct {
		name     string
		contexts map[uuid.UUID]*execContext
		args     args
		want     Job
		wantErr  func(error) bool
	}{
		{
			name:    "not found",
			wantErr: utils.ErrorTextContains(t, "no job found"),
		},
		{
			name:     "found",
			contexts: map[uuid.UUID]*execContext{uuid.Nil: {Job: Job{}}},
			wantErr:  utils.NoError(t),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Executor{
				mu:       sync.RWMutex{},
				cgroot:   "",
				contexts: tt.contexts,
			}
			got, err := m.Lookup(tt.args.id)
			if !tt.wantErr(err) {
				t.Errorf("Executor.Lookup() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Executor.Lookup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutor_Watch(t *testing.T) {
	type args struct {
		id uuid.UUID
	}
	tests := []struct {
		name     string
		contexts map[uuid.UUID]*execContext
		args     args
		want     *safe.NotifyingBufferReader
		wantErr  func(error) bool
	}{
		{
			name:    "not found",
			wantErr: utils.ErrorTextContains(t, "no job found"),
			want:    nil,
		},
		{
			name:     "found",
			contexts: map[uuid.UUID]*execContext{uuid.Nil: {Job: Job{}, buf: safe.NewNotifyingBuffer()}},
			wantErr:  utils.NoError(t),
			want:     safe.NewNotifyingBuffer().Reader(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Executor{
				mu:       sync.RWMutex{},
				cgroot:   "",
				contexts: tt.contexts,
			}
			got, err := m.Watch(tt.args.id)
			if !tt.wantErr(err) {
				t.Errorf("Executor.Watch() error = %v", err)
				return
			}
			if (tt.want != nil) != (got != nil) {
				t.Errorf("Executor.Watch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecutor_Wait(t *testing.T) {
	m := &Executor{
		mu:       sync.RWMutex{},
		cgroot:   "",
		contexts: make(map[uuid.UUID]*execContext),
	}
	if err := m.Wait(uuid.Nil); !utils.ErrorTextContains(t, "no job found")(err) {
		t.Errorf("Executor.Wait() error = %v", err)
	}

	m.contexts = map[uuid.UUID]*execContext{uuid.Nil: {}}
	if err := m.Wait(uuid.Nil); err != nil {
		t.Errorf("Executor.Wait() error = %v", err)
	}

	buf := safe.NewNotifyingBuffer()
	buf.Write([]byte{0, 1})

	m.contexts = map[uuid.UUID]*execContext{uuid.Nil: {
		buf: buf,
	}}

	done := make(chan struct{})
	go func() {
		m.Wait(uuid.Nil)
		close(done)
	}()

	select {
	case <-done:
		t.Error("Executor.Wait() returned unexpectedly")
	default:
	}

	buf.Close()

	select {
	case <-time.After(200 * time.Millisecond):
		t.Error("Executor.Wait() did not return after buf close")
	case <-done:
	}
}
