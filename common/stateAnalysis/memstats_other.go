//go:build !unix

package stateAnalysis

import "time"

func processCPUTime() time.Duration {
	return 0
}

func residentSetBytes() int64 {
	return 0
}
