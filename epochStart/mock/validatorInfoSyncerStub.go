package mock

import (
	data "github.com/ElrondNetwork/elrond-go/data"
)

// ValidatorInfoSyncerStub -
type ValidatorInfoSyncerStub struct {
}

// SyncMiniBlocks -
func (vip *ValidatorInfoSyncerStub) SyncMiniBlocks(_ data.HeaderHandler) ([][]byte, data.BodyHandler, error) {
	return nil, nil, nil
}

// IsInterfaceNil -
func (vip *ValidatorInfoSyncerStub) IsInterfaceNil() bool {
	return vip == nil
}
