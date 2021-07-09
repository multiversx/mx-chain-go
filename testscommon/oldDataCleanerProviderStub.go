package testscommon

// OldDataCleanerProviderStub -
type OldDataCleanerProviderStub struct {
	ShouldCleanCalled func() bool
}

// ShouldClean -
func (o *OldDataCleanerProviderStub) ShouldClean() bool {
	if o.ShouldCleanCalled != nil {
		return o.ShouldCleanCalled()
	}

	return false
}

// IsInterfaceNil -
func (o *OldDataCleanerProviderStub) IsInterfaceNil() bool {
	return o == nil
}
