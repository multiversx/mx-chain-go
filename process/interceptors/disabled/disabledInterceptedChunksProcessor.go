package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/process"
)

type disabledInterceptedChunksProcessor struct {
}

// NewDisabledInterceptedChunksProcessor returns a disabled implementation that can not process chunks
func NewDisabledInterceptedChunksProcessor() *disabledInterceptedChunksProcessor {
	return &disabledInterceptedChunksProcessor{}
}

// CheckBatch returns a checked chunk result that signals that no chunk has been received
func (d *disabledInterceptedChunksProcessor) CheckBatch(_ *batch.Batch) (process.CheckedChunkResult, error) {
	return process.CheckedChunkResult{
		IsChunk:        false,
		HaveAllChunks:  false,
		CompleteBuffer: nil,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledInterceptedChunksProcessor) IsInterfaceNil() bool {
	return d == nil
}
