package cgroup2

import (
	"fmt"
	"iter"
	"reflect"
	"testing"
)

func Test_ioConstraints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		deviceIter iter.Seq2[string, error]
		want       []Constraint
		wantErr    bool
	}{
		{name: "empty"},
		{
			name: "error in iter",
			deviceIter: func(yield func(string, error) bool) {
				yield("major:minor", fmt.Errorf("testing"))
			},
			wantErr: true,
		},
		{
			name: "single block device",
			deviceIter: func(yield func(string, error) bool) {
				yield("major:minor", nil)
			},
			want: []Constraint{{
				FileName: "io.max",
				Value:    ioMajorMinorConstraint("major:minor"),
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ioConstraints(tt.deviceIter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ioConstraints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ioConstraints() = %v, want %v", got, tt.want)
			}
		})
	}
}
