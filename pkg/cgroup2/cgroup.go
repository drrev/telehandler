package cgroup2

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sys/unix"
)

var (
	requiredControllers = []string{"cpu", "memory", "io"}

	constraints = []Constraint{
		{"cpu.max", "100000 1000000"},
		{"memory.max", "512M"},
		{"memory.high", "384M"},
		{"memory.swap.max", "0"},
	}
)

// Create a new cgroup v2 at the given base path.
// CPU, memory, and IO constraints are automatically added.
func Create(basePath string) (err error) {
	defer func() {
		if err != nil {
			os.Remove(basePath)
		}
	}()

	if err = createGroup(basePath); err != nil {
		return
	}

	if err = validateCgroupv2(basePath); err != nil {
		return
	}

	return applyAllConstraints(basePath)
}

// Cleanup removes the cgroup created at the given basePath.
func Cleanup(basePath string) error {
	return os.Remove(basePath)
}

// applyAllConstraints applies constraints for cpu, memory, and io.
func applyAllConstraints(root string) error {
	blockDeviceIter, err := blockDevices()
	if err != nil {
		return err
	}

	blockConstraints, err := ioConstraints(blockDeviceIter)
	if err != nil {
		return err
	}

	for _, c := range constraints {
		if err := applyConstraint(root, c); err != nil {
			return fmt.Errorf("failed to apply constraint to file %s: %w", c.FileName, err)
		}
	}

	for _, c := range blockConstraints {
		if err := applyConstraint(root, c); err != nil {
			return fmt.Errorf("failed to apply constraint to file %s: %w", c.FileName, err)
		}
	}

	return nil
}

// applyConstraint writes a constraint value to the target based at root.
func applyConstraint(root string, c Constraint) error {
	return os.WriteFile(filepath.Join(root, c.FileName), []byte(c.Value), fs.FileMode(0))
}

// validateCgroupv2 checks to ensure that only cgroup v2 is in use on the host.
func validateCgroupv2(basePath string) error {
	var st unix.Statfs_t
	if err := unix.Statfs(basePath, &st); err != nil {
		return fmt.Errorf("failed to stat cgroup root: %w", err)
	}
	if st.Type != unix.CGROUP2_SUPER_MAGIC {
		return fmt.Errorf("unsupported cgroup configuration, only cgroup v2 is supported")
	}
	return nil
}

// createGroup makes the base directory and ensures all [requiredControllers]
// are available in the group.
func createGroup(path string) (err error) {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create cgroup path: %w", err)
	}

	// check for any missing controllers
	raw, err := os.ReadFile(filepath.Join(path, "cgroup.controllers"))
	if err != nil {
		return fmt.Errorf("failed to open cgroup controllers: %w", err)
	}

	ctrl := strings.Fields(string(raw))
	missing := make([]string, 0, len(requiredControllers))
	for _, c := range requiredControllers {
		if !slices.Contains(ctrl, c) {
			missing = append(missing, c)
		}
	}

	// This code DOES NOT enable subgroup_controllers in the parent.
	// The user is responsible for adding any missing controllers to the parent.
	if len(missing) > 0 {
		return fmt.Errorf("missing required controllers %v in %v", missing, ctrl)
	}

	return nil
}

// Constraint is a simple file/value format for
// writing a formatted string into a cgroup file
// to apply limits.
type Constraint struct {
	// FileName is the cgroup v2 relative file name.
	FileName string
	// Value is a preformatted constraint to write to FileName.
	Value string
}
