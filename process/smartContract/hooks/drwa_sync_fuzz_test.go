package hooks

import (
	"math"
	"testing"
)

func FuzzValidateDRWASyncVersion(f *testing.F) {
	f.Add(uint64(0), uint64(1))
	f.Add(uint64(7), uint64(6))
	f.Add(uint64(7), uint64(7))
	f.Add(uint64(7), uint64(8))
	// Boundary: maximum uint64 — monotonic increment from max must fail (overflow
	// wraps to 0 which is less than current), not silently succeed.
	f.Add(uint64(math.MaxUint64), uint64(0))
	f.Add(uint64(math.MaxUint64-1), uint64(math.MaxUint64))
	f.Add(uint64(math.MaxUint64), uint64(math.MaxUint64))

	f.Fuzz(func(t *testing.T, currentVersion uint64, nextVersion uint64) {
		err := validateDRWASyncVersion(currentVersion, nextVersion)

		switch {
		case nextVersion < currentVersion:
			if err == nil || err.Error() != drwaSyncRejectReplayStale {
				t.Fatalf("expected stale rejection for current=%d next=%d, got %v", currentVersion, nextVersion, err)
			}
		case nextVersion == currentVersion:
			if err == nil || err.Error() != drwaSyncRejectReplayConflict {
				t.Fatalf("expected conflict rejection for current=%d next=%d, got %v", currentVersion, nextVersion, err)
			}
		default:
			if err != nil {
				t.Fatalf("expected success for current=%d next=%d, got %v", currentVersion, nextVersion, err)
			}
		}
	})
}
