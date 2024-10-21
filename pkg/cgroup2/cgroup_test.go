package cgroup2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_createGroup(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	// fake controllers
	cgPath := filepath.Join(tmp, "cgroup.controllers")
	err := os.WriteFile(cgPath, []byte(strings.Join(requiredControllers, " ")), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	if err := createGroup(tmp); err != nil {
		t.Errorf("createGroup() error = %v", err)
	}

	os.Remove(cgPath)
	if err := createGroup(tmp); err == nil {
		t.Errorf("createGroup() should error if required controllers are missing")
	}

	if err := createGroup(filepath.Join(tmp, "empty")); err == nil {
		t.Errorf("createGroup() should error if required controllers are missing")
	}
}

func Test_validateCgroupv2(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	if err := validateCgroupv2(tmp); err == nil {
		t.Errorf("validateCgroupv2() did not receive expected error")
	}
}

func Test_applyAllConstraints(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	fp, err := os.OpenFile(filepath.Join(tmp, ioConstraintFileName), os.O_CREATE, 0o644)
	if err != nil {
		t.Fatalf("Failed to initialize test environment: %v", err)
	}
	fp.Close()
	for _, c := range constraints {
		fp, err := os.OpenFile(filepath.Join(tmp, c.FileName), os.O_CREATE, 0o644)
		if err != nil {
			t.Fatalf("Failed to initialize test environment: %v", err)
		}
		fp.Close()
	}

	if err := applyAllConstraints(tmp); err != nil {
		t.Errorf("applyAllConstraints() unexpected error = %v", err)
	}

	// check static constraints
	for _, c := range constraints {
		data, err := os.ReadFile(filepath.Join(tmp, c.FileName))
		if err != nil {
			t.Errorf("failed to read constraint file %s: %v", c.FileName, err)
		}

		if string(data) != c.Value {
			t.Errorf("applyAllConstraints() %s invalid got = %s, expected = %s", c.FileName, data, c.Value)
		}
	}

	// check dynamic constraints
	{
		data, err := os.ReadFile(filepath.Join(tmp, ioConstraintFileName))
		if err != nil {
			t.Errorf("failed to read constraint file %s: %v", ioConstraintFileName, err)
		}

		t.Log(string(data))

		nomm := strings.TrimSpace(ioMajorMinorConstraint(""))
		for _, line := range strings.Split(string(data), "\n") {
			if !strings.HasSuffix(line, nomm) {
				t.Errorf("applyAllConstraints() %s invalid got = %s, expected to have suffix = %s", ioConstraintFileName, line, nomm)
			}
		}
	}
}
