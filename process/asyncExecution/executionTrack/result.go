package executionTrack

// CleanResult represents the outcome of a clean operation.
type CleanResult int

const (
	// CleanResultOK means the clean operation completed successfully.
	CleanResultOK CleanResult = 1

	// CleanResultNotFound means the target of the clean operation was not found.
	CleanResultNotFound CleanResult = 2

	// CleanResultMismatch means the operation returned a different result than expected.
	CleanResultMismatch CleanResult = 3
)

// CleanInfo holds the result of a clean operation and the last matching nonce.
type CleanInfo struct {
	CleanResult             CleanResult
	LastMatchingResultNonce uint64
}

func (cr CleanResult) String() string {
	switch cr {
	case CleanResultOK:
		return "ok"
	case CleanResultNotFound:
		return "not found"
	case CleanResultMismatch:
		return "mismatch"
	default:
		return "unknown"
	}
}
