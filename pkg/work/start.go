package work

import (
	"os/exec"
	"runtime"
)

// Command not found
// See: https://www.gnu.org/software/bash/manual/html_node/Exit-Status.html
const CannotExecute = 127

// startCmd starts the child process and watches for termination.
//
// The OS thread is locked as required by Pdeathsig, since the
// parent of the process with exec.Cmd is the thread. This
// prevents the thread from terminating early, which would
// kill the child process.
func startCmd(c *exec.Cmd, done func(exitCode int)) error {
	errCh := make(chan error)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if err := c.Start(); err != nil {
			errCh <- err
			done(CannotExecute)
			return
		}
		close(errCh)

		code := 0
		if err := c.Wait(); err != nil {
			// Use out of range status in case we cannot
			// determine the exit status.
			//
			// https://tldp.org/LDP/abs/html/exitcodes.html
			code = 255
			if exitErr, ok := err.(*exec.ExitError); ok {
				code = exitErr.ExitCode()
			}
		}
		done(code)
	}()

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}
