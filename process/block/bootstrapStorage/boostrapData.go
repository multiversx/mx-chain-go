package bootstrapStorage

import "github.com/multiversx/mx-chain-go/common"

// IsFullyProcessed returns if the mini block at the given index is fully processed or not
func (m *MiniBlocksInMeta) IsFullyProcessed(index int) bool {
	fullyProcessed := true
	if m.FullyProcessed != nil && index < len(m.FullyProcessed) {
		fullyProcessed = m.FullyProcessed[index]
	}

	return fullyProcessed
}

// GetIndexOfLastTxProcessedInMiniBlock returns index of the last transaction processed in the mini block with the given index
func (m *MiniBlocksInMeta) GetIndexOfLastTxProcessedInMiniBlock(index int) int32 {
	indexOfLastTxProcessed := common.MaxIndexOfTxInMiniBlock
	if m.IndexOfLastTxProcessed != nil && index < len(m.IndexOfLastTxProcessed) {
		indexOfLastTxProcessed = m.IndexOfLastTxProcessed[index]
	}

	return indexOfLastTxProcessed
}
