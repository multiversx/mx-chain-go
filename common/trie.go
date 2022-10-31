package common

import "bytes"

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
