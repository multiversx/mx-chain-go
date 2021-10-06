package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
)

// EpochStartTriesSyncHandlerMock -
type EpochStartTriesSyncHandlerMock struct {
	SyncTriesFromCalled func(meta *block.MetaBlock, ownShardId uint32) error
}

// SyncTriesFrom -
func (es *EpochStartTriesSyncHandlerMock) SyncTriesFrom(meta *block.MetaBlock, ownShardId uint32) error {
	if es.SyncTriesFromCalled != nil {
		return es.SyncTriesFromCalled(meta, ownShardId)
	}
	return nil
}

// IsInterfaceNil -
func (es *EpochStartTriesSyncHandlerMock) IsInterfaceNil() bool {
	return es == nil
}
