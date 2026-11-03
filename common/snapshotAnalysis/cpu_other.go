//go:build !unix

package snapshotAnalysis

import "time"

// processCPUTime is not available on this platform
func processCPUTime() time.Duration {
	return 0
}
