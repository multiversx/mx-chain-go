package mock

type SyncStarterStub struct {
	SyncBlockCalled func() error
}

func (sss *SyncStarterStub) SyncBlock() error {
	return sss.SyncBlockCalled()
}
