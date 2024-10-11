package cgroup2

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
)

const (
	ioRBpsLimit  = 83886080
	ioWBpsLimit  = 41943040
	ioRiopsLimit = 1000
	ioWiopsLimit = 1000
)

func ioConstraints(deviceIter iter.Seq2[string, error]) ([]Constraint, error) {
	constraints := []Constraint{}

	for mm, err := range deviceIter {
		if err != nil {
			return nil, err
		}

		v := fmt.Sprintf(
			"%s rbps=%d wbps=%d riops=%d wiops=%d",
			mm,
			ioRBpsLimit,
			ioWBpsLimit,
			ioRiopsLimit,
			ioWiopsLimit,
		)

		constraints = append(constraints, Constraint{"io.max", v})
	}

	return constraints, nil
}

func blockDevices() (iter.Seq2[string, error], error) {
	devFiles, err := filepath.Glob("/sys/block/*/dev")
	if err != nil {
		return nil, fmt.Errorf("failed to glob block device files: %w", err)
	}

	return func(yield func(string, error) bool) {
		for _, fp := range devFiles {
			raw, err := os.ReadFile(fp)
			if !yield(string(raw), err) {
				return
			}
		}
	}, nil
}
