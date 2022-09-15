package testscommon

// ESDTGlobalSettingsHandlerStub -
type ESDTGlobalSettingsHandlerStub struct {
	IsPausedCalled          func(esdtTokenKey []byte) bool
	IsLimitedTransferCalled func(esdtTokenKey []byte) bool
}

// IsPaused -
func (e *ESDTGlobalSettingsHandlerStub) IsPaused(esdtTokenKey []byte) bool {
	if e.IsPausedCalled != nil {
		return e.IsPausedCalled(esdtTokenKey)
	}
	return false
}

// IsLimitedTransfer -
func (e *ESDTGlobalSettingsHandlerStub) IsLimitedTransfer(esdtTokenKey []byte) bool {
	if e.IsLimitedTransferCalled != nil {
		return e.IsLimitedTransferCalled(esdtTokenKey)
	}
	return false
}

// IsInterfaceNil -
func (e *ESDTGlobalSettingsHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
