package foreman

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/drrev/telehandler/internal/auth"
	"github.com/drrev/telehandler/pkg/work"
	"github.com/google/uuid"
)

func TestService_resolveJob(t *testing.T) {
	ctx := auth.CommonNamesToCtx(context.Background(), []string{"test-user"})
	job := work.NewJob("test-user", "", []string{})

	noError := func(e error) bool { return e == nil }
	errorTextContains := func(str string) func(e error) bool {
		return func(e error) bool {
			if e == nil {
				return len(str) == 0
			}
			return strings.Contains(e.Error(), str)
		}
	}

	type args struct {
		ctx   context.Context
		jobID string
	}
	tests := []struct {
		name    string
		exe     Executor
		args    args
		want    work.Job
		wantErr func(e error) bool
	}{
		{
			name:    "empty ctx",
			exe:     nil,
			args:    args{ctx: context.Background(), jobID: ""},
			wantErr: errorTextContains("missing CommonName"),
		},
		{
			name:    "invalid job ID",
			exe:     nil,
			args:    args{ctx: ctx, jobID: "invalid-id"},
			wantErr: errorTextContains("invalid job id"),
		},
		{
			name:    "missing job",
			exe:     mockExecFn(nil, fmt.Errorf("not found")),
			args:    args{ctx: ctx, jobID: "8dbc07b7-084f-47d1-bc72-081f007f7e7a"},
			wantErr: errorTextContains("no job found for id"),
		},
		{
			name:    "ownership mismatch",
			exe:     mockExecFn(work.NewJob("invalid", "", []string{}), nil),
			args:    args{ctx: ctx, jobID: "8dbc07b7-084f-47d1-bc72-081f007f7e7a"},
			wantErr: errorTextContains("PermissionDenied"),
		},
		{
			name:    "valid access",
			exe:     mockExecFn(job, nil),
			args:    args{ctx: ctx, jobID: "8dbc07b7-084f-47d1-bc72-081f007f7e7a"},
			want:    *job,
			wantErr: noError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{tt.exe}
			got, err := s.resolveJob(tt.args.ctx, tt.args.jobID)
			if !tt.wantErr(err) {
				t.Errorf("Service.resolveJob() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.resolveJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockExecFn(job *work.Job, err error) *mockExec {
	return &mockExec{cb: func(id uuid.UUID) (*work.Job, error) {
		return job, err
	}}
}

type mockExec struct {
	Executor
	cb func(id uuid.UUID) (*work.Job, error)
}

func (m *mockExec) Find(id uuid.UUID) (job work.Job, err error) {
	var j *work.Job

	j, err = m.cb(id)
	if j != nil {
		job = *j
	}

	return
}
