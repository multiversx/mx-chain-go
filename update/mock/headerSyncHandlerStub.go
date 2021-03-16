package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// HeaderSyncHandlerStub -
type HeaderSyncHandlerStub struct {
	SyncUnFinishedMetaHeadersCalled func(epoch uint32) error
	GetEpochStartMetaBlockCalled    func() (data.MetaHeaderHandler, error)
	GetUnFinishedMetaBlocksCalled   func() (map[string]data.MetaHeaderHandler, error)
}

// SyncUnFinishedMetaHeaders -
func (hsh *HeaderSyncHandlerStub) SyncUnFinishedMetaHeaders(epoch uint32) error {
	if hsh.SyncUnFinishedMetaHeadersCalled != nil {
		return hsh.SyncUnFinishedMetaHeadersCalled(epoch)
	}
	return nil
}

// GetEpochStartMetaBlock -
func (hsh *HeaderSyncHandlerStub) GetEpochStartMetaBlock() (data.MetaHeaderHandler, error) {
	if hsh.GetEpochStartMetaBlockCalled != nil {
		return hsh.GetEpochStartMetaBlockCalled()
	}
	return nil, nil
}

// GetUnFinishedMetaBlocks -
func (hsh *HeaderSyncHandlerStub) GetUnFinishedMetaBlocks() (map[string]data.MetaHeaderHandler, error) {
	if hsh.GetUnFinishedMetaBlocksCalled != nil {
		return hsh.GetUnFinishedMetaBlocksCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (hsh *HeaderSyncHandlerStub) IsInterfaceNil() bool {
	return hsh == nil
}
