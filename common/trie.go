package common

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

// TrieNodeVersion defines the type of the trie node version
type TrieNodeVersion uint8

const (
	// NotSpecified means that the value not populated or is not important
	NotSpecified TrieNodeVersion = iota

	// AutoBalanceEnabled is used for data tries, and only after the activation of AutoBalanceDataTriesEnableEpoch flag
	AutoBalanceEnabled
)

// EmptyTrieHash returns the value with empty trie hash
var EmptyTrieHash = make([]byte, 32)

// IsEmptyTrie returns true if the given root is for an empty trie
func IsEmptyTrie(root []byte) bool {
	if len(root) == 0 {
		return true
	}
	if bytes.Equal(root, EmptyTrieHash) {
		return true
	}
	return false
}

// TrimSuffixFromValue returns the value without the suffix
func TrimSuffixFromValue(value []byte, suffixLength int) ([]byte, error) {
	if suffixLength == 0 {
		return value, nil
	}

	dataLength := len(value) - suffixLength
	if dataLength < 0 {
		return nil, core.ErrSuffixNotPresentOrInIncorrectPosition
	}

	return value[:dataLength], nil
}
