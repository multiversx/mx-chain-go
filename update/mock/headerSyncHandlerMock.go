package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
)

type HeaderSyncHandlerMock struct {
	SyncEpochStartMetaHeaderCalled func(epoch uint32, waitTime time.Duration) (*block.MetaBlock, error)
	GetMetaBlockCalled             func() (*block.MetaBlock, error)
}

func (hsh *HeaderSyncHandlerMock) SyncEpochStartMetaHeader(epoch uint32, waitTime time.Duration) (*block.MetaBlock, error) {
	if hsh.SyncEpochStartMetaHeaderCalled != nil {
		return hsh.SyncEpochStartMetaHeaderCalled(epoch, waitTime)
	}
	return nil, nil
}

func (hsh *HeaderSyncHandlerMock) GetMetaBlock() (*block.MetaBlock, error) {
	if hsh.GetMetaBlockCalled != nil {
		return hsh.GetMetaBlockCalled()
	}
	return nil, nil
}

func (hsh *HeaderSyncHandlerMock) IsInterfaceNil() bool {
	return hsh == nil
}
