package core

// KeyValStorage will store a (key, value) pair
type KeyValStorage struct {
	KeyField []byte
	ValField []byte
}

// Key returns the containing key field
func (kvs *KeyValStorage) Key() []byte {
	return kvs.KeyField
}

// Val returns the containing value field
func (kvs *KeyValStorage) Val() []byte {
	return kvs.ValField
}
