//+build darwin

package redirects

import (
	"os"
	"syscall"
)

// RedirectStderr redirects the output of the stderr to the file passed in
func RedirectStderr(f *os.File) error {
	err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		return err
	}

	return nil
}
