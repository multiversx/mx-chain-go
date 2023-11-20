package common

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
)

// TrieLeafHolder defines the behaviour of a component that will hold the retrieved
// data for a trie leaf node
type TrieLeafHolder interface {
	Value() []byte
	Depth() uint32
	Version() core.TrieNodeVersion
}

type trieLeafHolder struct {
	value   []byte
	depth   uint32
	version core.TrieNodeVersion
}

// NewTrieLeafHolder creates a new instance of trie leaf storage
func NewTrieLeafHolder(value []byte, depth uint32, version core.TrieNodeVersion) *trieLeafHolder {
	return &trieLeafHolder{
		value:   value,
		depth:   depth,
		version: version,
	}
}

// Value returns the value of the trie leaf
func (t *trieLeafHolder) Value() []byte {
	return t.value
}

// Depth returns the depth of the trie leaf
func (t *trieLeafHolder) Depth() uint32 {
	return t.depth
}

// Version returns the version of the trie leaf
func (t *trieLeafHolder) Version() core.TrieNodeVersion {
	return t.version
}

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
