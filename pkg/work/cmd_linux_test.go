//go:build linux
// +build linux

package work

import (
	"os/exec"
	"slices"
	"testing"

	"github.com/drrev/telehandler/pkg/safe"
)

func Test_makeCommand(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	buf := safe.NewNotifyingBuffer()
	cmd, cancel := makeCommand(buf, tmp, "bash", "-c", "echo hello, world!")

	if cmd.Path != selfExePath {
		t.Errorf("makeCommand() cmd.Path wanted %v, got %v", selfExePath, cmd.Path)
	}

	args := []string{selfExePath, "reexec", "--cgroup-root", tmp, "--", "bash", "-c", "echo hello, world!"}
	if !slices.Equal(cmd.Args, args) {
		t.Errorf("makeCommand() cmd.Args wanted %v, got %v", args, cmd.Args)
	}

	cmd.Path = "/usr/bin/env"
	cmd.Args = []string{"/usr/bin/env", "sleep", "60"}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// make sure cancel works
	cancel()

	err := cmd.Wait()
	if err == nil {
		t.Error("cmd.Wait() did not receive expected error")
	} else if exitErr, ok := err.(*exec.ExitError); !ok {
		t.Errorf("cmd.Wait() invalid error type wanted  exec.ExitError, got %#v", exitErr)
	} else {
		if exitErr.ExitCode() >= 0 {
			t.Errorf("cmd.Wait() exit_code wanted < 0, got %v", exitErr.ExitCode())
		}
	}
}
