package disabled

type disabledCustomDatabaseRemover struct{}

// NewDisabledCustomDatabaseRemover returns a new instance of disabledCustomDatabaseRemover
func NewDisabledCustomDatabaseRemover() *disabledCustomDatabaseRemover {
	return &disabledCustomDatabaseRemover{}
}

// ShouldRemove returns false
func (d *disabledCustomDatabaseRemover) ShouldRemove(_ string, _ uint32) bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledCustomDatabaseRemover) IsInterfaceNil() bool {
	return d == nil
}
