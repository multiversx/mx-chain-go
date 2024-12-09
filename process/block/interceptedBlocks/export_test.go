package interceptedBlocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

// IsMetaHeaderOutOfRange -
func (imh *InterceptedMetaHeader) IsMetaHeaderOutOfRange() bool {
	return imh.isMetaHeaderEpochOutOfRange()
}

// CheckMiniBlocksHeaders -
func (isbh *interceptedSovereignBlockHeader) CheckMiniBlocksHeaders(mbHeaders []data.MiniBlockHeaderHandler, coordinator sharding.Coordinator) error {
	return isbh.checkMiniBlocksHeaders(mbHeaders, coordinator)
}
