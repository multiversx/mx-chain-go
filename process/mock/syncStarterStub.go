package mock

// SyncStarterStub -
type SyncStarterStub struct {
	SyncBlockCalled func() error
}

// SyncBlock -
func (sss *SyncStarterStub) SyncBlock() error {
	return sss.SyncBlockCalled()
}
