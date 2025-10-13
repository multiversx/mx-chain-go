package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

// ChunkProcessorStub -
type ChunkProcessorStub struct {
	CheckBatchCalled   func(b *batch.Batch, w process.WhiteListHandler, broadcastMethod p2p.BroadcastMethod) (process.CheckedChunkResult, error)
	MarkVerifiedCalled func(b *batch.Batch, broadcastMethod p2p.BroadcastMethod)
	CloseCalled        func() error
}

// CheckBatch -
func (c *ChunkProcessorStub) CheckBatch(b *batch.Batch, w process.WhiteListHandler, broadcastMethod p2p.BroadcastMethod) (process.CheckedChunkResult, error) {
	if c.CheckBatchCalled != nil {
		return c.CheckBatchCalled(b, w, broadcastMethod)
	}

	return process.CheckedChunkResult{}, nil
}

// MarkVerified -
func (c *ChunkProcessorStub) MarkVerified(b *batch.Batch, broadcastMethod p2p.BroadcastMethod) {
	if c.MarkVerifiedCalled != nil {
		c.MarkVerifiedCalled(b, broadcastMethod)
	}
}

// Close -
func (c *ChunkProcessorStub) Close() error {
	if c.CloseCalled != nil {
		return c.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (c *ChunkProcessorStub) IsInterfaceNil() bool {
	return c == nil
}
