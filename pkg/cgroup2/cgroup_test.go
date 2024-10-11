package cgroup2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_createGroup(t *testing.T) {
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
}
