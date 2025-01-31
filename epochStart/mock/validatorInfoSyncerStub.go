package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state"
)

// ValidatorInfoSyncerStub -
type ValidatorInfoSyncerStub struct {
	SyncMiniBlocksCalled     func(hdr data.HeaderHandler) ([][]byte, data.BodyHandler, error)
	SyncValidatorsInfoCalled func(body data.BodyHandler) ([][]byte, map[string]*state.ShardValidatorInfo, error)
}

// SyncMiniBlocks -
func (vip *ValidatorInfoSyncerStub) SyncMiniBlocks(hdr data.HeaderHandler) ([][]byte, data.BodyHandler, error) {
	if vip.SyncMiniBlocksCalled != nil {
		return vip.SyncMiniBlocksCalled(hdr)
	}

	return nil, nil, nil
}

// SyncValidatorsInfo -
func (vip *ValidatorInfoSyncerStub) SyncValidatorsInfo(body data.BodyHandler) ([][]byte, map[string]*state.ShardValidatorInfo, error) {
	if vip.SyncValidatorsInfoCalled != nil {
		return vip.SyncValidatorsInfoCalled(body)
	}

	return nil, nil, nil
}

// IsInterfaceNil -
func (vip *ValidatorInfoSyncerStub) IsInterfaceNil() bool {
	return vip == nil
}
