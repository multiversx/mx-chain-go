package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type HeaderSyncHandlerStub struct {
	SyncUnFinishedMetaHeadersCalled func(epoch uint32) error
	GetEpochStartMetaBlockCalled    func() (*block.MetaBlock, error)
	GetUnfinishedMetaBlocksCalled   func() (map[string]*block.MetaBlock, error)
}

func (hsh *HeaderSyncHandlerStub) SyncUnFinishedMetaHeaders(epoch uint32) error {
	if hsh.SyncUnFinishedMetaHeadersCalled != nil {
		return hsh.SyncUnFinishedMetaHeadersCalled(epoch)
	}
	return nil
}

func (hsh *HeaderSyncHandlerStub) GetEpochStartMetaBlock() (*block.MetaBlock, error) {
	if hsh.GetEpochStartMetaBlockCalled != nil {
		return hsh.GetEpochStartMetaBlockCalled()
	}
	return nil, nil
}

func (hsh *HeaderSyncHandlerStub) GetUnfinishedMetaBlocks() (map[string]*block.MetaBlock, error) {
	if hsh.GetUnfinishedMetaBlocksCalled != nil {
		return hsh.GetUnfinishedMetaBlocksCalled()
	}
	return nil, nil
}

func (hsh *HeaderSyncHandlerStub) IsInterfaceNil() bool {
	return hsh == nil
}
