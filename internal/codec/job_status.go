package codec

import (
	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/pkg/work"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JobToJobStatePb is a convenience function for
// converting a [work.Job] into a [foremanpb.JobStatus].
func JobToJobStatePb(job work.Job) *foremanpb.JobStatus {
	return &foremanpb.JobStatus{
		Id:        job.ID.String(),
		State:     JobStateToPb(job.State),
		StartTime: timestamppb.New(job.StartTime),
		EndTime:   timestamppb.New(job.EndTime),
		ExitCode:  int32(job.ExitCode),
	}
}
