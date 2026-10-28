package keyBuilder

import (
	"github.com/multiversx/mx-chain-go/common"
)

const (
	// NibbleSize marks the size of a byte nibble
	NibbleSize = 4

	// HexTerminator is the terminator for a trie hex key
	HexTerminator = 16

	keyLength  = 32
	nibbleMask = 0x0f
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

// GetRawKey returns the key as it is, without transforming it
func (kb *keyBuilder) GetRawKey() []byte {
	return kb.key
}

// ShallowClone returns a new KeyBuilder with the same key. The key slice points to the same memory location.
func (kb *keyBuilder) ShallowClone() common.KeyBuilder {
	return &keyBuilder{
		key: kb.key,
	}
}

// DeepClone returns a new KeyBuilder with the same key. This allocates a new memory location for the key slice.
func (kb *keyBuilder) DeepClone() common.KeyBuilder {
	clonedKey := make([]byte, len(kb.key))
	copy(clonedKey, kb.key)

	return &keyBuilder{
		key: clonedKey,
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

// Size returns the size of the key
func (kb *keyBuilder) Size() uint {
	return uint(len(kb.key))
}

// KeyBytesToHex transforms key bytes into hex nibbles. The key nibbles are reversed, meaning that the
// last key nibble will be the first in the hex key. A hex terminator is added at the end of the hex key.
func KeyBytesToHex(str []byte) []byte {
	hexLength := len(str)*2 + 1
	nibbles := make([]byte, hexLength)

	hexSliceIndex := 0
	nibbles[hexLength-1] = HexTerminator

	for i := hexLength - 2; i > 0; i -= 2 {
		nibbles[i] = str[hexSliceIndex] >> NibbleSize
		nibbles[i-1] = str[hexSliceIndex] & nibbleMask
		hexSliceIndex++
	}

	return nibbles
}

// IsInterfaceNil returns true if there is no value under the interface
func (kb *keyBuilder) IsInterfaceNil() bool {
	return kb == nil
}
