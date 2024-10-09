package foreman

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/drrev/telehandler/internal/auth"
	"github.com/drrev/telehandler/pkg/work"
	"github.com/google/uuid"
)

func TestService_resolveJob(t *testing.T) {
	ctx := auth.CommonNamesToCtx(context.TODO(), []string{"test-user"})
	job := work.NewJob("test-user", "", []string{})

	type args struct {
		ctx   context.Context
		jobID string
	}
	tests := []struct {
		name    string
		exe     Executor
		args    args
		want    *work.Job
		wantErr bool
	}{
		{
			name:    "empty ctx",
			exe:     nil,
			args:    args{ctx: context.TODO(), jobID: ""},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid job ID",
			exe:     nil,
			args:    args{ctx: ctx, jobID: "invalid-id"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing job",
			exe:     mockExecFn(nil, fmt.Errorf("not found")),
			args:    args{ctx: ctx, jobID: "invalid-id"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ownership mismatch",
			exe:     mockExecFn(work.NewJob("invalid", "", []string{}), nil),
			args:    args{ctx: ctx, jobID: "8dbc07b7-084f-47d1-bc72-081f007f7e7a"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "valid access",
			exe:     mockExecFn(job, nil),
			args:    args{ctx: ctx, jobID: "8dbc07b7-084f-47d1-bc72-081f007f7e7a"},
			want:    job,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{tt.exe}
			got, err := s.resolveJob(tt.args.ctx, tt.args.jobID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.resolveJob() error = %v, wantErr %v", err, tt.wantErr)
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
