package metachain

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/validatorInfo"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.RewardsCreator = (*rewardsCreatorV2)(nil)

type nodeRewardsData struct {
	baseReward   *big.Int
	topUpReward  *big.Int
	fullRewards  *big.Int
	topUpStake   *big.Int
	powerInShard *big.Int
	valInfo      *state.ValidatorInfo
}

// RewardsCreatorArgsV2 holds the data required to create end of epoch rewards
type RewardsCreatorArgsV2 struct {
	BaseRewardsCreatorArgs
	StakingDataProvider   epochStart.StakingDataProvider
	EconomicsDataProvider epochStart.EpochEconomicsDataProvider
	TopUpRewardFactor     float64
	TopUpGradientPoint    *big.Int
}

type rewardsCreatorV2 struct {
	*baseRewardsCreator
	stakingDataProvider   epochStart.StakingDataProvider
	economicsDataProvider epochStart.EpochEconomicsDataProvider
	topUpRewardFactor     float64
	topUpGradientPoint    *big.Int
}

// NewRewardsCreatorV2 creates a new rewards creator object
func NewRewardsCreatorV2(args RewardsCreatorArgsV2) (*rewardsCreatorV2, error) {
	brc, err := NewBaseRewardsCreator(args.BaseRewardsCreatorArgs)
	if err != nil {
		return nil, err
	}

	if check.IfNil(args.StakingDataProvider) {
		return nil, epochStart.ErrNilStakingDataProvider
	}
	if check.IfNil(args.EconomicsDataProvider) {
		return nil, epochStart.ErrNilEconomicsDataProvider
	}
	if args.TopUpGradientPoint.Cmp(zero) < 0 {
		return nil, epochStart.ErrInvalidRewardsTopUpGradientPoint
	}
	if args.TopUpRewardFactor < 0 || args.TopUpRewardFactor > 1 {
		return nil, epochStart.ErrInvalidRewardsTopUpFactor
	}

	rc := &rewardsCreatorV2{
		baseRewardsCreator:    brc,
		economicsDataProvider: args.EconomicsDataProvider,
		stakingDataProvider:   args.StakingDataProvider,
		topUpRewardFactor:     args.TopUpRewardFactor,
		topUpGradientPoint:    args.TopUpGradientPoint,
	}

	return rc, nil
}

// CreateRewardsMiniBlocks creates the rewards miniblocks according to economics data and validator info.
// This method applies the rewards according to the economics version 2 proposal, which takes into consideration
// stake top-up values per node
func (rc *rewardsCreatorV2) CreateRewardsMiniBlocks(
	metaBlock *block.MetaBlock,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) (block.MiniBlockSlice, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}
	if computedEconomics == nil {
		return nil, epochStart.ErrNilEconomicsData
	}

	rc.mutRewardsData.Lock()
	defer rc.mutRewardsData.Unlock()

	log.Debug("rewardsCreatorV2.CreateRewardsMiniBlocks",
		"totalToDistribute", computedEconomics.TotalToDistribute,
		"rewardsForProtocolSustainability", computedEconomics.RewardsForProtocolSustainability,
		"rewardsPerBlock", computedEconomics.RewardsPerBlock,
		"devFeesInEpoch", metaBlock.DevFeesInEpoch,
		"rewardsForBlocks no fees", rc.economicsDataProvider.RewardsToBeDistributedForBlocks(),
		"numberOfBlocks", rc.economicsDataProvider.NumberOfBlocks(),
		"numberOfBlocksPerShard", rc.economicsDataProvider.NumberOfBlocksPerShard(),
	)

	miniBlocks := rc.initializeRewardsMiniBlocks()
	rc.clean()
	rc.flagDelegationSystemSCEnabled.Toggle(metaBlock.GetEpoch() >= rc.delegationSystemSCEnableEpoch)

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

func (rc *rewardsCreatorV2) adjustProtocolSustainabilityRewards(protocolSustainabilityRwdTx *rewardTx.RewardTx, dustRewards *big.Int) {
	if protocolSustainabilityRwdTx.Value.Cmp(big.NewInt(0)) < 0 {
		log.Error("negative rewards protocol sustainability")
		protocolSustainabilityRwdTx.Value.SetUint64(0)
	}

	if dustRewards.Cmp(big.NewInt(0)) < 0 {
		log.Error("trying to adjust protocol rewards with negative value", "dustRewards", dustRewards.String())
		return
	}

	protocolSustainabilityRwdTx.Value.Add(protocolSustainabilityRwdTx.Value, dustRewards)

	log.Debug("baseRewardsCreator.adjustProtocolSustainabilityRewards - rewardsCreatorV2",
		"epoch", protocolSustainabilityRwdTx.GetEpoch(),
		"destination", protocolSustainabilityRwdTx.GetRcvAddr(),
		"value", protocolSustainabilityRwdTx.GetValue().String())

	rc.protocolSustainabilityValue.Set(protocolSustainabilityRwdTx.Value)
}

// VerifyRewardsMiniBlocks verifies if received rewards miniblocks are correct
func (rc *rewardsCreatorV2) VerifyRewardsMiniBlocks(
	metaBlock *block.MetaBlock,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	computedEconomics *block.Economics,
) error {
	if check.IfNil(metaBlock) {
		return epochStart.ErrNilHeaderHandler
	}

	createdMiniBlocks, err := rc.CreateRewardsMiniBlocks(metaBlock, validatorsInfo, computedEconomics)
	if err != nil {
		return err
	}

	return rc.verifyCreatedRewardMiniBlocksWithMetaBlock(metaBlock, createdMiniBlocks)
}

func (rc *rewardsCreatorV2) addValidatorRewardsToMiniBlocks(
	metaBlock *block.MetaBlock,
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
			continue
		}

		rc.accumulatedRewards.Add(rc.accumulatedRewards, rwdTx.Value)
		mbId := rc.shardCoordinator.ComputeId([]byte(rwdInfo.address))
		if mbId == core.MetachainShardId {
			mbId = rc.shardCoordinator.NumberOfShards()

			if !rc.flagDelegationSystemSCEnabled.IsSet() || !rc.isSystemDelegationSC(rwdTx.RcvAddr) {
				log.Debug("rewardsCreator.addValidatorRewardsToMiniBlocks - not supported metaChain address",
					"move to protocol vault", rwdTx.GetValue())
				accumulatedDust.Add(accumulatedDust, rwdTx.GetValue())
				continue
			}
		}

		if rwdTx.Value.Cmp(zero) < 0 {
			log.Error("negative rewards", "rcv", rwdTx.RcvAddr)
			continue
		}

		log.Debug("rewardsCreatorV2.addValidatorRewardsToMiniBlocks",
			"epoch", rwdTx.GetEpoch(),
			"destination", rwdTx.GetRcvAddr(),
			"value", rwdTx.GetValue().String())

		rc.currTxs.AddTx(rwdTxHash, rwdTx)
		miniBlocks[mbId].TxHashes = append(miniBlocks[mbId].TxHashes, rwdTxHash)
	}

	return accumulatedDust, nil
}

func (rc *rewardsCreatorV2) computeValidatorInfoPerRewardAddress(
	nodesRewardInfo map[uint32][]*nodeRewardsData,
) (map[string]*rewardInfoData, *big.Int) {

	rwdAddrValidatorInfo := make(map[string]*rewardInfoData)
	accumulatedUnassigned := big.NewInt(0)
	distributedLeaderFees := big.NewInt(0)

	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			if nodeInfo.valInfo.LeaderSuccess == 0 && nodeInfo.valInfo.ValidatorSuccess == 0 {
				accumulatedUnassigned.Add(accumulatedUnassigned, nodeInfo.fullRewards)
				continue
			}

			rwdInfo, ok := rwdAddrValidatorInfo[string(nodeInfo.valInfo.RewardAddress)]
			if !ok {
				rwdInfo = &rewardInfoData{
					accumulatedFees:     big.NewInt(0),
					rewardsFromProtocol: big.NewInt(0),
					address:             string(nodeInfo.valInfo.RewardAddress),
				}
				rwdAddrValidatorInfo[string(nodeInfo.valInfo.RewardAddress)] = rwdInfo
			}

			distributedLeaderFees.Add(distributedLeaderFees, nodeInfo.valInfo.AccumulatedFees)
			rwdInfo.accumulatedFees.Add(rwdInfo.accumulatedFees, nodeInfo.valInfo.AccumulatedFees)
			rwdInfo.rewardsFromProtocol.Add(rwdInfo.rewardsFromProtocol, nodeInfo.fullRewards)
		}
	}

	estimatedLeaderFees := rc.economicsDataProvider.LeaderFees()
	dustLeaderFees := big.NewInt(0).Sub(estimatedLeaderFees, distributedLeaderFees)
	if dustLeaderFees.Cmp(big.NewInt(0)) < 0 {
		log.Warn("computeValidatorInfoPerRewardAddress negative rewards",
			"estimatedLeaderFees", estimatedLeaderFees,
			"distributedLeaderFees", distributedLeaderFees,
			"dust", dustLeaderFees)
	} else {
		accumulatedUnassigned.Add(accumulatedUnassigned, dustLeaderFees)
	}

	return rwdAddrValidatorInfo, accumulatedUnassigned
}

// IsInterfaceNil return true if underlying object is nil
func (rc *rewardsCreatorV2) IsInterfaceNil() bool {
	return rc == nil
}

func (rc *rewardsCreatorV2) computeRewardsPerNode(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) (map[uint32][]*nodeRewardsData, *big.Int) {

	var baseRewardsPerBlock *big.Int

	nodesRewardInfo := rc.initNodesRewardsInfo(validatorsInfo)

	// totalTopUpEligible is the cumulative top-up stake value for eligible nodes
	totalTopUpEligible := rc.stakingDataProvider.GetTotalTopUpStakeEligibleNodes()
	totalString := totalTopUpEligible.String()
	if len(totalString) > 18 {
		totalString = totalString[0 : len(totalString)-18]
	} else {
		totalString = "0"
	}

	totalTopupString := fmt.Sprintf("topup eligible %s EGLD", totalString)
	log.Info(totalTopupString)

	remainingToBeDistributed := rc.economicsDataProvider.RewardsToBeDistributedForBlocks()
	topUpRewards := rc.computeTopUpRewards(remainingToBeDistributed, totalTopUpEligible)
	baseRewards := big.NewInt(0).Sub(remainingToBeDistributed, topUpRewards)

	topUpRewardsString := topUpRewards.String()
	if len(topUpRewardsString) > 18 {
		topUpRewardsString = topUpRewardsString[0 : len(topUpRewardsString)-18]
	} else {
		topUpRewardsString = "0"
	}

	baseRewardsString := baseRewards.String()
	baseRewardsString = baseRewardsString[0 : len(baseRewardsString)-18]

	baseRewardValue := fmt.Sprintf("baseRewards %s EGLD", baseRewardsString)
	topupRewardValue := fmt.Sprintf("topUpRewards %s EGLD", topUpRewardsString)
	log.Info(baseRewardValue)
	log.Info(topupRewardValue)

	nbBlocks := big.NewInt(int64(rc.economicsDataProvider.NumberOfBlocks()))
	if nbBlocks.Cmp(zero) <= 0 {
		baseRewardsPerBlock = big.NewInt(0)
	} else {
		baseRewardsPerBlock = big.NewInt(0).Div(baseRewards, nbBlocks)
	}

	rc.fillBaseRewardsPerBlockPerNode(baseRewardsPerBlock)

	accumulatedDust := big.NewInt(0)
	dust := rc.computeBaseRewardsPerNode(nodesRewardInfo, baseRewards)
	accumulatedDust.Add(accumulatedDust, dust)
	dust = rc.computeTopUpRewardsPerNode(nodesRewardInfo, topUpRewards)
	accumulatedDust.Add(accumulatedDust, dust)
	aggregateBaseAndTopUpRewardsPerNode(nodesRewardInfo)

	return nodesRewardInfo, accumulatedDust
}

func (rc *rewardsCreatorV2) initNodesRewardsInfo(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) map[uint32][]*nodeRewardsData {
	nodesRewardsInfo := make(map[uint32][]*nodeRewardsData)

	for shardID, valInfoList := range validatorsInfo {
		nodesRewardsInfo[shardID] = make([]*nodeRewardsData, 0, len(valInfoList))
		for _, valInfo := range valInfoList {
			if validatorInfo.WasEligibleInCurrentEpoch(valInfo) {
				rewardsInfo := &nodeRewardsData{
					baseReward:   big.NewInt(0),
					topUpReward:  big.NewInt(0),
					fullRewards:  big.NewInt(0),
					topUpStake:   big.NewInt(0),
					powerInShard: big.NewInt(0),
					valInfo:      valInfo,
				}
				nodesRewardsInfo[shardID] = append(nodesRewardsInfo[shardID], rewardsInfo)
			}
		}
	}

	return nodesRewardsInfo
}

func (rc *rewardsCreatorV2) computeBaseRewardsPerNode(
	nodesRewardInfo map[uint32][]*nodeRewardsData,
	baseRewards *big.Int,
) *big.Int {
	accumulatedRewards := big.NewInt(0)

	for shardID, nodeRewardsInfoList := range nodesRewardInfo {
		for _, nodeRewardsInfo := range nodeRewardsInfoList {
			nodeRewardsInfo.baseReward = big.NewInt(0).Mul(
				rc.mapBaseRewardsPerBlockPerValidator[shardID],
				big.NewInt(int64(nodeRewardsInfo.valInfo.NumSelectedInSuccessBlocks)))
			accumulatedRewards.Add(accumulatedRewards, nodeRewardsInfo.baseReward)
		}
	}

	return big.NewInt(0).Sub(baseRewards, accumulatedRewards)
}

func (rc *rewardsCreatorV2) computeTopUpRewardsPerNode(
	nodesRewardInfo map[uint32][]*nodeRewardsData,
	topUpRewards *big.Int,
) *big.Int {

	accumulatedTopUpRewards := big.NewInt(0)
	rc.getTopUpForAllEligibleNodes(nodesRewardInfo)
	topUpRewardPerShard := rc.computeTopUpRewardsPerShard(topUpRewards, nodesRewardInfo)
	totalPowerInShard := computeNodesPowerInShard(nodesRewardInfo)

	// topUpRewardPerNodeInShardX = nodePowerInShardX*topUpRewardsShardX/totalPowerInShardX
	for shardID, nodeInfoList := range nodesRewardInfo {
		if totalPowerInShard[shardID].Cmp(zero) == 0 {
			log.Debug("rewardsCreatorV2.computeTopUpRewardsPerNode",
				"shardPower", 0,
				"shardID", shardID)
			continue
		}

		for _, nodeInfo := range nodeInfoList {
			nodeInfo.topUpReward = big.NewInt(0).Mul(nodeInfo.powerInShard, topUpRewardPerShard[shardID])
			nodeInfo.topUpReward.Div(nodeInfo.topUpReward, totalPowerInShard[shardID])
			accumulatedTopUpRewards.Add(accumulatedTopUpRewards, nodeInfo.topUpReward)
		}
	}

	return big.NewInt(0).Sub(topUpRewards, accumulatedTopUpRewards)
}

//      (2*k/pi)*atan(x/p), where:
//     k is the rewards per day limit for top-up stake k = c * economics.TotalToDistribute, c - constant, e.g c = 0.25
//     x is the cumulative top-up stake value for eligible nodes
//     p is the cumulative eligible stake where rewards per day reach 1/2 of k (includes topUp for the eligible nodes)
//     pi is the mathematical constant pi = 3.1415...
func (rc *rewardsCreatorV2) computeTopUpRewards(totalToDistribute *big.Int, totalTopUpEligible *big.Int) *big.Int {
	if totalToDistribute.Cmp(zero) <= 0 || totalTopUpEligible.Cmp(zero) <= 0 {
		return big.NewInt(0)
	}

	log.Info("totalToDistribute", "value", totalToDistribute.String())
	log.Info("totalTopUpEligible", "value", totalTopUpEligible)
	log.Info("topUpGradientPoint", "value", rc.topUpGradientPoint)
	log.Info("topUpRewardFactor", "value", rc.topUpRewardFactor)

	// k = c * economics.TotalToDistribute, c = top-up reward factor (constant)
	k := core.GetIntTrimmedPercentageOfValue(totalToDistribute, rc.topUpRewardFactor)
	// p is the cumulative eligible stake where rewards per day reach 1/2 of k (constant)
	// x/p - argument for atan
	totalTopUpEligibleFloat := big.NewFloat(0).SetInt(totalTopUpEligible)
	topUpGradientPointFloat := big.NewFloat(0).SetInt(rc.topUpGradientPoint)

	floatArg, _ := big.NewFloat(0).Quo(totalTopUpEligibleFloat, topUpGradientPointFloat).Float64()
	// atan(x/p)
	res1 := math.Atan(floatArg)
	// 2*k/pi
	res2 := big.NewFloat(0).SetInt(big.NewInt(0).Mul(k, big.NewInt(2)))
	res2 = big.NewFloat(0).Quo(res2, big.NewFloat(math.Pi))
	// topUpReward:= (2*k/pi)*atan(x/p)
	topUpRewards, _ := big.NewFloat(0).Mul(big.NewFloat(res1), res2).Int(nil)

	return topUpRewards
}

// top-Up rewards are distributed to shard nodes, depending on the top-up ratio and the number of blocks
// ratio in the shard, with respect to the entire network
func (rc *rewardsCreatorV2) computeTopUpRewardsPerShard(
	topUpRewards *big.Int,
	nodesRewardInfo map[uint32][]*nodeRewardsData,
) map[uint32]*big.Int {
	blocksPerShard := rc.economicsDataProvider.NumberOfBlocksPerShard()

	if len(blocksPerShard) == 0 {
		log.Warn("rewardsCreatorV2.computeTopUpRewardsPerShard",
			"error", "empty blocksPerShard")
		blocksPerShard = make(map[uint32]uint64)
		shardMap := createShardsMap(rc.shardCoordinator)
		for shardID := range shardMap {
			// initialize the rewards per shard
			blocksPerShard[shardID] = 0
		}
	}
	shardsTopUp := computeTopUpPerShard(nodesRewardInfo)
	totalPower, shardPower := computeShardsPower(shardsTopUp, blocksPerShard)

	return computeRewardsForPowerPerShard(shardPower, totalPower, topUpRewards)
}

func computeShardsPower(
	shardsTopUp map[uint32]*big.Int,
	blocksPerShard map[uint32]uint64,
) (*big.Int, map[uint32]*big.Int) {
	shardPower := make(map[uint32]*big.Int)
	totalPower := big.NewInt(0)
	// shardXPower = shardXTopUp * shardXProducedBlocks
	for shardID, nbBlocks := range blocksPerShard {
		shardPower[shardID] = big.NewInt(0).Mul(big.NewInt(int64(nbBlocks)), shardsTopUp[shardID])
		totalPower.Add(totalPower, shardPower[shardID])
	}
	return totalPower, shardPower
}

func computeTopUpPerShard(nodesRewardInfo map[uint32][]*nodeRewardsData) map[uint32]*big.Int {
	shardsTopUp := make(map[uint32]*big.Int)
	for shardID, nodeRewardInfoList := range nodesRewardInfo {
		shardsTopUp[shardID] = big.NewInt(0)
		for _, nodeInfo := range nodeRewardInfoList {
			shardsTopUp[shardID].Add(shardsTopUp[shardID], nodeInfo.topUpStake)
		}
	}
	return shardsTopUp
}

func computeRewardsForPowerPerShard(
	shardPower map[uint32]*big.Int,
	totalPower *big.Int,
	topUpRewards *big.Int,
) map[uint32]*big.Int {
	rewardsTopUpPerShard := make(map[uint32]*big.Int)
	for shardID := range shardPower {
		// initialize the rewards per shard
		rewardsTopUpPerShard[shardID] = big.NewInt(0)
	}

	if totalPower.Cmp(zero) <= 0 {
		return rewardsTopUpPerShard
	}

	// shardXTopUpRewards = shardXPower/totalPower * topUpRewards
	for shardID, power := range shardPower {
		rewardsTopUpPerShard[shardID] = big.NewInt(0).Mul(power, topUpRewards)
		rewardsTopUpPerShard[shardID].Div(rewardsTopUpPerShard[shardID], totalPower)
	}

	return rewardsTopUpPerShard
}

func computeNodesPowerInShard(
	nodesRewardInfo map[uint32][]*nodeRewardsData,
) map[uint32]*big.Int {
	totalShardNodesPower := make(map[uint32]*big.Int)

	for shardID, nodeInfoList := range nodesRewardInfo {
		totalShardNodesPower[shardID] = big.NewInt(0)
		for _, nodeInfo := range nodeInfoList {
			nodeInfo.powerInShard = computeNodePowerInShard(nodeInfo.valInfo, nodeInfo.topUpStake)
			totalShardNodesPower[shardID].Add(totalShardNodesPower[shardID], nodeInfo.powerInShard)
		}
	}

	return totalShardNodesPower
}

// power in epoch is computed as nbBlocks*nodeTopUp, where nbBlocks represents the number of blocks the node
// participated at creation/validation
func computeNodePowerInShard(nodeInfo *state.ValidatorInfo, nodeTopUp *big.Int) *big.Int {
	// if node was offline, it had no power, so the rewards should go to the others
	if nodeInfo.LeaderSuccess == 0 && nodeInfo.ValidatorSuccess == 0 {
		return big.NewInt(0)
	}

	nbBlocks := big.NewInt(0).SetUint64(uint64(nodeInfo.NumSelectedInSuccessBlocks))
	return big.NewInt(0).Mul(nbBlocks, nodeTopUp)
}

func aggregateBaseAndTopUpRewardsPerNode(nodesRewardInfo map[uint32][]*nodeRewardsData) {
	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			nodeInfo.fullRewards = big.NewInt(0).Add(
				nodeInfo.baseReward,
				nodeInfo.topUpReward)
		}
	}
}

func (rc *rewardsCreatorV2) getTopUpForAllEligibleNodes(
	nodesRewardInfo map[uint32][]*nodeRewardsData,
) {
	var err error
	for _, nodeRewardList := range nodesRewardInfo {
		for _, nodeInfo := range nodeRewardList {
			nodeInfo.topUpStake, err = rc.stakingDataProvider.GetNodeStakedTopUp(nodeInfo.valInfo.GetPublicKey())
			if err != nil {
				nodeInfo.topUpStake = big.NewInt(0)

				log.Warn("rewardsCreatorV2.getTopUpForAllEligible",
					"error", err.Error(),
					"blsKey", nodeInfo.valInfo.GetPublicKey())
				continue
			}
		}
	}
}

func createShardsMap(shardCoordinator sharding.Coordinator) map[uint32]struct{} {
	shardsMap := make(map[uint32]struct{}, shardCoordinator.NumberOfShards())
	for shardID := uint32(0); shardID < shardCoordinator.NumberOfShards(); shardID++ {
		shardsMap[shardID] = struct{}{}
	}
	shardsMap[core.MetachainShardId] = struct{}{}
	return shardsMap
}
