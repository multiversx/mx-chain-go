package rating

// DisabledListIndexUpdater is a nil implementation for an list index updater
type DisabledListIndexUpdater struct {
}

// UpdateListAndIndex will return nil
func (n *DisabledListIndexUpdater) UpdateListAndIndex(pubKey string, shardID uint32, list string, index int) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *DisabledListIndexUpdater) IsInterfaceNil() bool {
	return n == nil
}
