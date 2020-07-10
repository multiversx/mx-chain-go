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

// GetKey returns the key in the key-value storage
func (k *keyValStorage) GetKey() []byte {
	return k.key
}

// GetValue returns the value in the key-value storage
func (k *keyValStorage) GetValue() []byte {
	return k.value
}
