package mock

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/update"
)

var _ update.PendingEpochStartShardHeaderSyncHandler = (*PendingEpochStartShardHeaderStub)(nil)

type PendingEpochStartShardHeaderStub struct {
	SyncEpochStartShardHeaderCalled func(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error
	GetEpochStartHeaderCalled       func() (data.HeaderHandler, []byte, error)
	ClearFieldsCalled               func()
}

// SyncEpochStartShardHeader(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error
func (p *PendingEpochStartShardHeaderStub) SyncEpochStartShardHeader(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error {
	if p.SyncEpochStartShardHeaderCalled == nil {
		return nil
	}

	return p.SyncEpochStartShardHeaderCalled(shardId, epoch, startNonce, ctx)
}

// GetEpochStartHeader() (data.HeaderHandler, []byte, error)
func (p *PendingEpochStartShardHeaderStub) GetEpochStartHeader() (data.HeaderHandler, []byte, error) {
	if p.GetEpochStartHeaderCalled == nil {
		return nil, nil, nil
	}

	return p.GetEpochStartHeaderCalled()
}

// ClearFields()
func (p *PendingEpochStartShardHeaderStub) ClearFields() {
	if p.ClearFieldsCalled == nil {
		return
	}

	p.ClearFieldsCalled()
}

// IsInterfaceNil -
func (p *PendingEpochStartShardHeaderStub) IsInterfaceNil() bool {
	return p == nil
}
