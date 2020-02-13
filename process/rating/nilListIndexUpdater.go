package rating

// NilListIndexUpdater is a nil implementation for an list index updater
type NilListIndexUpdater struct {
}

// UpdateListAndIndex will return nil
func (n *NilListIndexUpdater) UpdateListAndIndex(pubKey string, list string, index int) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *NilListIndexUpdater) IsInterfaceNil() bool {
	return n == nil
}
