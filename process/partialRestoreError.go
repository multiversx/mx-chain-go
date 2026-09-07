package process

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
)

// MovedMetaBlock is a meta block moved out of committed storage by an interrupted restore
type MovedMetaBlock struct {
	Hash   []byte
	Header data.HeaderHandler
}

// PartialRestoreError reports a failed restore whose moved meta blocks could not be written back
// into committed storage; they need a storage repair before the roll back may be abandoned
type PartialRestoreError struct {
	UnrestoredMetaBlocks []MovedMetaBlock
	cause                error
}

// NewPartialRestoreError wraps the restore failure cause with the meta blocks left unrestored
func NewPartialRestoreError(unrestoredMetaBlocks []MovedMetaBlock, cause error) *PartialRestoreError {
	return &PartialRestoreError{
		UnrestoredMetaBlocks: unrestoredMetaBlocks,
		cause:                cause,
	}
}

// Error returns the cause prefixed with the partial restore context
func (err *PartialRestoreError) Error() string {
	return fmt.Sprintf("partial restore, %d meta block(s) not written back: %v",
		len(err.UnrestoredMetaBlocks), err.cause)
}

// Unwrap returns the restore failure cause
func (err *PartialRestoreError) Unwrap() error {
	return err.cause
}
