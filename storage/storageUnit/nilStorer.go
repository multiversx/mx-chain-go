package storageUnit

// NilStorer resembles a disabled implementation of the Storer interface
type NilStorer struct {
}

// NewNilStorer will return a nil storer
func NewNilStorer() *NilStorer {
	return new(NilStorer)
}

// GetFromEpoch will do nothing
func (ns *NilStorer) GetFromEpoch(_ []byte, _ uint32) ([]byte, error) {
	return nil, nil
}

// GetBulkFromEpoch will do nothing
func (ns *NilStorer) GetBulkFromEpoch(_ [][]byte, _ uint32) (map[string][]byte, error) {
	return nil, nil
}

// HasInEpoch will do nothing
func (ns *NilStorer) HasInEpoch(_ []byte, _ uint32) error {
	return nil
}

// SearchFirst will do nothing
func (ns *NilStorer) SearchFirst(_ []byte) ([]byte, error) {
	return nil, nil
}

// Put will do nothing
func (ns *NilStorer) Put(_, _ []byte) error {
	return nil
}

// PutInEpoch will do nothing
func (ns *NilStorer) PutInEpoch(_, _ []byte, _ uint32) error {
	return nil
}

// Close will do nothing
func (ns *NilStorer) Close() error {
	return nil
}

// Get will do nothing
func (ns *NilStorer) Get(_ []byte) ([]byte, error) {
	return nil, nil
}

// Has will do nothing
func (ns *NilStorer) Has(_ []byte) error {
	return nil
}

// Remove will do nothing
func (ns *NilStorer) Remove(_ []byte) error {
	return nil
}

// ClearCache will do nothing
func (ns *NilStorer) ClearCache() {
}

// DestroyUnit will do nothing
func (ns *NilStorer) DestroyUnit() error {
	return nil
}

// RangeKeys does nothing
func (ns *NilStorer) RangeKeys(_ func(key []byte, val []byte) bool) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (ns *NilStorer) IsInterfaceNil() bool {
	return ns == nil
}
