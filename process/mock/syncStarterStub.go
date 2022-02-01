package mock

import "context"

// SyncStarterStub -
type SyncStarterStub struct {
	SyncBlockCalled func() error
}

// SyncBlock -
func (sss *SyncStarterStub) SyncBlock(_ context.Context) error {
	return sss.SyncBlockCalled()
}
