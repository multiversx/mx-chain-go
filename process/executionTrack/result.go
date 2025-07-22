package executionTrack

// CleanResult represents the outcome of a clean operation.
type CleanResult string

const (
	// CleanResultOK means the clean operation completed successfully.
	CleanResultOK CleanResult = "ok"

	// CleanResultNotFound means the target of the clean operation was not found.
	CleanResultNotFound CleanResult = "result not found"

	// CleanResultMismatch means the operation returned a different result than expected.
	CleanResultMismatch CleanResult = "different result"
)

// CleanInfo holds the result of a clean operation and the last matching nonce.
type CleanInfo struct {
	CleanResult             CleanResult
	LastMatchingResultNonce uint64
}
