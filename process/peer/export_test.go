package peer

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// CheckForMissedBlocks -
func (vs *validatorStatistics) CheckForMissedBlocks(
	currentHeaderRound uint64,
	previousHeaderRound uint64,
	prevRandSeed []byte,
	shardId uint32,
	epoch uint32,
) error {
	return vs.checkForMissedBlocks(currentHeaderRound, previousHeaderRound, prevRandSeed, shardId, epoch)
}

// LoadPeerAccount -
func (vs *validatorStatistics) LoadPeerAccount(address []byte) (state.PeerAccountHandler, error) {
	return vs.loadPeerAccount(address)
}

// GetMatchingPrevShardData -
func (vs *validatorStatistics) GetMatchingPrevShardData(currentShardData block.ShardData, shardInfo []block.ShardData) *block.ShardData {
	return vs.getMatchingPrevShardData(currentShardData, shardInfo)
}

// GetLeaderDecreaseCount -
func (vs *validatorStatistics) GetLeaderDecreaseCount(key []byte) uint32 {
	vs.mutValidatorStatistics.RLock()
	defer vs.mutValidatorStatistics.RUnlock()

	return vs.missedBlocksCounters.get(key).leaderDecreaseCount
}

// UpdateMissedBlocksCounters -
func (vs *validatorStatistics) UpdateMissedBlocksCounters() error {
	return vs.updateMissedBlocksCounters()
}

// GetCache -
func (ptp *PeerTypeProvider) GetCache() map[string]*peerListAndShard {
	ptp.mutCache.RLock()
	defer ptp.mutCache.RUnlock()
	return ptp.cache
}

// GetCache -
func (ptp *validatorsProvider) GetCache() map[string]*state.ValidatorApiResponse {
	ptp.lock.RLock()
	defer ptp.lock.RUnlock()
	return ptp.cache
}

// UpdateShardDataPeerState -
func (vs *validatorStatistics) UpdateShardDataPeerState(
	header data.HeaderHandler,
	cacheMap map[string]data.HeaderHandler,
) error {
	return vs.updateShardDataPeerState(header, cacheMap)
}

// GetActualList -
func GetActualList(peerAccount state.PeerAccountHandler) string {
	return getActualList(peerAccount)
}
