package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/process"
)

type disabledInterceptedChunksProcessor struct {
}

// NewDisabledInterceptedChunksProcessor returns a disabled implementation that can not process chunks
func NewDisabledInterceptedChunksProcessor() *disabledInterceptedChunksProcessor {
	return &disabledInterceptedChunksProcessor{}
}

// CheckBatch returns a checked chunk result that signals that no chunk has been received
func (d *disabledInterceptedChunksProcessor) CheckBatch(_ *batch.Batch, _ process.WhiteListHandler) (process.CheckedChunkResult, error) {
	return process.CheckedChunkResult{
		IsChunk:        false,
		HaveAllChunks:  false,
		CompleteBuffer: nil,
	}, nil
}

// Close returns nil
func (d *disabledInterceptedChunksProcessor) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledInterceptedChunksProcessor) IsInterfaceNil() bool {
	return d == nil
}
