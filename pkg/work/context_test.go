package work

import (
	"reflect"
	"sync"
	"testing"

	"github.com/drrev/telehandler/pkg/safe"
	"github.com/drrev/telehandler/tests/utils"
)

func Test_execContext_interrupt(t *testing.T) {
	called := func(v *int) func() {
		*v = 0
		return func() {
			*v++
		}
	}

	nilfn := func(v *int) func() {
		return nil
	}

	type fields struct {
		Job  Job
		stop func(v *int) func()
	}
	tests := []struct {
		name      string
		fields    fields
		wantCount int
		wantErr   func(e error) bool
	}{
		{
			name:    "job stopped",
			fields:  fields{Job: Job{State: Stopped}, stop: called},
			wantErr: utils.ErrorTextContains(t, string(Stopped)),
		},
		{
			name:    "job failed",
			fields:  fields{Job: Job{State: Failed}, stop: called},
			wantErr: utils.ErrorTextContains(t, string(Failed)),
		},
		{
			name:    "unstoppable",
			fields:  fields{Job: Job{State: Running}, stop: nilfn},
			wantErr: utils.ErrorTextContains(t, "cannot stop"),
		},
		{
			name:      "success",
			fields:    fields{Job: Job{State: Running}, stop: called},
			wantCount: 1,
			wantErr:   utils.NoError(t),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0

			e := &execContext{
				Job:  tt.fields.Job,
				stop: tt.fields.stop(&callCount),
				m:    sync.Mutex{},
				buf:  safe.NewNotifyingBuffer(),
			}
			if err := e.interrupt(); !tt.wantErr(err) {
				t.Errorf("execContext.interrupt() error = %v", err)
			}
			if callCount != tt.wantCount {
				t.Errorf("execContext.interrupt() called stop %v times, expected %v", callCount, tt.wantCount)
			}
		})
	}
}

func Test_execContext_exit(t *testing.T) {
	tests := []struct {
		name     string
		job      Job
		exitCode int
		wantJob  Job
		stopped  bool
	}{
		{
			name:     "success",
			exitCode: 0,
			job:      Job{},
			wantJob:  Job{State: Completed},
		},
		{
			name:     "fail",
			exitCode: 255,
			job:      Job{},
			wantJob:  Job{State: Failed, ExitCode: 255},
		},
		{
			name:     "interrupted",
			exitCode: -1,
			stopped:  true,
			job:      Job{},
			wantJob:  Job{State: Stopped, ExitCode: -1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &execContext{
				Job:  tt.job,
				stop: nil,
				m:    sync.Mutex{},
				buf:  safe.NewNotifyingBuffer(),
			}
			e.stopped.Store(tt.stopped)
			e.exit(tt.exitCode)

			// copy EndTime to prevent valid equality failures
			tt.wantJob.EndTime = e.EndTime

			if !reflect.DeepEqual(e.Job, tt.wantJob) {
				t.Errorf("execContext.exit() = %v, wanted = %v", e.Job, tt.wantJob)
			}
		})
	}
}
