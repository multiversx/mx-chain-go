package peer

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignChainValidatorStatistics struct {
	*validatorStatistics
}

// NewSovereignChainValidatorStatisticsProcessor instantiates a new sovereignChainValidatorStatistics structure
// responsible for keeping account of each validator actions in the consensus process
func NewSovereignChainValidatorStatisticsProcessor(validatorStatistics *validatorStatistics) (*sovereignChainValidatorStatistics, error) {
	if validatorStatistics == nil {
		return nil, process.ErrNilValidatorStatistics
	}

	scvs := &sovereignChainValidatorStatistics{
		validatorStatistics,
	}

	scvs.updateShardDataPeerStateFunc = scvs.updateShardDataPeerState
	return scvs, nil
}

func (vs *sovereignChainValidatorStatistics) updateShardDataPeerState(
	header data.CommonHeaderHandler,
	cacheMap map[string]data.CommonHeaderHandler,
) error {
	if header.GetNonce() == vs.genesisNonce {
		return nil
	}

	epoch := computeEpoch(header)
	shardConsensus, shardInfoErr := vs.nodesCoordinator.ComputeConsensusGroup(
		header.GetPrevRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		epoch,
	)
	if shardInfoErr != nil {
		return shardInfoErr
	}

	log.Debug("updateShardDataPeerState - registering shard leader fees",
		"shard header round", header.GetRound(),
		"accumulatedFees", header.GetAccumulatedFees().String(),
		"developerFees", header.GetDeveloperFees().String(),
	)
	shardInfoErr = vs.updateValidatorInfoOnSuccessfulBlock(
		shardConsensus,
		header.GetPubKeysBitmap(),
		big.NewInt(0).Sub(header.GetAccumulatedFees(), header.GetDeveloperFees()),
		header.GetShardID(),
	)
	if shardInfoErr != nil {
		return shardInfoErr
	}

	if header.GetNonce() == vs.genesisNonce+1 {
		return nil
	}

	prevShardData, shardInfoErr := vs.searchInMap(header.GetPrevHash(), cacheMap)
	if shardInfoErr != nil {
		return shardInfoErr
	}

	return vs.checkForMissedBlocks(
		header.GetRound(),
		prevShardData.GetRound(),
		prevShardData.GetRandSeed(),
		header.GetShardID(),
		epoch,
	)
}
