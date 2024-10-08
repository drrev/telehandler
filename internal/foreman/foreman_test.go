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

	type fields struct {
		store JobStore
	}
	type args struct {
		ctx   context.Context
		jobID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *work.Job
		wantErr bool
	}{
		{
			name:    "empty ctx",
			fields:  fields{store: nil},
			args:    args{ctx: context.TODO(), jobID: ""},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid job ID",
			fields:  fields{store: nil},
			args:    args{ctx: ctx, jobID: "invalid-id"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "missing job",
			fields:  fields{store: mockJobStore(nil, fmt.Errorf("not found"))},
			args:    args{ctx: ctx, jobID: "invalid-id"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ownership mismatch",
			fields:  fields{store: mockJobStore(work.NewJob("invalid", "", []string{}), nil)},
			args:    args{ctx: ctx, jobID: "8dbc07b7-084f-47d1-bc72-081f007f7e7a"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "valid access",
			fields:  fields{store: mockJobStore(job, nil)},
			args:    args{ctx: ctx, jobID: "8dbc07b7-084f-47d1-bc72-081f007f7e7a"},
			want:    job,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: tt.fields.store,
			}
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

func mockJobStore(job *work.Job, err error) *mockStore {
	return &mockStore{cb: func(id uuid.UUID) (*work.Job, error) {
		return job, err
	}}
}

type mockStore struct {
	JobStore
	cb func(id uuid.UUID) (*work.Job, error)
}

func (m *mockStore) Find(id uuid.UUID) (*work.Job, error) {
	return m.cb(id)
}
