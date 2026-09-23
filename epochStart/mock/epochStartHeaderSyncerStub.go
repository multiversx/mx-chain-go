package mock

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/epochStart"
)

var _ epochStart.PendingEpochStartShardHeaderSyncer = (*PendingEpochStartShardHeaderStub)(nil)

// PendingEpochStartShardHeaderStub -
type PendingEpochStartShardHeaderStub struct {
	SyncEpochStartShardHeaderCalled func(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error
	GetEpochStartHeaderCalled       func() (data.HeaderHandler, []byte, error)
	ClearFieldsCalled               func()
}

// SyncEpochStartShardHeader -
func (p *PendingEpochStartShardHeaderStub) SyncEpochStartShardHeader(shardId uint32, epoch uint32, startNonce uint64, ctx context.Context) error {
	if p.SyncEpochStartShardHeaderCalled == nil {
		return nil
	}

	return p.SyncEpochStartShardHeaderCalled(shardId, epoch, startNonce, ctx)
}

// GetEpochStartHeader -
func (p *PendingEpochStartShardHeaderStub) GetEpochStartHeader() (data.HeaderHandler, []byte, error) {
	if p.GetEpochStartHeaderCalled == nil {
		return nil, nil, nil
	}

	return p.GetEpochStartHeaderCalled()
}

// ClearFields -
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
