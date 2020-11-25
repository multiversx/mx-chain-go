package mock

// SoftwareVersionCheckerMock -
type SoftwareVersionCheckerMock struct {
}

// StartCheckSoftwareVersion -
func (svcm *SoftwareVersionCheckerMock) StartCheckSoftwareVersion() {
}

// IsInterfaceNil -
func (svcm *SoftwareVersionCheckerMock) IsInterfaceNil() bool {
	return svcm == nil
}

// Close -
func (svcm *SoftwareVersionCheckerMock) Close() error {
	return nil
}
