package mock

type BootstrapMock struct {
	ShouldSyncCalled func() bool
}

func (boot *BootstrapMock) ShouldSync() bool {
	return boot.ShouldSyncCalled()
}
