package keyBuilder

import (
	"github.com/multiversx/mx-chain-go/common"
)

const (
	// NibbleSize marks the size of a byte nibble
	NibbleSize = 4

	keyLength = 32
)

type keyBuilder struct {
	key []byte
}

// NewKeyBuilder creates a new key builder. This is used for building trie keys when traversing the trie.
// Use this only if you traverse the trie from the root, else hexToTrieKeyBytes might fail
func NewKeyBuilder() *keyBuilder {
	return &keyBuilder{
		key: make([]byte, 0, keyLength),
	}
}

// BuildKey appends the given byte array to the existing key
func (kb *keyBuilder) BuildKey(keyPart []byte) {
	kb.key = append(kb.key, keyPart...)
}

// GetKey transforms the key from hex to trie key, and returns it.
// Is mandatory that GetKey always returns a new byte slice, not a pointer to an existing one
func (kb *keyBuilder) GetKey() ([]byte, error) {
	return hexToTrieKeyBytes(kb.key)
}

// Clone returns a new KeyBuilder with the same key
func (kb *keyBuilder) Clone() common.KeyBuilder {
	return &keyBuilder{
		key: kb.key,
	}
}

// hexToTrieKeyBytes transforms hex nibbles into key bytes. The hex terminator is removed from the end of the hex slice,
// and then the hex slice is reversed when forming the key bytes.
func hexToTrieKeyBytes(hex []byte) ([]byte, error) {
	hex = hex[:len(hex)-1]
	length := len(hex)
	if length%2 != 0 {
		return nil, ErrInvalidLength
	}

	key := make([]byte, length/2)
	hexSliceIndex := 0
	for i := len(key) - 1; i >= 0; i-- {
		key[i] = hex[hexSliceIndex+1]<<NibbleSize | hex[hexSliceIndex]
		hexSliceIndex += 2
	}

	return key, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (kb *keyBuilder) IsInterfaceNil() bool {
	return kb == nil
}
