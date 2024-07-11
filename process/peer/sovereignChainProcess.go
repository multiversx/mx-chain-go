package peer

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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

	// TODO: MX-15586 Analyse this func and refactor it for sovereign epoch change
	scvs.updateShardDataPeerStateFunc = scvs.updateShardDataPeerState

	return scvs, nil
}

func (vs *sovereignChainValidatorStatistics) updateShardDataPeerState(
	header data.CommonHeaderHandler,
	cacheMap map[string]data.CommonHeaderHandler,
) error {

	var currentHeader data.CommonHeaderHandler

	if header.GetNonce() == vs.genesisNonce {
		return nil
	}

	//currentHeader, ok = cacheMap[string(h.HeaderHash)]
	//if !ok {
	//	return fmt.Errorf("%w - updateShardDataPeerState header from cache - hash: %s, round: %v, nonce: %v",
	//		process.ErrMissingHeader,
	//		hex.EncodeToString(h.HeaderHash),
	//		h.GetRound(),
	//		h.GetNonce())
	//}

	currentHeader = header
	h := currentHeader.(*block.SovereignChainHeader).Header

	epoch := computeEpoch(currentHeader)

	shardConsensus, shardInfoErr := vs.nodesCoordinator.ComputeConsensusGroup(h.PrevRandSeed, h.Round, h.ShardID, epoch)
	if shardInfoErr != nil {
		return shardInfoErr
	}

	log.Debug("updateShardDataPeerState - registering shard leader fees", "shard headerHash", "h.HeaderHash", "accumulatedFees", h.AccumulatedFees.String(), "developerFees", h.DeveloperFees.String())
	shardInfoErr = vs.updateValidatorInfoOnSuccessfulBlock(
		shardConsensus,
		h.PubKeysBitmap,
		big.NewInt(0).Sub(h.AccumulatedFees, h.DeveloperFees),
		h.ShardID,
	)
	if shardInfoErr != nil {
		return shardInfoErr
	}

	if h.Nonce == vs.genesisNonce+1 {
		return nil
	}

	prevShardData, shardInfoErr := vs.searchInMap(h.PrevHash, cacheMap)
	if shardInfoErr != nil {
		return shardInfoErr
	}

	shardInfoErr = vs.checkForMissedBlocks(
		h.Round,
		prevShardData.GetRound(),
		prevShardData.GetRandSeed(),
		h.ShardID,
		epoch,
	)
	if shardInfoErr != nil {
		return shardInfoErr
	}

	return nil
}
