package interceptedBlocks

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/sharding"
)

type interceptedSovereignBlockHeader struct {
	*InterceptedHeader
}

// NewSovereignInterceptedBlockHeader creates a new intercepted sovereign block header
func NewSovereignInterceptedBlockHeader(blockHeaderInterceptor *InterceptedHeader) (*interceptedSovereignBlockHeader, error) {
	if check.IfNil(blockHeaderInterceptor) {
		return nil, errNilInterceptedBlockHeader
	}

	sovInterceptedBlock := &interceptedSovereignBlockHeader{
		blockHeaderInterceptor,
	}

	sovInterceptedBlock.mbHeadersChecker = sovInterceptedBlock
	return sovInterceptedBlock, nil
}

func (isbh *interceptedSovereignBlockHeader) checkMiniBlocksHeaders(mbHeaders []data.MiniBlockHeaderHandler, coordinator sharding.Coordinator) error {
	return checkMiniBlocksHeaders(mbHeaders, coordinator, core.MainChainShardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (isbh *interceptedSovereignBlockHeader) IsInterfaceNil() bool {
	return isbh == nil
}
