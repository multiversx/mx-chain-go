package testscommon

import (
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
)

// ProcessedMiniBlocksTrackerStub -
type ProcessedMiniBlocksTrackerStub struct {
	SetProcessedMiniBlockInfoCalled            func(metaBlockHash []byte, miniBlockHash []byte, processedMbInfo *processedMb.ProcessedMiniBlockInfo)
	RemoveHeaderHashCalled                     func(metaBlockHash []byte)
	RemoveMiniBlockHashCalled                  func(miniBlockHash []byte)
	GetProcessedMiniBlocksInfoCalled           func(metaBlockHash []byte) map[string]*processedMb.ProcessedMiniBlockInfo
	GetProcessedMiniBlockInfoCalled            func(miniBlockHash []byte) (*processedMb.ProcessedMiniBlockInfo, []byte)
	IsMiniBlockFullyProcessedCalled            func(metaBlockHash []byte, miniBlockHash []byte) bool
	ConvertProcessedMiniBlocksMapToSliceCalled func() []bootstrapStorage.MiniBlocksInMeta
	ConvertSliceToProcessedMiniBlocksMapCalled func(miniBlocksInMetaBlocks []bootstrapStorage.MiniBlocksInMeta)
	DisplayProcessedMiniBlocksCalled           func()
}

// SetProcessedMiniBlockInfo -
func (pmbts *ProcessedMiniBlocksTrackerStub) SetProcessedMiniBlockInfo(metaBlockHash []byte, miniBlockHash []byte, processedMbInfo *processedMb.ProcessedMiniBlockInfo) {
	if pmbts.SetProcessedMiniBlockInfoCalled != nil {
		pmbts.SetProcessedMiniBlockInfoCalled(metaBlockHash, miniBlockHash, processedMbInfo)
	}
}

// RemoveHeaderHash -
func (pmbts *ProcessedMiniBlocksTrackerStub) RemoveHeaderHash(metaBlockHash []byte) {
	if pmbts.RemoveHeaderHashCalled != nil {
		pmbts.RemoveHeaderHashCalled(metaBlockHash)
	}
}

// RemoveMiniBlockHash -
func (pmbts *ProcessedMiniBlocksTrackerStub) RemoveMiniBlockHash(miniBlockHash []byte) {
	if pmbts.RemoveMiniBlockHashCalled != nil {
		pmbts.RemoveMiniBlockHashCalled(miniBlockHash)
	}
}

// GetProcessedMiniBlocksInfo -
func (pmbts *ProcessedMiniBlocksTrackerStub) GetProcessedMiniBlocksInfo(metaBlockHash []byte) map[string]*processedMb.ProcessedMiniBlockInfo {
	if pmbts.GetProcessedMiniBlocksInfoCalled != nil {
		return pmbts.GetProcessedMiniBlocksInfoCalled(metaBlockHash)
	}
	return make(map[string]*processedMb.ProcessedMiniBlockInfo)
}

// GetProcessedMiniBlockInfo -
func (pmbts *ProcessedMiniBlocksTrackerStub) GetProcessedMiniBlockInfo(miniBlockHash []byte) (*processedMb.ProcessedMiniBlockInfo, []byte) {
	if pmbts.GetProcessedMiniBlockInfoCalled != nil {
		return pmbts.GetProcessedMiniBlockInfoCalled(miniBlockHash)
	}
	return &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed:         false,
		IndexOfLastTxProcessed: -1,
	}, nil
}

// IsMiniBlockFullyProcessed -
func (pmbts *ProcessedMiniBlocksTrackerStub) IsMiniBlockFullyProcessed(metaBlockHash []byte, miniBlockHash []byte) bool {
	if pmbts.IsMiniBlockFullyProcessedCalled != nil {
		return pmbts.IsMiniBlockFullyProcessedCalled(metaBlockHash, miniBlockHash)
	}
	return false
}

// ConvertProcessedMiniBlocksMapToSlice -
func (pmbts *ProcessedMiniBlocksTrackerStub) ConvertProcessedMiniBlocksMapToSlice() []bootstrapStorage.MiniBlocksInMeta {
	if pmbts.ConvertProcessedMiniBlocksMapToSliceCalled != nil {
		return pmbts.ConvertProcessedMiniBlocksMapToSliceCalled()
	}
	return nil
}

// ConvertSliceToProcessedMiniBlocksMap -
func (pmbts *ProcessedMiniBlocksTrackerStub) ConvertSliceToProcessedMiniBlocksMap(miniBlocksInMetaBlocks []bootstrapStorage.MiniBlocksInMeta) {
	if pmbts.ConvertSliceToProcessedMiniBlocksMapCalled != nil {
		pmbts.ConvertSliceToProcessedMiniBlocksMapCalled(miniBlocksInMetaBlocks)
	}
}

// DisplayProcessedMiniBlocks -
func (pmbts *ProcessedMiniBlocksTrackerStub) DisplayProcessedMiniBlocks() {
	if pmbts.DisplayProcessedMiniBlocksCalled != nil {
		pmbts.DisplayProcessedMiniBlocksCalled()
	}
}

// IsInterfaceNil -
func (pmbts *ProcessedMiniBlocksTrackerStub) IsInterfaceNil() bool {
	return pmbts == nil
}
