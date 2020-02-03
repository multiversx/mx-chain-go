package storageUnit

type nilStorer struct {
}

// NewNilStorer will return a nil storer
func NewNilStorer() *nilStorer {
	return new(nilStorer)
}

// GetFromEpoch will do nothing
func (ns *nilStorer) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	return nil, nil
}

// HasInEpoch will do nothing
func (ns *nilStorer) HasInEpoch(key []byte, epoch uint32) error {
	return nil
}

// SearchFirst will do nothing
func (ns *nilStorer) SearchFirst(key []byte) ([]byte, error) {
	return nil, nil
}

// Put will do nothing
func (ns *nilStorer) Put(key, data []byte) error {
	return nil
}

// Close will do nothing
func (ns *nilStorer) Close() error {
	return nil
}

// Get will do nothing
func (ns *nilStorer) Get(key []byte) ([]byte, error) {
	return nil, nil
}

// Has will do nothing
func (ns *nilStorer) Has(key []byte) error {
	return nil
}

// Remove will do nothing
func (ns *nilStorer) Remove(key []byte) error {
	return nil
}

// ClearCache will do nothing
func (ns *nilStorer) ClearCache() {
}

// DestroyUnit will do nothing
func (ns *nilStorer) DestroyUnit() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ns *nilStorer) IsInterfaceNil() bool {
	return ns == nil
}
