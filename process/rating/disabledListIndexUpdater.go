package rating

// DisabledListIndexUpdater is a nil implementation for an list index updater
type DisabledListIndexUpdater struct {
}

// UpdateListAndIndex will return nil
func (n *DisabledListIndexUpdater) UpdateListAndIndex(_ string, _ uint32, _ string, _ uint32) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (n *DisabledListIndexUpdater) IsInterfaceNil() bool {
	return n == nil
}
