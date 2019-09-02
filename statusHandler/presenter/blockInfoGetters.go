package presenter

import "github.com/ElrondNetwork/elrond-go/core"

// GetNumTxInBlock will return how many transactions are in block
func (psh *PresenterStatusHandler) GetNumTxInBlock() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumTxInBlock)
}

// GetNumMiniBlocks will return how many miniblocks are in a block
func (psh *PresenterStatusHandler) GetNumMiniBlocks() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumMiniBlocks)
}

// GetCrossCheckBlockHeight will return cross block height
func (psh *PresenterStatusHandler) GetCrossCheckBlockHeight() string {
	return psh.getFromCacheAsString(core.MetricCrossCheckBlockHeight)
}

// GetConsensusState will return consensus state of node
func (psh *PresenterStatusHandler) GetConsensusState() string {
	return psh.getFromCacheAsString(core.MetricConsensusState)
}

// GetConsensusRoundState will return consensus round state
func (psh *PresenterStatusHandler) GetConsensusRoundState() string {
	return psh.getFromCacheAsString(core.MetricConsensusRoundState)
}
