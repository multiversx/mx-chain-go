//+build linux darwin

package logger

import (
	"os"
	"syscall"
)

// redirectStderr redirects the output of the stderr to the file passed in
func redirectStderr(f *os.File) error {
	err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		return err
	}

	return nil
}
