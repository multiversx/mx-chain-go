package mock

// MockDB -
type MockDB struct {
}

// Put -
func (MockDB) Put(_, _ []byte) error {
	return nil
}

// Get -
func (MockDB) Get(_ []byte) ([]byte, error) {
	return []byte{}, nil
}

// Has -
func (MockDB) Has(_ []byte) error {
	return nil
}

// Init -
func (MockDB) Init() error {
	return nil
}

// Close -
func (MockDB) Close() error {
	return nil
}

// Remove -
func (MockDB) Remove(_ []byte) error {
	return nil
}

// Destroy -
func (MockDB) Destroy() error {
	return nil
}

// DestroyClosed -
func (MockDB) DestroyClosed() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s MockDB) IsInterfaceNil() bool {
	return false
}
