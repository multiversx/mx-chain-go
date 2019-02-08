package mock

// BootstraperMock mocks the implementation for a Bootstraper
type BootstraperMock struct {
	ShouldSyncCalled func() bool
}

// ShouldSync returns true if the node is not synchronized, otherwise it returns false
func (boot *BootstraperMock) ShouldSync() bool {
	return boot.ShouldSyncCalled()
}
