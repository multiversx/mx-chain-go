package metachain

import (
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.EpochStartRewardsCreator = (*rewardsCreatorV2)(nil)

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

// NewEpochStartRewardsCreator creates a new rewards creator object
func NewEpochStartRewardsCreatorV2(args RewardsCreatorArgsV2) (*rewardsCreatorV2, error) {
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
	if big.NewInt(0).Cmp(args.TopUpGradientPoint) >= 0 {
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
func (rc *rewardsCreatorV2) CreateRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) (block.MiniBlockSlice, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}

	rc.mutRewardsData.Lock()
	defer rc.mutRewardsData.Unlock()

	miniBlocks := rc.initializeRewardsMiniBlocks()

	err := rc.prepareRewardsData(metaBlock, validatorsInfo)
	if err != nil {
		return nil, err
	}

	protRwdTx, protRwdShardId, err := rc.createProtocolSustainabilityRewardTransaction(metaBlock)
	if err != nil {
		return nil, err
	}

	rewardsPerNode, dustFromRewardsPerNode := rc.computeRewardsPerNode(validatorsInfo)
	log.Debug("arithmetic difference from dust rewards per node", "value", dustFromRewardsPerNode.String())

	dust, err := rc.addValidatorRewardsToMiniBlocks(
		validatorsInfo, metaBlock, miniBlocks, rewardsPerNode,
	)
	if err != nil {
		return nil, err
	}

	dust.Add(dust, dustFromRewardsPerNode)
	rc.adjustProtocolSustainabilityRewards(protRwdTx, dust)
	err = rc.addProtocolRewardToMiniBlocks(protRwdTx, miniBlocks, protRwdShardId)
	if err != nil {
		return nil, err
	}

	return rc.finalizeMiniBlocks(miniBlocks), nil
}

// VerifyRewardsMiniBlocks verifies if received rewards miniblocks are correct
func (rc *rewardsCreatorV2) VerifyRewardsMiniBlocks(metaBlock *block.MetaBlock, validatorsInfo map[uint32][]*state.ValidatorInfo) error {
	if check.IfNil(metaBlock) {
		return epochStart.ErrNilHeaderHandler
	}

	createdMiniBlocks, err := rc.CreateRewardsMiniBlocks(metaBlock, validatorsInfo)
	if err != nil {
		return err
	}

	return rc.verifyCreatedRewardMiniBlocksWithMetaBlock(metaBlock, createdMiniBlocks)
}

func (rc *rewardsCreatorV2) addValidatorRewardsToMiniBlocks(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	metaBlock *block.MetaBlock,
	miniBlocks block.MiniBlockSlice,
	rewardsPerNode map[uint32][]*big.Int,
) (*big.Int, error) {
	rwdAddrValidatorInfo, accumulatedDust := rc.computeValidatorInfoPerRewardAddress(validatorsInfo, rewardsPerNode)

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
				accumulatedDust.Add(accumulatedDust, rwdTx.GetValue())
				continue
			}
		}

		if rwdTx.Value.Cmp(big.NewInt(0)) < 0 {
			log.Error("negative rewards", "rcv", rwdTx.RcvAddr)
			continue
		}
		rc.currTxs.AddTx(rwdTxHash, rwdTx)
		miniBlocks[mbId].TxHashes = append(miniBlocks[mbId].TxHashes, rwdTxHash)
	}

	return accumulatedDust, nil
}

func (rc *rewardsCreatorV2) computeValidatorInfoPerRewardAddress(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	rewardsPerNode map[uint32][]*big.Int,
) (map[string]*rewardInfoData, *big.Int) {

	rwdAddrValidatorInfo := make(map[string]*rewardInfoData)
	accumulatedDust := big.NewInt(0)

	for shardID, shardValidatorsInfo := range validatorsInfo {
		for i, validatorInfo := range shardValidatorsInfo {
			if validatorInfo.LeaderSuccess == 0 && validatorInfo.ValidatorSuccess == 0 {
				accumulatedDust.Add(accumulatedDust, rewardsPerNode[shardID][i])
				continue
			}

			rwdInfo, ok := rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)]
			if !ok {
				rwdInfo = &rewardInfoData{
					accumulatedFees: big.NewInt(0),
					protocolRewards: big.NewInt(0),
					address:         string(validatorInfo.RewardAddress),
				}
				rwdAddrValidatorInfo[string(validatorInfo.RewardAddress)] = rwdInfo
			}

			rwdInfo.accumulatedFees.Add(rwdInfo.accumulatedFees, validatorInfo.AccumulatedFees)
			rwdInfo.protocolRewards.Add(rwdInfo.protocolRewards, rewardsPerNode[shardID][i])
		}
	}

	return rwdAddrValidatorInfo, accumulatedDust
}

// IsInterfaceNil return true if underlying object is nil
func (rc *rewardsCreatorV2) IsInterfaceNil() bool {
	return rc == nil
}

func (rc *rewardsCreatorV2) prepareStakingData(eligibleNodesKeys map[uint32][][]byte) error {
	sw := core.NewStopWatch()
	sw.Start("prepareRewardsFromStakingSC")
	defer func() {
		sw.Stop("prepareRewardsFromStakingSC")
		log.Debug("rewardsCreator.prepareRewardsFromStakingSC time measurements", sw.GetMeasurements())
	}()

	return rc.stakingDataProvider.PrepareStakingData(eligibleNodesKeys)
}

func (rc *rewardsCreatorV2) getEligibleNodesKeyMap(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) map[uint32][][]byte {

	eligibleNodesKeys := make(map[uint32][][]byte)
	for shardID, validatorsInfoSlice := range validatorsInfo {
		eligibleNodesKeys[shardID] = make([][]byte, len(validatorsInfoSlice))
		for i, validatorInfo := range validatorsInfoSlice {
			eligibleNodesKeys[shardID][i] = validatorInfo.PublicKey
		}
	}

	return eligibleNodesKeys
}

func (rc *rewardsCreatorV2) getEligibleNodesStakeTopUp(eligibleNodesKeys map[uint32][][]byte) (map[uint32][]*big.Int, error) {
	stakeTopUpPerNode := make(map[uint32][]*big.Int)

	for shardID, shardKeys := range eligibleNodesKeys {
		stakeTopUpPerNode[shardID] = make([]*big.Int, len(shardKeys))
		for i, key := range shardKeys {
			topUp, err := rc.stakingDataProvider.GetNodeStakedTopUp(key)
			if err != nil {
				return nil, err
			}
			stakeTopUpPerNode[shardID][i] = topUp
		}
	}

	return stakeTopUpPerNode, nil
}

func (rc *rewardsCreatorV2) computeRewardsPerNode(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) (map[uint32][]*big.Int, *big.Int) {

	// totalTopUpEligible is the cumulative top-up stake value for eligible nodes
	totalTopUpEligible := rc.stakingDataProvider.GetTotalTopUpStakeEligibleNodes()
	remainingToBeDistributed := rc.economicsDataProvider.RewardsToBeDistributedForBlocks()
	topUpRewards := rc.computeTopUpRewards(remainingToBeDistributed, totalTopUpEligible)
	baseRewards := big.NewInt(0).Sub(remainingToBeDistributed, topUpRewards)
	nbBlocks := big.NewInt(int64(rc.economicsDataProvider.NumberOfBlocks()))
	baseRewardsPerBlock := big.NewInt(0).Div(baseRewards, nbBlocks)
	rc.fillBaseRewardsPerBlockPerNode(baseRewardsPerBlock)

	accumulatedDust := big.NewInt(0)
	baseRewardsPerNode, dust := rc.computeBaseRewardsPerNode(validatorsInfo, baseRewards)
	accumulatedDust.Add(accumulatedDust, dust)
	topUpRewardsPerNode, dust := rc.computeTopUpRewardsPerNode(topUpRewards, validatorsInfo)
	accumulatedDust.Add(accumulatedDust, dust)

	return aggregateBaseAndTopUpRewardsPerNode(baseRewardsPerNode, topUpRewardsPerNode), accumulatedDust
}

func (rc *rewardsCreatorV2) computeBaseRewardsPerNode(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	baseRewards *big.Int,
) (map[uint32][]*big.Int, *big.Int) {
	baseRewardsPerNode := make(map[uint32][]*big.Int)
	accumulatedRewards := big.NewInt(0)

	for shardID, valInfoList := range validatorsInfo {
		baseRewardsPerNode[shardID] = make([]*big.Int, len(valInfoList))
		for i, valInfo := range valInfoList {
			baseRewardsPerNode[shardID][i] = big.NewInt(0).Mul(
				rc.mapBaseRewardsPerBlockPerValidator[shardID],
				big.NewInt(0).SetInt64(int64(valInfo.NumSelectedInSuccessBlocks)))
			accumulatedRewards.Add(accumulatedRewards, baseRewardsPerNode[shardID][i])
		}
	}

	return baseRewardsPerNode, big.NewInt(0).Sub(baseRewards, accumulatedRewards)
}

func (rc *rewardsCreatorV2) computeTopUpRewardsPerNode(
	topUpRewards *big.Int,
	nodesInfo map[uint32][]*state.ValidatorInfo,
) (map[uint32][]*big.Int, *big.Int) {

	topUpRewardsPerNode := make(map[uint32][]*big.Int)
	accumulatedTopUpRewards := big.NewInt(0)
	topUpPerNode := rc.getTopUpForAllEligibleNodes(nodesInfo)
	topUpRewardPerShard := rc.computeTopUpRewardsPerShard(topUpRewards, topUpPerNode)
	totalPowerInShard, nodesPower := computeNodesPowerInShard(nodesInfo, topUpPerNode)

	// topUpRewardPerNodeInShardX = nodePowerInShardX*topUpRewardsShardX/totalPowerInShardX
	for shardID, nodesPowerList := range nodesPower {
		topUpRewardsPerNode[shardID] = make([]*big.Int, len(nodesPowerList))
		for i, nodePower := range nodesPowerList {
			if totalPowerInShard[shardID].Cmp(big.NewInt(0)) == 0 {
				// avoid division by zero
				log.Warn("rewardsCreatorV2.computeTopUpRewardsPerNode",
					"error", "shardPower zero",
					"shardID", shardID)
				topUpRewardsPerNode[shardID][i] = big.NewInt(0)

				continue
			}

			topUpRewardsPerNode[shardID][i] = big.NewInt(0).Mul(nodePower, topUpRewardPerShard[shardID])
			topUpRewardsPerNode[shardID][i].Div(topUpRewardsPerNode[shardID][i], totalPowerInShard[shardID])
			accumulatedTopUpRewards.Add(accumulatedTopUpRewards, topUpRewardsPerNode[shardID][i])
		}
	}

	return topUpRewardsPerNode, big.NewInt(0).Sub(topUpRewards, accumulatedTopUpRewards)
}

//      (2*k/pi)*atan(x/p), where:
//     k is the rewards per day limit for top-up stake k = c * economics.TotalToDistribute, c - constant, e.g c = 0.25
//     x is the cumulative top-up stake value for eligible nodes
//     p is the cumulative eligible stake where rewards per day reach 1/2 of k (includes topUp for the eligible nodes)
//     pi is the mathematical constant pi = 3.1415...
func (rc *rewardsCreatorV2) computeTopUpRewards(totalToDistribute *big.Int, totalTopUpEligible *big.Int) *big.Int {
	zero := big.NewInt(0)
	if totalToDistribute.Cmp(zero) <= 0 || totalTopUpEligible.Cmp(zero) <= 0 {
		return zero
	}

	// k = c * economics.TotalToDistribute, c = top-up reward factor (constant)
	k := core.GetPercentageOfValue(totalToDistribute, rc.topUpRewardFactor)
	// p is the cumulative eligible stake where rewards per day reach 1/2 of k (constant)
	// x/p - argument for atan
	totalTopUpEligibleFloat := new(big.Float).SetInt(totalTopUpEligible)
	topUpGradientPointFloat := new(big.Float).SetInt(rc.topUpGradientPoint)

	floatArg, _ := big.NewFloat(0).Quo(totalTopUpEligibleFloat, topUpGradientPointFloat).Float64()
	// atan(x/p)
	res1 := math.Atan(floatArg)
	// 2*k/pi
	res2 := new(big.Float).SetInt(big.NewInt(0).Mul(k, big.NewInt(2)))
	res2 = new(big.Float).Quo(res2, big.NewFloat(math.Pi))
	// topUpReward:= (2*k/pi)*atan(x/p)
	topUpRewards, _ := new(big.Float).Mul(big.NewFloat(res1), res2).Int(nil)

	return topUpRewards
}

// top-Up rewards are distributed to shard nodes, depending on the top-up ratio and the number of blocks
// ratio in the shard, with respect to the entire network
func (rc *rewardsCreatorV2) computeTopUpRewardsPerShard(
	topUpRewards *big.Int,
	stakeTopUpPerNode map[uint32][]*big.Int,
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
	shardsTopUp := computeTopUpPerShard(stakeTopUpPerNode)
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

func computeTopUpPerShard(stakeTopUpPerNode map[uint32][]*big.Int) map[uint32]*big.Int {
	shardsTopUp := make(map[uint32]*big.Int)
	for shardID, nodesTopUpList := range stakeTopUpPerNode {
		shardsTopUp[shardID] = big.NewInt(0)
		for _, nodeTopUp := range nodesTopUpList {
			shardsTopUp[shardID].Add(shardsTopUp[shardID], nodeTopUp)
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

	if totalPower.Cmp(big.NewInt(0)) <= 0 {
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
	nodesInfo map[uint32][]*state.ValidatorInfo,
	nodesTopUp map[uint32][]*big.Int,
) (map[uint32]*big.Int, map[uint32][]*big.Int) {
	totalShardNodesPower := make(map[uint32]*big.Int)
	nodesPower := make(map[uint32][]*big.Int)

	for shardID, nodeList := range nodesInfo {
		nodesPower[shardID] = make([]*big.Int, len(nodeList))
		totalShardNodesPower[shardID] = big.NewInt(0)
		for i, nodeInfo := range nodeList {
			nodesPower[shardID][i] = computeNodePowerInShard(nodeInfo, nodesTopUp[shardID][i])
			totalShardNodesPower[shardID].Add(totalShardNodesPower[shardID], nodesPower[shardID][i])
		}
	}

	return totalShardNodesPower, nodesPower
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

func aggregateBaseAndTopUpRewardsPerNode(baseRewards, topUpRewards map[uint32][]*big.Int) map[uint32][]*big.Int {
	fullRewards := make(map[uint32][]*big.Int)
	for shardID, rewardList := range topUpRewards {
		fullRewards[shardID] = make([]*big.Int, len(rewardList))
		for i, topUpReward := range rewardList {
			fullRewards[shardID][i] = big.NewInt(0).Add(
				baseRewards[shardID][i],
				topUpReward)
		}
	}

	return fullRewards
}

func (rc *rewardsCreatorV2) getTopUpForAllEligibleNodes(
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) map[uint32][]*big.Int {
	var err error
	validatorsTopUp := make(map[uint32][]*big.Int)
	for shardID, nodeListInfo := range validatorsInfo {
		validatorsTopUp[shardID] = make([]*big.Int, len(nodeListInfo))
		for i, nodeInfo := range nodeListInfo {
			validatorsTopUp[shardID][i], err = rc.stakingDataProvider.GetNodeStakedTopUp(nodeInfo.GetPublicKey())
			if err != nil {
				validatorsTopUp[shardID][i] = big.NewInt(0)

				log.Error("getTopUpForAllEligible",
					"error", err.Error(),
					"blsKey", nodeInfo.GetPublicKey())
				continue
			}
		}
	}

	return validatorsTopUp
}

func (rc *rewardsCreatorV2) prepareRewardsData(
	metaBlock *block.MetaBlock,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
) error {
	rc.clean()
	rc.flagDelegationSystemSCEnabled.Toggle(metaBlock.GetEpoch() >= rc.delegationSystemSCEnableEpoch)
	eligibleNodesKeys := rc.getEligibleNodesKeyMap(validatorsInfo)
	err := rc.prepareStakingData(eligibleNodesKeys)
	if err != nil {
		return err
	}

	return nil
}

func createShardsMap(shardCoordinator sharding.Coordinator) map[uint32]struct{} {
	shardsMap := make(map[uint32]struct{}, shardCoordinator.NumberOfShards())
	for shardID := uint32(0); shardID < shardCoordinator.NumberOfShards(); shardID++ {
		shardsMap[shardID] = struct{}{}
	}
	shardsMap[core.MetachainShardId] = struct{}{}
	return shardsMap
}
