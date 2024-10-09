package foreman

import (
	"context"
	"log/slog"
	"time"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/internal/auth"
	"github.com/drrev/telehandler/internal/codec"
	"github.com/drrev/telehandler/internal/safe"
	"github.com/drrev/telehandler/pkg/work"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JobStore is the minimal interface needed by Service to handle Job persistence.
type JobStore interface {
	Find(id uuid.UUID) (*work.Job, error)
	Save(j *work.Job) error
}

// Executor is the minimal interface needed to manage jobs for Start/Stop/WatchOuput.
type Executor interface {
	Start(j *work.Job) error
	Output(id uuid.UUID) (*safe.Buffer, error)
	Stop(id uuid.UUID) error
}

var _ foremanpb.ForemanServiceServer = (*Service)(nil)

// Service implements [foremanpb.ForemanServiceServer].
type Service struct {
	store JobStore
	exe   Executor
}

// NewService creates a new [Service] instance that implements [foremanpb.ForemanServiceServer] and
// can be registered with [foremanpb.RegisterForemanServiceServer].
func NewService(store JobStore, exe Executor) *Service {
	return &Service{store, exe}
}

// GetJobStatus implements foremanpb.ForemanServiceServer.
func (s *Service) GetJobStatus(ctx context.Context, req *foremanpb.GetJobStatusRequest) (*foremanpb.JobStatus, error) {
	job, err := s.resolveJob(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &foremanpb.JobStatus{
		Id:        job.ID.String(),
		State:     codec.JobStateToPb(job.LoadState()),
		StartTime: timestamppb.New(job.LoadStartTime()),
		EndTime:   timestamppb.New(job.LoadEndTime()),
		ExitCode:  int32(job.LoadExitCode()),
	}, nil
}

// StartJob implements foremanpb.ForemanServiceServer.
func (s *Service) StartJob(ctx context.Context, req *foremanpb.StartJobRequest) (*foremanpb.JobResponse, error) {
	names, err := auth.CommonNamesFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// if there are multiple names, just pick the first
	// TODO: which name should this pick? without a resource specifier or indicator in the message itself, there's no way to know
	name := names[0]

	job := work.NewJob(name, req.GetCommand(), req.GetArgs())

	if err := s.store.Save(job); err != nil {
		return nil, status.Error(codes.Internal, "failed to save job")
	}

	if err := s.exe.Start(job); err != nil {
		slog.ErrorContext(ctx, "Failed to start job", slog.String("cmd", req.GetCommand()), slog.Any("args", req.GetArgs()))
		return nil, status.Error(codes.Internal, "failed to start job")
	}

	return &foremanpb.JobResponse{
		Id:    job.ID.String(),
		State: codec.JobStateToPb(job.LoadState()),
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
		return nil, status.Error(codes.Internal, "")
	}

	return &emptypb.Empty{}, nil
}

// WatchJobOutput implements foremanpb.ForemanServiceServer.
func (s *Service) WatchJobOutput(req *foremanpb.WatchJobOutputRequest, srv grpc.ServerStreamingServer[foremanpb.JobOutput]) error {
	job, err := s.resolveJob(srv.Context(), req.GetId())
	if err != nil {
		return err
	}

	out, err := s.exe.Output(job.ID)
	if err != nil {
		return err
	}

	// stream output to the client
	ctx := srv.Context()
	off := int64(0)
	seq := int64(0)
	buf := make([]byte, 4096)

	// TODO: Move to a dedicated OutputReader
	// drain buffer
	for {
		// wait a max of 5 seconds for changes
		tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		seq = out.Wait(tctx, seq)
		cancel()

		bl := int64(out.Len())
		if off >= bl {
			// we have reached the end of available data
			// no more data is coming if the job is not running, bail
			if !job.Running() {
				return nil
			}

			// send empty to make sure the client is waiting
			err := srv.Send(&foremanpb.JobOutput{})
			if err != nil {
				return err
			}
		}

		// send any available data
		for off < bl {
			n, _ := out.ReadAt(buf, off)
			off += int64(n)

			err := srv.Send(&foremanpb.JobOutput{Data: buf[:n]})
			if err != nil {
				return err
			}
		}

		// avoid checking job state here in case state changed, but bl increased
		// this will cause an extra 5s wait, but that is acceptable
	}
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
