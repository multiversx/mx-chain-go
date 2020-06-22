package mock

// ImportStartHandlerStub -
type ImportStartHandlerStub struct {
	SetStartImportCalled            func() error
	ResetStartImportCalled          func() error
	IsAfterExportBeforeImportCalled func() bool
	ShouldStartImportCalled         func() bool
}

// IsAfterExportBeforeImport -
func (ish *ImportStartHandlerStub) IsAfterExportBeforeImport() bool {
	if ish.IsAfterExportBeforeImportCalled != nil {
		return ish.IsAfterExportBeforeImportCalled()
	}

	return false
}

// ShouldStartImport -
func (ish *ImportStartHandlerStub) ShouldStartImport() bool {
	if ish.ShouldStartImportCalled != nil {
		return ish.ShouldStartImportCalled()
	}

	return false
}

// ResetStartImport -
func (ish *ImportStartHandlerStub) ResetStartImport() error {
	if ish.ResetStartImportCalled != nil {
		return ish.ResetStartImportCalled()
	}

	return nil
}

// SetStartImport -
func (ish *ImportStartHandlerStub) SetStartImport() error {
	if ish.SetStartImportCalled != nil {
		return ish.SetStartImportCalled()
	}

	return nil
}

// IsInterfaceNil -
func (ish *ImportStartHandlerStub) IsInterfaceNil() bool {
	return ish == nil
}
