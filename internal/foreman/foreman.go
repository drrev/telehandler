package foreman

import (
	"context"
	"fmt"
	"log/slog"

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
)

// Executor is the minimal interface needed to manage jobs for Start/Stop/WatchOuput.
type Executor interface {
	Start(j work.Job) (work.Job, error)
	Find(id uuid.UUID) (work.Job, error)
	Output(id uuid.UUID) (*safe.Buffer, error)
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
	names, err := auth.CommonNamesFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// if there are multiple names, just pick the first
	// TODO: which name should this pick? without a resource specifier or indicator in the message itself, there's no way to know
	name := names[0]

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
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to stop job: %w", err).Error())
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

	out, err := s.exe.Output(job.ID)
	if err != nil {
		return err
	}

	// stream output to the client
	off := int64(0)
	seq := int64(0)
	buf := make([]byte, 4096)

	// TODO: Move to a dedicated OutputReader
	// drain buffer
	for {
		// wait for changes
		seq = out.Wait(ctx, seq)
		running, ok := s.exe.Running(job.ID)
		if !ok {
			return nil
		}

		bl := int64(out.Len())

		if !running && off >= bl {
			// we have reached the end of available data
			// no more data is coming if the job is not running, bail
			return nil
		}

		// send any available data
		for off < bl {
			n, rerr := out.ReadAt(buf, off)
			off += int64(n)

			// TODO: determine if it is worth using a resource pool to prevent unnecessary allocation here
			err := srv.Send(&foremanpb.JobOutput{Data: append([]byte{}, buf[:n]...)})
			if err != nil {
				return err
			}

			if rerr != nil {
				return nil
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
func (s *Service) resolveJob(ctx context.Context, jobID string) (job work.Job, err error) {
	names, err := auth.CommonNamesFromCtx(ctx)
	if err != nil {
		err = status.Error(codes.PermissionDenied, err.Error())
		return
	}

	// 1. Resolve the job
	id, e := uuid.Parse(jobID)
	if e != nil {
		err = status.Errorf(codes.InvalidArgument, "invalid job id '%v'", jobID)
		return
	}

	found, e := s.exe.Find(id)
	if e != nil {
		err = status.Errorf(codes.NotFound, "no job found for id '%v'", jobID)
		return
	}

	// 2. Validate that the requester is admin or the user that created the job
	if e := auth.ValidateAccess(&found, names); e != nil {
		slog.Info("Denied access", slog.Any("names", names), slog.Any("job", found))
		err = status.Error(codes.PermissionDenied, "")
		return
	}

	return found, nil
}
