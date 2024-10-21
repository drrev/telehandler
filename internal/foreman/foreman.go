package foreman

import (
	"context"
	"errors"
	"io"
	"log/slog"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/internal/auth"
	"github.com/drrev/telehandler/internal/codec"
	"github.com/drrev/telehandler/pkg/safe"
	"github.com/drrev/telehandler/pkg/work"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Executor is the minimal interface needed to manage jobs for Start/Stop/WatchOuput.
type Executor interface {
	Start(j work.Job) (work.Job, error)
	Find(id uuid.UUID) (work.Job, error)
	Output(id uuid.UUID) (*safe.NotifyingBufferReader, error)
	Running(jobID uuid.UUID) (v bool, ok bool)
	Stop(id uuid.UUID) error
}

// Service implements [foremanpb.ForemanServiceServer].
type Service struct {
	exe Executor
}

// NewService creates a new [Service] instance that implements [foremanpb.ForemanServiceServer] and
// can be registered with [foremanpb.RegisterForemanServiceServer].
func NewService(exe Executor) *Service {
	return &Service{exe}
}

// GetJobStatus implements foremanpb.ForemanServiceServer.
func (s *Service) GetJobStatus(ctx context.Context, req *foremanpb.GetJobStatusRequest) (*foremanpb.JobStatus, error) {
	job, err := s.resolveJob(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return codec.JobToJobStatePb(job), nil
}

// StartJob implements foremanpb.ForemanServiceServer.
func (s *Service) StartJob(ctx context.Context, req *foremanpb.StartJobRequest) (*foremanpb.JobResponse, error) {
	name, err := auth.CommonNameFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	job, err := s.exe.Start(*work.NewJob(name, req.GetCommand(), req.GetArgs()))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to start job", slog.String("cmd", req.GetCommand()), slog.Any("args", req.GetArgs()))
		return nil, status.Error(codes.Internal, "failed to start job")
	}

	return &foremanpb.JobResponse{
		Id:    job.ID.String(),
		State: codec.JobStateToPb(job.State),
	}, nil
}

// StopJob implements foremanpb.ForemanServiceServer.
func (s *Service) StopJob(ctx context.Context, req *foremanpb.StopJobRequest) (*emptypb.Empty, error) {
	job, err := s.resolveJob(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	if err := s.exe.Stop(job.ID); err != nil {
		slog.ErrorContext(ctx, "Failed to stop job", slog.String("id", job.ID.String()))
		return nil, status.Errorf(codes.Internal, "failed to stop job: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// WatchJobOutput implements foremanpb.ForemanServiceServer.
func (s *Service) WatchJobOutput(req *foremanpb.WatchJobOutputRequest, srv grpc.ServerStreamingServer[foremanpb.JobOutput]) error {
	ctx := srv.Context()

	job, err := s.resolveJob(ctx, req.GetId())
	if err != nil {
		return err
	}

	r, err := s.exe.Output(job.ID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to open job output: %v", err)
	}
	context.AfterFunc(ctx, func() { r.Close() })

	buf := make([]byte, 10240)

	// drain buffer
	for ctx.Err() == nil {
		n, e := r.Read(buf)

		if n > 0 {
			if err := srv.Send(&foremanpb.JobOutput{Data: append([]byte{}, buf[:n]...)}); err != nil {
				return err
			}
		}

		if e != nil {
			if errors.Is(e, io.EOF) {
				return nil
			}
			return status.Errorf(codes.Internal, "failed to stream output: %v", e)
		}
	}

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
func (s *Service) resolveJob(ctx context.Context, jobID string) (work.Job, error) {
	var job work.Job
	name, err := auth.CommonNameFromCtx(ctx)
	if err != nil {
		return job, status.Error(codes.PermissionDenied, err.Error())
	}

	// 1. Resolve the job
	id, err := uuid.Parse(jobID)
	if err != nil {
		return job, status.Errorf(codes.InvalidArgument, "invalid job id '%v'", jobID)
	}

	found, err := s.exe.Find(id)
	if err != nil {
		return job, status.Errorf(codes.NotFound, "no job found for id '%v'", jobID)
	}

	// 2. Validate that the requester is admin or the user that created the job
	if err := auth.ValidateAccess(&found, name); err != nil {
		slog.Info("Denied access", slog.Any("name", name), slog.Any("job", found))
		return job, status.Error(codes.PermissionDenied, "")
	}

	return found, nil
}