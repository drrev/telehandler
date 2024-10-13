//go:build linux
// +build linux

package work

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/drrev/telehandler/pkg/safe"
)

// makeCommand creates an [exec.Cmd] to execute the given job.
func makeCommand(buf *safe.NotifyingBuffer, cgroot string, job Job) *exec.Cmd {
	return &exec.Cmd{
		Path:   "/proc/self/exe", // reexec--this ONLY works on Linux
		Args:   append([]string{"reexec", "reexec", "--cgroup-root", cgroot, "--", job.Cmd}, job.Args...),
		Stdout: buf,
		Stderr: buf,
		// TODO: create empty ENV to prevent leaking any PI.
		// For now, just leave it, as it breaks PATH and a few others.
		// Env:    []string{},
		SysProcAttr: &syscall.SysProcAttr{
			// TODO: should cmd also apply Pdeathsig and require runtime.LockOSThread()?
			// That seems like potentially a lot of extra threads to handle killing child processes here.
			// It might be okay to just require the group be killed to kill all of the child processes.
			// Pdeathsig:    syscall.SIGTERM,
			Cloneflags:   syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER | syscall.CLONE_NEWUTS,
			Unshareflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
			Credential:   &syscall.Credential{Uid: 0, Gid: 0},
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
		},
	}
}
