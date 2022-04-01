package common

// StorageAccessType defines the storage strategy in order to provide better prioritization levels
type StorageAccessType string

const (
	// LowPriority defines the low storage priority value
	LowPriority StorageAccessType = "low"

	// HighPriority defines the high storage priority value
	HighPriority StorageAccessType = "high"
)

const (
	// ProcessPriority is the time critical priority, usually used when processing blocks
	ProcessPriority = HighPriority

	// APIPriority is the priority used when serving API calls
	APIPriority = HighPriority

	// ResolveRequestPriority is the priority used when serving requests
	ResolveRequestPriority = LowPriority

	// SnapshotPriority is the priority used when doing a trie snapshot
	SnapshotPriority = LowPriority

	// TriePruningPriority is the priority used when doing a trie pruning operation
	TriePruningPriority = LowPriority

	// HeartbeatPriority is the priority used for the heartbeat operation
	// TODO remove this after heartbeat v2 is in place
	HeartbeatPriority = LowPriority

	// TestPriority is the priority used when executing test code
	TestPriority = LowPriority
)

// IsStorageAccessValid returns true if the provided priority is valid
func IsStorageAccessValid(priority StorageAccessType) bool {
	return priority == LowPriority || priority == HighPriority
}
