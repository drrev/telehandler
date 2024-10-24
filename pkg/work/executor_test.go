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
)

func TestExecutor_Start(t *testing.T) {
	t.Parallel()
	noopStart := func(v *int, err error) commandStarter {
		return func(c *exec.Cmd, done func(exitCode int)) error {
			return nil
		}
	}
	mockStart := func(v *int, err error) commandStarter {
		return func(c *exec.Cmd, done func(exitCode int)) error {
			if err == nil {
				done(0)
			} else {
				done(1)
			}
			*v++
			return err
		}
	}
	mockStartNoDone := func(v *int, err error) commandStarter {
		return func(c *exec.Cmd, done func(exitCode int)) error {
			*v++
			return err
		}
	}

	type fields struct {
		contexts map[string]*execContext
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
		want      Job
		wantCalls int
		injectErr error
	}{
		{
			name:    "existing non-running job",
			fields:  fields{contexts: map[string]*execContext{"": {}}},
			args:    args{j: Job{Name: ""}},
			wantErr: utils.ErrorTextContains(t, "invalid state"),
			startFn: noopStart,
		},
		{
			name:    "existing running job",
			fields:  fields{contexts: map[string]*execContext{"": {Job: Job{State: Running}}}},
			args:    args{j: Job{Name: ""}},
			want:    Job{State: Running},
			wantErr: utils.NoError(t),
			startFn: noopStart,
		},
		{
			name:      "start new job immediate exit",
			fields:    fields{contexts: make(map[string]*execContext)},
			args:      args{j: Job{Name: ""}},
			wantErr:   utils.NoError(t),
			startFn:   mockStart,
			want:      Job{StartTime: time.Now(), State: Completed},
			wantCalls: 1,
		},
		{
			name:      "start new job",
			fields:    fields{contexts: make(map[string]*execContext)},
			args:      args{j: Job{Name: ""}},
			wantErr:   utils.NoError(t),
			startFn:   mockStartNoDone,
			want:      Job{StartTime: time.Now(), State: Running},
			wantCalls: 1,
		},
		{
			name:      "start new job with error",
			fields:    fields{contexts: make(map[string]*execContext)},
			args:      args{j: Job{Name: ""}},
			wantErr:   utils.ErrorTextContains(t, "testing error"),
			startFn:   mockStart,
			injectErr: errors.New("testing error"),
			want:      Job{StartTime: time.Now(), State: Failed, ExitCode: 1},
			wantCalls: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startCalls := 0
			m := &Executor{
				mu:       sync.RWMutex{},
				cgroot:   "/tmp",
				contexts: tt.fields.contexts,
				startCmd: tt.startFn(&startCalls, tt.injectErr),
			}
			got, err := m.Start(tt.args.j)

			if !tt.wantErr(err) {
				t.Errorf("Executor.Start() error = %v", err)
			}

			if tt.want.StartTime.IsZero() != got.StartTime.IsZero() {
				t.Errorf("Executor.Start() invalid StartTime %v", got.StartTime)
			}

			// ignore times for deep equals
			tt.want.StartTime = got.StartTime
			tt.want.EndTime = got.EndTime
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Executor.Start() = %+v, want %+v", got, tt.want)
			}

			if startCalls != tt.wantCalls {
				t.Errorf("Executor.Start() got %v calls, expected %v", startCalls, tt.wantCalls)
			}
		})
	}
}

func TestExecutor_Stop(t *testing.T) {
	t.Parallel()
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		contexts map[string]*execContext
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
			contexts: map[string]*execContext{"": {}},
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
	t.Parallel()
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		contexts map[string]*execContext
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
			contexts: map[string]*execContext{"": {Job: Job{}}},
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

func TestExecutor_OpenReader(t *testing.T) {
	t.Parallel()
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		contexts map[string]*execContext
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
			contexts: map[string]*execContext{"": {Job: Job{}, buf: safe.NewNotifyingBuffer()}},
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
			got, err := m.OpenReader(tt.args.id)
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
	t.Parallel()
	m := &Executor{
		mu:       sync.RWMutex{},
		cgroot:   "",
		contexts: make(map[string]*execContext),
	}
	if err := m.Wait(""); !utils.ErrorTextContains(t, "no job found")(err) {
		t.Errorf("Executor.Wait() error = %v", err)
	}

	m.contexts = map[string]*execContext{"": {}}
	if err := m.Wait(""); err != nil {
		t.Errorf("Executor.Wait() error = %v", err)
	}

	buf := safe.NewNotifyingBuffer()
	_, _ = buf.Write([]byte{0, 1})

	ec := &execContext{
		Job: Job{State: Running},
		buf: buf,
	}

	m.contexts = map[string]*execContext{"": ec}

	done := make(chan struct{})
	go func() {
		_ = m.Wait("")
		close(done)
	}()

	select {
	case <-done:
		t.Error("Executor.Wait() returned unexpectedly")
	default:
	}

	ec.m.Lock()
	ec.State = Completed
	ec.m.Unlock()

	select {
	case <-time.After(200 * time.Millisecond):
		t.Error("Executor.Wait() did not return after buf close")
	case <-done:
	}
}
