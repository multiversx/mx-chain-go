package presenter

import (
	"github.com/multiversx/mx-chain-go/common"
)

// GetNumTxInBlock will return how many transactions are in block
func (psh *PresenterStatusHandler) GetNumTxInBlock() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNumTxInBlock)
}

// GetNumMiniBlocks will return how many miniblocks are in a block
func (psh *PresenterStatusHandler) GetNumMiniBlocks() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNumMiniBlocks)
}

// GetCrossCheckBlockHeight will return cross block height
func (psh *PresenterStatusHandler) GetCrossCheckBlockHeight() string {
	return psh.getFromCacheAsString(common.MetricCrossCheckBlockHeight)
}

// GetConsensusState will return consensus state of node
func (psh *PresenterStatusHandler) GetConsensusState() string {
	return psh.getFromCacheAsString(common.MetricConsensusState)
}

// GetConsensusRoundState will return consensus round state
func (psh *PresenterStatusHandler) GetConsensusRoundState() string {
	return psh.getFromCacheAsString(common.MetricConsensusRoundState)
}

// GetCurrentBlockHash will return current block hash
func (psh *PresenterStatusHandler) GetCurrentBlockHash() string {
	return psh.getFromCacheAsString(common.MetricCurrentBlockHash)
}

// GetEpochNumber will return current epoch
func (psh *PresenterStatusHandler) GetEpochNumber() uint64 {
	return psh.getFromCacheAsUint64(common.MetricEpochNumber)
}

// GetCurrentRoundTimestamp will return current round timestamp
func (psh *PresenterStatusHandler) GetCurrentRoundTimestamp() uint64 {
	return psh.getFromCacheAsUint64(common.MetricCurrentRoundTimestamp)
}

// GetBlockSize will return current block size
func (psh *PresenterStatusHandler) GetBlockSize() uint64 {
	miniBlocksSize := psh.getFromCacheAsUint64(common.MetricMiniBlocksSize)
	headerSize := psh.getFromCacheAsUint64(common.MetricHeaderSize)

	return miniBlocksSize + headerSize
}

// GetHighestFinalBlock will return the highest nonce block notarized by metachain for current shard
func (psh *PresenterStatusHandler) GetHighestFinalBlock() uint64 {
	return psh.getFromCacheAsUint64(common.MetricHighestFinalBlock)
}
