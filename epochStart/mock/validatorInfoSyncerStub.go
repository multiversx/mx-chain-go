package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/state"
)

// ValidatorInfoSyncerStub -
type ValidatorInfoSyncerStub struct {
}

// SyncMiniBlocks -
func (vip *ValidatorInfoSyncerStub) SyncMiniBlocks(_ data.HeaderHandler) ([][]byte, data.BodyHandler, error) {
	return nil, nil, nil
}

// SyncValidatorsInfo -
func (vip *ValidatorInfoSyncerStub) SyncValidatorsInfo(_ data.BodyHandler, _ uint32) ([][]byte, map[string]*state.ShardValidatorInfo, error) {
	return nil, nil, nil
}

// IsInterfaceNil -
func (vip *ValidatorInfoSyncerStub) IsInterfaceNil() bool {
	return vip == nil
}
