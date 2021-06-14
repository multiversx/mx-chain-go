package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ChunkProcessorStub -
type ChunkProcessorStub struct {
	CheckBatchCalled func(b *batch.Batch) (process.CheckedChunkResult, error)
}

// CheckBatch -
func (c *ChunkProcessorStub) CheckBatch(b *batch.Batch) (process.CheckedChunkResult, error) {
	if c.CheckBatchCalled != nil {
		return c.CheckBatchCalled(b)
	}

	return process.CheckedChunkResult{}, nil
}

// IsInterfaceNil -
func (c *ChunkProcessorStub) IsInterfaceNil() bool {
	return c == nil
}
