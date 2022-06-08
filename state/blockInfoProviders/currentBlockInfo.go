package blockInfoProviders

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	chainData "github.com/ElrondNetwork/elrond-go-core/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/holders"
)

var log = logger.GetOrCreate("state/blockinfoproviders")

type currentBlockInfo struct {
	chainHandler chainData.ChainHandler
}

// NewCurrentBlockInfo creates a new instance of type currentBlockInfo
func NewCurrentBlockInfo(chainHandler chainData.ChainHandler) (*currentBlockInfo, error) {
	if check.IfNil(chainHandler) {
		return nil, ErrNilChainHandler
	}

	return &currentBlockInfo{
		chainHandler: chainHandler,
	}, nil
}

// GetBlockInfo returns the current block info
func (provider *currentBlockInfo) GetBlockInfo() common.BlockInfo {
	block := provider.chainHandler.GetCurrentBlockHeader()
	if check.IfNil(block) {
		log.Debug("currentBlockInfo.GetBlockInfo: returning empty block info", "reason", "block is nil")
		return holders.NewBlockInfo(nil, 0, nil)
	}

	hash := provider.chainHandler.GetCurrentBlockHeaderHash()
	rootHash := provider.chainHandler.GetCurrentBlockRootHash()

	return holders.NewBlockInfo(hash, block.GetNonce(), rootHash)
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *currentBlockInfo) IsInterfaceNil() bool {
	return provider == nil
}
