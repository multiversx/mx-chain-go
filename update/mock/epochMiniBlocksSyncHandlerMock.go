package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
)

type EpochStartPendingMiniBlocksSyncHandlerMock struct {
	SyncPendingMiniBlocksFromMetaCalled func(meta *block.MetaBlock, waitTime time.Duration) error
	GetMiniBlocksCalled                 func() (map[string]*block.MiniBlock, error)
}

func (ep *EpochStartPendingMiniBlocksSyncHandlerMock) SyncPendingMiniBlocksFromMeta(meta *block.MetaBlock, waitTime time.Duration) error {
	if ep.SyncPendingMiniBlocksFromMetaCalled != nil {
		return ep.SyncPendingMiniBlocksFromMetaCalled(meta, waitTime)
	}
	return nil
}

func (ep *EpochStartPendingMiniBlocksSyncHandlerMock) GetMiniBlocks() (map[string]*block.MiniBlock, error) {
	if ep.GetMiniBlocksCalled != nil {
		return ep.GetMiniBlocksCalled()
	}
	return nil, nil
}

func (ep *EpochStartPendingMiniBlocksSyncHandlerMock) IsInterfaceNil() bool {
	return ep == nil
}
