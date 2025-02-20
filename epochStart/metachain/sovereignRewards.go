package metachain

import (
	"bytes"
	"math/big"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
)

type fillBaseRewardsPerBlockPerNodeFunc func(baseRewardsPerNode *big.Int)

type sovereignRewards struct {
	*rewardsCreatorV2
}

// NewSovereignRewards creates a sovereign rewards computer
func NewSovereignRewards(rc *rewardsCreatorV2) (*sovereignRewards, error) {
	if check.IfNil(rc) {
		return nil, process.ErrNilRewardsCreator
	}

	sr := &sovereignRewards{
		rc,
	}

	sr.fillBaseRewardsPerBlockPerNodeFunc = sr.fillBaseRewardsPerBlockPerNode
	sr.flagDelegationSystemSCEnabled.SetValue(true)
	return sr, nil
}

// CreateRewardsMiniBlocks creates the rewards miniblocks according to economics data and validator info.
// This method applies the rewards according to the economics version 2 proposal, which takes into consideration
// stake top-up values per node. For each created reward tx, it will add directly to the reward txs pool
func (rc *sovereignRewards) CreateRewardsMiniBlocks(
	metaBlock data.MetaHeaderHandler,
	validatorsInfo state.ShardValidatorsInfoMapHandler,
	computedEconomics *block.Economics,
) (block.MiniBlockSlice, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}
	if computedEconomics == nil {
		return nil, epochStart.ErrNilEconomicsData
	}
	if validatorsInfo == nil {
		return nil, errNilValidatorsInfoMap
	}

	rc.mutRewardsData.Lock()
	defer rc.mutRewardsData.Unlock()

	log.Debug("sovereignRewards.CreateRewardsMiniBlocks",
		"totalToDistribute", computedEconomics.TotalToDistribute,
		"rewardsForProtocolSustainability", computedEconomics.RewardsForProtocolSustainability,
		"rewardsPerBlock", computedEconomics.RewardsPerBlock,
		"devFeesInEpoch", metaBlock.GetDevFeesInEpoch(),
		"rewardsForBlocks no fees", rc.economicsDataProvider.RewardsToBeDistributedForBlocks(),
		"numberOfBlocks", rc.economicsDataProvider.NumberOfBlocks(),
		"numberOfBlocksPerShard", rc.economicsDataProvider.NumberOfBlocksPerShard(),
	)

	miniBlocks := rc.initializeRewardsMiniBlocks()
	rc.clean()

	protRwdTx, protRwdShardId, err := rc.createProtocolSustainabilityRewardTransaction(metaBlock, computedEconomics)
	if err != nil {
		return nil, err
	}

	nodesRewardInfo, dustFromRewardsPerNode := rc.computeRewardsPerNode(validatorsInfo)
	log.Debug("arithmetic difference from dust rewards per node", "value", dustFromRewardsPerNode)

	dust, err := rc.addValidatorRewardsToMiniBlocks(metaBlock, miniBlocks, nodesRewardInfo)
	if err != nil {
		return nil, err
	}

	dust.Add(dust, dustFromRewardsPerNode)
	log.Debug("accumulated dust for protocol sustainability", "value", dust)

	rc.adjustProtocolSustainabilityRewards(protRwdTx, dust)
	err = rc.addProtocolRewardToMiniBlocks(protRwdTx, miniBlocks, protRwdShardId)
	if err != nil {
		return nil, err
	}

	return rc.finalizeMiniBlocks(miniBlocks), nil
}

func (rc *sovereignRewards) addValidatorRewardsToMiniBlocks(
	metaBlock data.HeaderHandler,
	miniBlocks block.MiniBlockSlice,
	nodesRewardInfo map[uint32][]*nodeRewardsData,
) (*big.Int, error) {
	rwdAddrValidatorInfo, accumulatedDust := rc.computeValidatorInfoPerRewardAddress(nodesRewardInfo)

	for _, rwdInfo := range rwdAddrValidatorInfo {
		rwdTx, rwdTxHash, err := rc.createRewardFromRwdInfo(rwdInfo, metaBlock)
		if err != nil {
			return nil, err
		}
		if rwdTx.Value.Cmp(zero) <= 0 {
			log.Error("negative rewards", "rcv", rwdTx.RcvAddr)
			continue
		}

		rc.accumulatedRewards.Add(rc.accumulatedRewards, rwdTx.Value)

		log.Debug("sovereignRewards.addValidatorRewardsToMiniBlocks",
			"epoch", rwdTx.GetEpoch(),
			"destination", rwdTx.GetRcvAddr(),
			"value", rwdTx.GetValue().String())

		rc.currTxs.AddTx(rwdTxHash, rwdTx)

		miniBlocks[0].TxHashes = append(miniBlocks[0].TxHashes, rwdTxHash)
	}

	return accumulatedDust, nil
}

func (rc *sovereignRewards) initializeRewardsMiniBlocks() block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 1)
	miniBlocks[0] = &block.MiniBlock{
		SenderShardID:   core.SovereignChainShardId,
		ReceiverShardID: core.SovereignChainShardId,
		Type:            block.RewardsBlock,
		TxHashes:        make([][]byte, 0),
	}

	return miniBlocks
}

func (rc *sovereignRewards) addProtocolRewardToMiniBlocks(
	protocolSustainabilityRwdTx *rewardTx.RewardTx,
	miniBlocks block.MiniBlockSlice,
	protocolSustainabilityShardId uint32,
) error {
	protocolSustainabilityRwdHash, errHash := core.CalculateHash(rc.marshalizer, rc.hasher, protocolSustainabilityRwdTx)
	if errHash != nil {
		return errHash
	}

	rc.currTxs.AddTx(protocolSustainabilityRwdHash, protocolSustainabilityRwdTx)
	miniBlocks[protocolSustainabilityShardId].TxHashes = append(miniBlocks[protocolSustainabilityShardId].TxHashes, protocolSustainabilityRwdHash)
	rc.protocolSustainabilityValue.Set(protocolSustainabilityRwdTx.Value)

	return nil
}

func (rc *sovereignRewards) finalizeMiniBlocks(miniBlocks block.MiniBlockSlice) block.MiniBlockSlice {
	shId := core.SovereignChainShardId
	sort.Slice(miniBlocks[shId].TxHashes, func(i, j int) bool {
		return bytes.Compare(miniBlocks[shId].TxHashes[i], miniBlocks[shId].TxHashes[j]) < 0
	})

	finalMiniBlocks := make(block.MiniBlockSlice, 0)
	if len(miniBlocks[shId].TxHashes) > 0 {
		finalMiniBlocks = append(finalMiniBlocks, miniBlocks[shId])
		rc.addExecutionOrdering(miniBlocks[shId].TxHashes)
	}

	return finalMiniBlocks
}

func (rc *sovereignRewards) fillBaseRewardsPerBlockPerNode(baseRewardsPerNode *big.Int) {
	shardID := core.SovereignChainShardId
	consensusSize := big.NewInt(int64(rc.nodesConfigProvider.ConsensusGroupSize(shardID)))

	rc.mapBaseRewardsPerBlockPerValidator = make(map[uint32]*big.Int)
	rc.mapBaseRewardsPerBlockPerValidator[shardID] = big.NewInt(0).Div(baseRewardsPerNode, consensusSize)

	log.Debug("baseRewardsPerBlockPerValidator", "shardID", shardID, "value", rc.mapBaseRewardsPerBlockPerValidator[shardID].String())
}

// IsInterfaceNil checks if the underlying pointer is nil
func (rc *sovereignRewards) IsInterfaceNil() bool {
	return rc == nil
}
