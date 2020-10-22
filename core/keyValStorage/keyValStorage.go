package keyValStorage

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
)

// KeyValStorage holds a key and an associated value
type keyValStorage struct {
	key   []byte
	value []byte
}

// NewKeyValStorage creates a new key-value storage
func NewKeyValStorage(key []byte, val []byte) *keyValStorage {
	return &keyValStorage{
		key:   key,
		value: val,
	}
}

// Key returns the key in the key-value storage
func (k *keyValStorage) Key() []byte {
	return k.key
}

// Value returns the value in the key-value storage
func (k *keyValStorage) Value() []byte {
	return k.value
}

// ValueWithoutSuffix will try to trim the suffix from the current data field
func (k *keyValStorage) ValueWithoutSuffix(suffix []byte) ([]byte, error) {
	if len(suffix) == 0 {
		return k.value, nil
	}

	lenValue := len(k.value)
	lenSuffix := len(suffix)
	position := bytes.Index(k.value, suffix)
	if position != lenValue-lenSuffix || position < 0 {
		return nil, core.ErrSuffixNotPresentOrOnIncorrectPosition
	}

	newData := make([]byte, lenValue-lenSuffix)
	copy(newData, k.value[:position])

	return newData, nil
}
