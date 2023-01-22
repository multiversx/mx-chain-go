package epochStart

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data"
)

// HeadersByHashSyncerStub --
type HeadersByHashSyncerStub struct {
	SyncMissingHeadersByHashCalled func(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error
	GetHeadersCalled               func() (map[string]data.HeaderHandler, error)
	ClearFieldsCalled              func()
}

// SyncMissingHeadersByHash --
func (hhss *HeadersByHashSyncerStub) SyncMissingHeadersByHash(shardIDs []uint32, headersHashes [][]byte, ctx context.Context) error {
	if hhss.SyncMissingHeadersByHashCalled != nil {
		return hhss.SyncMissingHeadersByHashCalled(shardIDs, headersHashes, ctx)
	}
	return nil
}

// GetHeaders --
func (hhss *HeadersByHashSyncerStub) GetHeaders() (map[string]data.HeaderHandler, error) {
	if hhss.GetHeadersCalled != nil {
		return hhss.GetHeadersCalled()
	}
	return nil, nil
}

// ClearFields --
func (hhss *HeadersByHashSyncerStub) ClearFields() {
	if hhss.ClearFieldsCalled != nil {
		hhss.ClearFieldsCalled()
	}
}

// IsInterfaceNil --
func (hhss *HeadersByHashSyncerStub) IsInterfaceNil() bool {
	return hhss == nil
}
