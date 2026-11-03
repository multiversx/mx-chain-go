//go:build unix

package snapshotAnalysis

import (
	"syscall"
	"time"
)

// processCPUTime returns the user+system CPU time consumed by the process so far
func processCPUTime() time.Duration {
	var usage syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage)
	if err != nil {
		return 0
	}
	user := time.Duration(usage.Utime.Sec)*time.Second + time.Duration(usage.Utime.Usec)*time.Microsecond
	sys := time.Duration(usage.Stime.Sec)*time.Second + time.Duration(usage.Stime.Usec)*time.Microsecond
	return user + sys
}
