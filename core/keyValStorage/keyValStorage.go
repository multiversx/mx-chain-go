package keyValStorage

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
