package work

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/drrev/telehandler/pkg/cgroup2"
)

// Reexec is used to run a Linux command in a subprocess wrapper. This must not be called
// by anything other than the reexec command.
func Reexec(ctx context.Context, cgroupRoot string, args []string) (err error) {
	// Lock the OS thread to ensure that the currently executing thread does not die prematurely before this function returns.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := setupRuntime(cgroupRoot); err != nil {
		return fmt.Errorf("setup runtime failed: %w", err)
	}
	defer teardownRuntime(cgroupRoot)
	go func() {
		<-ctx.Done()
		teardownRuntime(cgroupRoot)
	}()

	// lookup command if we weren't given a full path
	path := args[0]
	if !filepath.IsAbs(args[0]) {
		path, err = exec.LookPath(args[0])
		if err != nil {
			return err
		}
	}

	// get the file descriptor for cgroupRoot
	// then set it in SysProcAttr to take advantage
	// of CLONE_INTO_CGROUP.
	// see: https://man.archlinux.org/man/core/man-pages/clone.2.en#CLONE_INTO_CGROUP
	fp, err := os.Open(cgroupRoot)
	if err != nil {
		return fmt.Errorf("failed to open cgroup")
	}
	defer fp.Close()

	// Rebind all, ensure Pdeathsig to kill child on parent death.
	cmd := &exec.Cmd{
		Path:   path,
		Args:   args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Env:    os.Environ(),
		SysProcAttr: &syscall.SysProcAttr{
			Pdeathsig:   syscall.SIGTERM,
			UseCgroupFD: true,
			CgroupFD:    int(fp.Fd()),
		},
	}

	return cmd.Run()
}

// setupRuntime is a convenience function to setup
// cgroups and perform any other setup BEFORE the child
// process is spawned.
func setupRuntime(cgroupRoot string) (err error) {
	defer func() {
		if err != nil {
			// eagerly teardown if any setup failed
			teardownRuntime(cgroupRoot)
		}
	}()

	if err := cgroup2.Create(cgroupRoot); err != nil {
		return fmt.Errorf("failed to create cgroup: %w", err)
	}

	if err := syscall.Sethostname([]byte("sandbox")); err != nil {
		return fmt.Errorf("failed to set hostname: %w", err)
	}

	// Since we are not moving the rootfs, we MUST mount recursively (MS_REC) /proc
	// as private (MS_PRIVATE) before replacing so we do not interfere with the host /proc.
	if err := syscall.Mount("/proc", "/proc", "proc", uintptr(syscall.MS_PRIVATE|syscall.MS_REC), ""); err != nil {
		return fmt.Errorf("failed to remount proc: %w", err)
	}

	// Now mount a new empty procfs with typical mount options (NOSUID, NOEXEC, NODEV).
	// This will isolate the child processes from the host /proc.
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOSUID|syscall.MS_NOEXEC|syscall.MS_NODEV), ""); err != nil {
		return fmt.Errorf("failed to mount over /proc: %w", err)
	}

	return
}

// teardownRuntime cleans up from setupRuntime
// AFTER the child process terminates.
//
// IMPORTANT: If the parent receives SIGKILL, or SIGSTOP,
// this function will not be executed, as those
// signals cannot be trapped: https://pkg.go.dev/os/signal#hdr-Types_of_signals.
func teardownRuntime(cgroupRoot string) {
	// errors are intentionally ignored, we do not want to
	// log into the wrapped command's output
	// and we cannot do anything about these errors
	_ = syscall.Unmount("/proc", 0)
	_ = cgroup2.Cleanup(cgroupRoot)
}
