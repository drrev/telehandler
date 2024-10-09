package codec

import (
	foremanpb "github.com/drrev/telehandler/gen/drrev/telehandler/foreman/v1alpha1"
	"github.com/drrev/telehandler/pkg/work"
)

// JobStateToPb is a convenience function to convert from
// [work.JobState] to [foremanpb.JobState].
func JobStateToPb(v work.JobState) foremanpb.JobState {
	sv, ok := foremanpb.JobState_value[string(v)]
	if ok {
		return foremanpb.JobState(sv)
	}
	return foremanpb.JobState_JOB_STATE_UNSPECIFIED
}
