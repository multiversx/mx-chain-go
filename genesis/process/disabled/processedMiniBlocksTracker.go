package disabled

import (
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
)

// ProcessedMiniBlocksTracker implements the ProcessedMiniBlocksTracker interface but does nothing as it is disabled
type ProcessedMiniBlocksTracker struct {
}

// SetProcessedMiniBlockInfo does nothing as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) SetProcessedMiniBlockInfo(_ []byte, _ []byte, _ *processedMb.ProcessedMiniBlockInfo) {
}

// RemoveHeaderHash does nothing as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) RemoveHeaderHash(_ []byte) {
}

// RemoveMiniBlockHash does nothing as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) RemoveMiniBlockHash(_ []byte) {
}

// GetProcessedMiniBlocksInfo returns nil as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) GetProcessedMiniBlocksInfo(_ []byte) map[string]*processedMb.ProcessedMiniBlockInfo {
	return nil
}

// GetProcessedMiniBlockInfo returns nil as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) GetProcessedMiniBlockInfo(_ []byte) (*processedMb.ProcessedMiniBlockInfo, []byte) {
	return nil, nil
}

// IsMiniBlockFullyProcessed returns false as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) IsMiniBlockFullyProcessed(_ []byte, _ []byte) bool {
	return false
}

// ConvertProcessedMiniBlocksMapToSlice returns nil as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) ConvertProcessedMiniBlocksMapToSlice() []bootstrapStorage.MiniBlocksInMeta {
	return nil
}

// ConvertSliceToProcessedMiniBlocksMap does nothing as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) ConvertSliceToProcessedMiniBlocksMap(_ []bootstrapStorage.MiniBlocksInMeta) {
}

// DisplayProcessedMiniBlocks does nothing as it is a disabled component
func (pmbt *ProcessedMiniBlocksTracker) DisplayProcessedMiniBlocks() {
}

// IsInterfaceNil returns true if underlying object is nil
func (pmbt *ProcessedMiniBlocksTracker) IsInterfaceNil() bool {
	return pmbt == nil
}
