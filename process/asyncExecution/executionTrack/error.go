package executionTrack

import "errors"

var (
	// ErrNilExecutionResult signals that a nil execution results has been provided
	ErrNilExecutionResult = errors.New("nil execution result")

	// ErrDifferentNoncesConfirmedExecutionResults signals that the confirmed execution results have different nonces
	ErrDifferentNoncesConfirmedExecutionResults = errors.New("confirmed execution results have different nonces")

	// ErrCannotFindExecutionResult signals that execution result is not found
	ErrCannotFindExecutionResult = errors.New("cannot find execution result")

	// ErrWrongExecutionResultNonce signals that the nonce of execution result is wrong
	ErrWrongExecutionResultNonce = errors.New("wrong execution result nonce")

	// ErrNilLastNotarizedExecutionResult signals that last notarized execution result is nil
	ErrNilLastNotarizedExecutionResult = errors.New("nil last notarized execution result")
)
