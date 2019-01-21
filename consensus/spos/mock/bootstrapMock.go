package mock

type BootstrapMock struct {
	CheckForkCalled func(uint64) bool
}

func (boot *BootstrapMock) CheckFork(nonce uint64) bool {
	return boot.CheckForkCalled(nonce)
}
