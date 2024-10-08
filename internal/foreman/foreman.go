package foreman

import (
	"context"
	"log/slog"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/internal/auth"
	"github.com/drrev/telehandler/pkg/work"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// JobStore is the minimal interface needed by Service to handle Job persistence.
type JobStore interface {
	Find(id uuid.UUID) (*work.Job, error)
	Save(j *work.Job) error
}

var _ foremanpb.ForemanServiceServer = (*Service)(nil)

// Service implements [foremanpb.ForemanServiceServer].
type Service struct {
	store JobStore
}

func NewService(store JobStore) *Service {
	return &Service{store}
}

// GetJobStatus implements foremanpb.ForemanServiceServer.
func (s *Service) GetJobStatus(ctx context.Context, req *foremanpb.GetJobStatusRequest) (*foremanpb.JobStatus, error) {
	return &foremanpb.JobStatus{}, nil
}

// StartJob implements foremanpb.ForemanServiceServer.
func (s *Service) StartJob(ctx context.Context, req *foremanpb.StartJobRequest) (*foremanpb.JobResponse, error) {
	return &foremanpb.JobResponse{}, nil
}

// StopJob implements foremanpb.ForemanServiceServer.
func (s *Service) StopJob(ctx context.Context, req *foremanpb.StopJobRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// WatchJobOutput implements foremanpb.ForemanServiceServer.
func (s *Service) WatchJobOutput(req *foremanpb.WatchJobOutputRequest, srv grpc.ServerStreamingServer[foremanpb.JobOutput]) error {
	return nil
}

// resolveJob attempts to provide a semi-generic way to authenticate that the requester has
// permission to view the job that is being requested.
//
// Note: the "admin" user can access any job.
//
// Since this version of the API uses raw Job IDs and not resource naming, we have to:
//  1. Resolve the Job from the job store
//  2. Validate that the requester is the user is "admin" or the user that created the job
//
// With resource names, all of this can be handled in an interceptor with protoreflect to get the "name" field,
// then validate that the resource name is scoped to the requester. If it isn't, reject the request.
// Then, each of these handlers would only be responsible for fetching the Job from storage.
func (s *Service) resolveJob(ctx context.Context, jobID string) (*work.Job, error) {
	names, err := auth.CommonNamesFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// 1. Resolve the job
	id, err := uuid.Parse(jobID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid job id '%v'", jobID)
	}

	job, err := s.store.Find(id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no job found for id '%v'", jobID)
	}

	// 2. Validate that the requester is admin or the user that created the job
	if err := auth.ValidateAccess(job, names); err != nil {
		slog.Info("Denied access", slog.Any("names", names), slog.Any("job", job))
		return nil, status.Error(codes.PermissionDenied, "")
	}

	return job, nil
}
