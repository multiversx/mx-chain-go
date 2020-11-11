package factory

type disabledFloodPreventer struct {
}

// IncreaseLoadGlobal won't do anything and will return nil
func (d *disabledFloodPreventer) IncreaseLoadGlobal(_ string, _ uint64) error {
	return nil
}

// IncreaseLoad won't do anything and will return nil
func (d *disabledFloodPreventer) IncreaseLoad(_ string, _ uint64) error {
	return nil
}

// Reset won't do anything
func (d *disabledFloodPreventer) Reset() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledFloodPreventer) IsInterfaceNil() bool {
	return d == nil
}
