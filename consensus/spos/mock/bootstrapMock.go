package mock

// BootstrapMock mocks the implementation for a Bootstraper
type BootstrapMock struct {
	ShouldSyncCalled func() bool
}

// ShouldSync returns true if the node is not synchronized, otherwise it returns false
func (boot *BootstrapMock) ShouldSync() bool {
	return boot.ShouldSyncCalled()
}
