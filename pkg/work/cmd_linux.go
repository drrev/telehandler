//go:build linux
// +build linux

package work

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/drrev/telehandler/pkg/safe"
)

const selfExePath = "/proc/self/exe"

// makeCommand creates an [exec.Cmd] to execute the given job.
func makeCommand(buf *safe.NotifyingBuffer, cgroot string, name string, args ...string) (cmd *exec.Cmd, cancel func()) {
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())

	// inject wrapper args
	cmdargs := append([]string{"reexec", "--cgroup-root", cgroot, "--", name}, args...)

	cmd = exec.CommandContext(ctx, selfExePath, cmdargs...)

	// max wait after Cancel() to send SIGKILL
	cmd.WaitDelay = 5 * time.Second
	cmd.Cancel = func() error {
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			return err
		}
		cmd.Stdout = nil
		cmd.Stderr = nil
		return nil
	}

	// mux IO to buf
	cmd.Stdout = buf
	cmd.Stderr = buf

	// setup Linux specific proc attrs for namespaces, ID mapping, and Pdeathsig
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Force setpgid so we do not accidentally kill the parent (Telehandler) process.
		Setpgid: true,
		// see: https://man7.org/linux/man-pages/man2/pr_set_pdeathsig.2const.html
		Pdeathsig: syscall.SIGTERM,
		// see: https://man7.org/linux/man-pages/man2/unshare.2.html#DESCRIPTION
		Cloneflags:   syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER | syscall.CLONE_NEWUTS,
		Unshareflags: syscall.CLONE_NEWNS,
		// map running UID/GID into root in the new user namespace
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Geteuid(), // use non-root user to run as root
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getegid(), // use non-root group to run as root
				Size:        1,
			},
		},
	}

	return
}
