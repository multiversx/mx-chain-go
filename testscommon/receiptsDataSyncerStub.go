package testscommon

import "context"

type ReceiptsDataSyncerStub struct {
	GetReceiptsDataCalled     func() (map[string][]byte, error)
	SyncReceiptsDataForCalled func(hashes [][]byte, ctx context.Context) error
}

// SyncReceiptsDataFor -
func (r *ReceiptsDataSyncerStub) SyncReceiptsDataFor(hashes [][]byte, ctx context.Context) error {
	if r.SyncReceiptsDataForCalled != nil {
		return r.SyncReceiptsDataForCalled(hashes, ctx)
	}
	return nil
}

// GetReceiptsData -
func (r *ReceiptsDataSyncerStub) GetReceiptsData() (map[string][]byte, error) {
	if r.GetReceiptsDataCalled != nil {
		return r.GetReceiptsDataCalled()
	}
	return nil, nil
}

// ClearFields -
func (r *ReceiptsDataSyncerStub) ClearFields() {

}

// IsInterfaceNil -
func (r *ReceiptsDataSyncerStub) IsInterfaceNil() bool {
	return r == nil
}
