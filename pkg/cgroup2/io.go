package cgroup2

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"
)

const (
	ioRBpsLimit  = 83886080
	ioWBpsLimit  = 41943040
	ioRiopsLimit = 1000
	ioWiopsLimit = 1000

	ioConstraintFileName = "io.max"
)

// ioConstraints uses the given deviceIter to create a Constraint slice, so that
// each block device is limited constrained.
func ioConstraints(deviceIter iter.Seq2[string, error]) ([]Constraint, error) {
	var constraints []Constraint

	if deviceIter == nil {
		return nil, nil
	}

	for mm, err := range deviceIter {
		if err != nil {
			return nil, err
		}

		constraints = append(constraints, Constraint{ioConstraintFileName, ioMajorMinorConstraint(mm)})
	}

	return constraints, nil
}

// ioMajorMinorConstraint converts mm (major:minor) into a formatted limit
// for Constraint.
func ioMajorMinorConstraint(mm string) string {
	return fmt.Sprintf(
		"%s rbps=%d wbps=%d riops=%d wiops=%d",
		mm,
		ioRBpsLimit,
		ioWBpsLimit,
		ioRiopsLimit,
		ioWiopsLimit,
	)
}

// blockDevices reads all block devices from /sys/block and exposes
// the discovered major/minor and any read errors as a [iter.Seq2].
func blockDevices() (iter.Seq2[string, error], error) {
	devFiles, err := filepath.Glob("/sys/block/*/dev")
	if err != nil {
		return nil, fmt.Errorf("failed to glob block device files: %w", err)
	}

	return func(yield func(string, error) bool) {
		for _, fp := range devFiles {
			raw, err := os.ReadFile(fp)
			if !yield(strings.TrimSpace(string(raw)), err) {
				return
			}
		}
	}, nil
}
