//go:build unix

package stateAnalysis

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// processCPUTime returns the user + system CPU time consumed by the process so far
func processCPUTime() time.Duration {
	var ru syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	if err != nil {
		return 0
	}
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}

// residentSetBytes returns the resident set size of the process (linux only, 0 elsewhere)
func residentSetBytes() int64 {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0
	}
	pages, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return pages * int64(os.Getpagesize())
}
