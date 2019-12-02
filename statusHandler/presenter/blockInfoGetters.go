package presenter

import (
	"github.com/ElrondNetwork/elrond-go/core/constants"
)

// GetNumTxInBlock will return how many transactions are in block
func (psh *PresenterStatusHandler) GetNumTxInBlock() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricNumTxInBlock)
}

// GetNumMiniBlocks will return how many miniblocks are in a block
func (psh *PresenterStatusHandler) GetNumMiniBlocks() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricNumMiniBlocks)
}

// GetCrossCheckBlockHeight will return cross block height
func (psh *PresenterStatusHandler) GetCrossCheckBlockHeight() string {
	return psh.getFromCacheAsString(constants.MetricCrossCheckBlockHeight)
}

// GetConsensusState will return consensus state of node
func (psh *PresenterStatusHandler) GetConsensusState() string {
	return psh.getFromCacheAsString(constants.MetricConsensusState)
}

// GetConsensusRoundState will return consensus round state
func (psh *PresenterStatusHandler) GetConsensusRoundState() string {
	return psh.getFromCacheAsString(constants.MetricConsensusRoundState)
}

// GetCurrentBlockHash will return current block hash
func (psh *PresenterStatusHandler) GetCurrentBlockHash() string {
	return psh.getFromCacheAsString(constants.MetricCurrentBlockHash)
}

// GetCurrentRoundTimestamp will return current round timestamp
func (psh *PresenterStatusHandler) GetCurrentRoundTimestamp() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricCurrentRoundTimestamp)
}

// GetBlockSize will return current block size
func (psh *PresenterStatusHandler) GetBlockSize() uint64 {
	miniBlocksSize := psh.getFromCacheAsUint64(constants.MetricMiniBlocksSize)
	headerSize := psh.getFromCacheAsUint64(constants.MetricHeaderSize)

	return miniBlocksSize + headerSize
}

// GetHighestFinalBlockInShard will return highest nonce block notarized by metachain for current shard
func (psh *PresenterStatusHandler) GetHighestFinalBlockInShard() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricHighestFinalBlockInShard)
}
