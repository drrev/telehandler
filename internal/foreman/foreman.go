package foreman

import (
	"context"

	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ foremanpb.ForemanServiceServer = (*Service)(nil)

// Service implements [foremanpb.ForemanServiceServer].
type Service struct {
}

func NewService() *Service {
	return &Service{}
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
