package foreman

import (
	"context"
	"errors"
	"io"
	"log/slog"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/internal/codec"
	"github.com/drrev/telehandler/pkg/safe"
	"github.com/drrev/telehandler/pkg/work"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Executor is the minimal interface needed to manage jobs for Start/Stop/WatchOuput.
type Executor interface {
	Start(j work.Job) (work.Job, error)
	Lookup(name string) (work.Job, error)
	OpenReader(name string) (*safe.NotifyingBufferReader, error)
	Stop(name string) error
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
	job, err := s.exe.Lookup(req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no job found for '%v'", req.GetName())
	}
	return codec.JobToJobStatePb(job), nil
}

// StartJob implements foremanpb.ForemanServiceServer.
func (s *Service) StartJob(ctx context.Context, req *foremanpb.StartJobRequest) (*foremanpb.JobResponse, error) {
	job, err := s.exe.Start(*work.NewJob(req.GetParent(), req.GetCommand(), req.GetArgs()))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to start job", slog.String("cmd", req.GetCommand()), slog.Any("args", req.GetArgs()))
		return nil, status.Error(codes.Internal, "failed to start job")
	}

	return &foremanpb.JobResponse{
		Name:  job.Name,
		State: codec.JobStateToPb(job.State),
	}, nil
}

// StopJob implements foremanpb.ForemanServiceServer.
func (s *Service) StopJob(ctx context.Context, req *foremanpb.StopJobRequest) (*emptypb.Empty, error) {
	if err := s.exe.Stop(req.GetName()); err != nil {
		slog.ErrorContext(ctx, "Failed to stop job", slog.String("name", req.GetName()))
		return nil, status.Errorf(codes.Internal, "failed to stop job: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// WatchJobOutput implements foremanpb.ForemanServiceServer.
func (s *Service) WatchJobOutput(req *foremanpb.WatchJobOutputRequest, srv grpc.ServerStreamingServer[foremanpb.JobOutput]) error {
	ctx := srv.Context()

	job, err := s.exe.Lookup(req.GetName())
	if err != nil {
		return err
	}

	r, err := s.exe.OpenReader(job.Name)
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
