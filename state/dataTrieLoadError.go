package state

import "errors"

var ErrDataTrieLoadFailed = errors.New("data trie load failed")

// DataTrieLoadError distinguishes failure to recreate a persisted data trie from
// an absent key or an error encountered later while traversing the trie.
type DataTrieLoadError struct {
	Cause error
}

func (err *DataTrieLoadError) Error() string {
	return ErrDataTrieLoadFailed.Error() + ": " + err.Cause.Error()
}

// Unwrap preserves the missing-node error and its key for state synchronization.
func (err *DataTrieLoadError) Unwrap() error {
	return err.Cause
}

func (err *DataTrieLoadError) Is(target error) bool {
	return target == ErrDataTrieLoadFailed
}
