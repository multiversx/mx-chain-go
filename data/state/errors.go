package state

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

// ErrorWrongSize is an error-compatible struct holding 2 values: Expected and Got
type ErrorWrongSize struct {
	Exp int
	Got int
}

// ErrorTrieNotNormalized is an error-compatible struct holding the hash length that is not normalized
type ErrorTrieNotNormalized struct {
	actualHashLength   int
	expectedHashLength int
}

// ErrMissingTrie is an error-compatible struct holding the root hash of the trie that is missing
type ErrMissingTrie struct {
	rootHash []byte
}

//------- ErrorWrongSize

// NewErrorWrongSize returns a new instantiated struct
func NewErrorWrongSize(exp int, got int) *ErrorWrongSize {
	return &ErrorWrongSize{Exp: exp, Got: got}
}

// Error returns the error as string
func (e *ErrorWrongSize) Error() string {
	return fmt.Sprintf("wrong size! expected: %d, got %d", e.Exp, e.Got)
}

//------- ErrorTrieNotNormalized

// NewErrorTrieNotNormalized returns a new instantiated struct
func NewErrorTrieNotNormalized(exp int, actual int) *ErrorTrieNotNormalized {
	return &ErrorTrieNotNormalized{expectedHashLength: exp, actualHashLength: actual}
}

// Error returns the error as string
func (e *ErrorTrieNotNormalized) Error() string {
	return "attempt to search a hash not normalized to" +
		strconv.Itoa(e.expectedHashLength) + "bytes (has:" +
		strconv.Itoa(e.actualHashLength) + ")"
}

//------- ErrMissingTrie

// NewErrMissingTrie  returns a new instantiated struct
func NewErrMissingTrie(rootHash []byte) *ErrMissingTrie {
	return &ErrMissingTrie{rootHash: rootHash}
}

// Error returns the error as string
func (e *ErrMissingTrie) Error() string {
	return "trie was not found for hash" + base64.StdEncoding.EncodeToString(e.rootHash)
}

// ErrNilAccountsHandler defines the error when trying to revert on nil accounts
var ErrNilAccountsHandler = errors.New("nil AccountsHandler")

// ErrNilAddress defines the error when trying to work with a nil address
var ErrNilAddress = errors.New("nil Address")

// ErrNilAccountState defines the error when trying to work with a nil account
var ErrNilAccountState = errors.New("nil AccountState")

// ErrNilDataTrie defines the error when trying to search a data key on an uninitialized trie
var ErrNilDataTrie = errors.New("nil data trie for account or data trie not loaded")

// ErrNilTrie signals that a trie is nil and no operation can be made
var ErrNilTrie = errors.New("attempt to search on a nil trie")

// ErrNilValue signals that an operation has been attempted to or with a nil value
var ErrNilValue = errors.New("nil value")
