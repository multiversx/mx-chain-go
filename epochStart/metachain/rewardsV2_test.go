package metachain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/require"
)

const (
	tuStake   = "topUpStake"
	tuRewards = "topUpRewards"
	bRewards  = "baseRewards"
	fRewards  = "fullRewards"
)

const defaultBlocksPerShard = uint32(14400)

// SetupRewardsResult -
type SetupRewardsResult struct {
	RewardsCreatorArgsV2
	smallestDivisionEGLD  *big.Int
	totalRewardsFirstYear *big.Int
	epochsInYear          *big.Int
	rewardsForBlocks      *big.Int
	baseStake             *big.Int
}

func TestNewRewardsCreator_InvalidBaseRewardsCreatorArgs(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.BaseRewardsCreatorArgs.NodesConfigProvider = nil

	rwd, err := NewRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrNilNodesConfigProvider, err)
}

func TestNewRewardsCreator_NilStakingDataProvider(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.StakingDataProvider = nil

	rwd, err := NewRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrNilStakingDataProvider, err)
}

func TestNewRewardsCreator_NilEconomicsDataProvider(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.EconomicsDataProvider = nil

	rwd, err := NewRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrNilEconomicsDataProvider, err)
}

func TestNewRewardsCreator_NilRewardsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.RewardsHandler = nil

	rwd, err := NewRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrNilRewardsHandler, err)
}

func TestNewRewardsCreatorOK(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)
}

func TestNewRewardsCreatorV2_initNodesRewardsInfo(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	valInfoEligible := createDefaultValidatorInfo(400, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	valInfoEligibleWithExtra := addNonEligibleValidatorInfo(100, valInfoEligible, string(common.WaitingList))

	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfoEligibleWithExtra)
	require.Equal(t, len(valInfoEligible), len(nodesRewardInfo))

	for shardID, nodeInfoList := range nodesRewardInfo {
		require.Equal(t, len(nodeInfoList), len(valInfoEligible[shardID]))
		for i, nodeInfo := range nodeInfoList {
			require.True(t, valInfoEligible[shardID][i] == nodeInfo.valInfo)
			require.Equal(t, zero, nodeInfo.topUpStake)
			require.Equal(t, zero, nodeInfo.powerInShard)
			require.Equal(t, zero, nodeInfo.baseReward)
			require.Equal(t, zero, nodeInfo.topUpReward)
			require.Equal(t, zero, nodeInfo.fullRewards)
		}
	}
}

func TestNewRewardsCreatorV2_getTopUpForAllEligibleNodes(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	topUpVal, _ := big.NewInt(0).SetString("100000000000000000000", 10)
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			topUp := big.NewInt(0).Set(topUpVal)
			return topUp, nil
		},
	}
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)

	rwd.getTopUpForAllEligibleNodes(nodesRewardInfo)
	for _, nodesInfoList := range nodesRewardInfo {
		require.Equal(t, nodesPerShard, uint32(len(nodesInfoList)))
		for _, nodeInfo := range nodesInfoList {
			require.Equal(t, topUpVal, nodeInfo.topUpStake)
		}
	}
}

func TestNewRewardsCreatorV2_getTopUpForAllEligibleSomeBLSKeysNotFoundZeroed(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	topUpVal, _ := big.NewInt(0).SetString("100000000000000000000", 10)
	notFoundKey := []byte("notFound")
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			if bytes.Equal(blsKey, notFoundKey) {
				return nil, fmt.Errorf("not found")
			}
			topUp := big.NewInt(0).Set(topUpVal)
			return topUp, nil
		},
	}
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(10)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	for _, valList := range valInfo {
		valList[0].PublicKey = notFoundKey
		valList[1].PublicKey = notFoundKey
	}
	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)

	rwd.getTopUpForAllEligibleNodes(nodesRewardInfo)
	for _, nodeInfoList := range nodesRewardInfo {
		require.Equal(t, nodesPerShard, uint32(len(nodeInfoList)))
		for i, nodeInfo := range nodeInfoList {
			if i < 2 {
				require.Equal(t, big.NewInt(0), nodeInfo.topUpStake)
			} else {
				require.Equal(t, topUpVal, nodeInfo.topUpStake)
			}
		}
	}
}

func TestRewardsCreatorV2_adjustProtocolSustainabilityRewardsPositiveValue(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	initialProtRewardValue := big.NewInt(1000000)
	protRwAddr, _ := args.PubkeyConverter.Decode(args.ProtocolSustainabilityAddress)
	protRwTx := &rewardTx.RewardTx{
		Round:   100,
		Value:   big.NewInt(0).Set(initialProtRewardValue),
		RcvAddr: protRwAddr,
		Epoch:   1,
	}

	protRwShard := args.ShardCoordinator.ComputeId(protRwAddr)
	mbSlice := createDefaultMiniBlocksSlice()
	_ = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	dust := big.NewInt(1000)
	rwd2 := rewardsCreatorV2{
		baseRewardsCreator: rwd,
	}
	rwd2.adjustProtocolSustainabilityRewards(protRwTx, dust)
	require.Zero(t, protRwTx.Value.Cmp(big.NewInt(0).Add(dust, initialProtRewardValue)))
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestRewardsCreatorV2_adjustProtocolSustainabilityRewardsNegValueNotAccepted(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	initialProtRewardValue := big.NewInt(10)
	protRwAddr, _ := args.PubkeyConverter.Decode(args.ProtocolSustainabilityAddress)
	protRwTx := &rewardTx.RewardTx{
		Round:   100,
		Value:   big.NewInt(0).Set(initialProtRewardValue),
		RcvAddr: protRwAddr,
		Epoch:   1,
	}

	protRwShard := args.ShardCoordinator.ComputeId(protRwAddr)
	mbSlice := createDefaultMiniBlocksSlice()
	_ = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	rwd2 := rewardsCreatorV2{
		baseRewardsCreator: rwd,
	}

	dust := big.NewInt(-10)
	rwd2.adjustProtocolSustainabilityRewards(protRwTx, dust)
	require.Zero(t, protRwTx.Value.Cmp(initialProtRewardValue))
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestRewardsCreatorV2_adjustProtocolSustainabilityRewardsInitialNegativeValue(t *testing.T) {
	t.Parallel()

	args := getBaseRewardsArguments()
	rwd, err := NewBaseRewardsCreator(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	initialProtRewardValue := big.NewInt(-100)
	protRwAddr, _ := args.PubkeyConverter.Decode(args.ProtocolSustainabilityAddress)
	protRwTx := &rewardTx.RewardTx{
		Round:   100,
		Value:   big.NewInt(0).Set(initialProtRewardValue),
		RcvAddr: protRwAddr,
		Epoch:   1,
	}

	protRwShard := args.ShardCoordinator.ComputeId(protRwAddr)
	mbSlice := createDefaultMiniBlocksSlice()
	_ = rwd.addProtocolRewardToMiniBlocks(protRwTx, mbSlice, protRwShard)

	rwd2 := rewardsCreatorV2{
		baseRewardsCreator: rwd,
	}

	dust := big.NewInt(0)
	rwd2.adjustProtocolSustainabilityRewards(protRwTx, dust)
	require.Zero(t, protRwTx.Value.Cmp(big.NewInt(0)))
	setProtValue := rwd.GetProtocolSustainabilityRewards()
	require.Zero(t, protRwTx.Value.Cmp(setProtValue))
}

func TestNewRewardsCreatorV2_aggregateBaseAndTopUpRewardsPerNode(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(10)
	validatorsInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	nodesRewardInfo := rwd.initNodesRewardsInfo(validatorsInfo)

	value := 1000 + nodesPerShard
	shardMap := createShardsMap(args.ShardCoordinator)
	for shardID := range shardMap {
		for i := uint32(0); i < nodesPerShard; i++ {
			nodesRewardInfo[shardID][i].baseReward = big.NewInt(int64(value - i))
			nodesRewardInfo[shardID][i].topUpReward = big.NewInt(int64(value + i))
		}
	}

	aggregateBaseAndTopUpRewardsPerNode(nodesRewardInfo)

	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			require.Equal(t, big.NewInt(2*int64(value)), nodeInfo.fullRewards)
		}
	}
}

func TestNewRewardsCreatorV2_computeNodePowerInShardOfflineNodeZeroPower(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	leaderSuccess := uint32(0)
	nbBlocksSelected := uint32(6000)
	proposerFee := 100

	valInfo := &state.ValidatorInfo{
		PublicKey:                  nil,
		ShardId:                    0,
		RewardAddress:              []byte("rewardAddr"),
		LeaderSuccess:              leaderSuccess,
		ValidatorSuccess:           0,
		NumSelectedInSuccessBlocks: nbBlocksSelected,
		AccumulatedFees:            big.NewInt(int64(proposerFee)),
	}

	nodeTopUp := big.NewInt(1000000)
	power := computeNodePowerInShard(valInfo, nodeTopUp)
	require.Equal(t, big.NewInt(0), power)
}

func TestNewRewardsCreatorV2_computeNodePowerInShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	leaderSuccess := uint32(100)
	nbBlocksSelected := uint32(6000)
	proposerFee := 100

	valInfo := &state.ValidatorInfo{
		PublicKey:                  nil,
		ShardId:                    0,
		RewardAddress:              []byte("rewardAddr"),
		LeaderSuccess:              leaderSuccess,
		ValidatorSuccess:           nbBlocksSelected - leaderSuccess,
		NumSelectedInSuccessBlocks: nbBlocksSelected,
		AccumulatedFees:            big.NewInt(int64(proposerFee)),
	}

	nodeTopUp := big.NewInt(1000000)
	expectedPower := big.NewInt(0).Mul(nodeTopUp, big.NewInt(0).SetInt64(int64(nbBlocksSelected)))
	power := computeNodePowerInShard(valInfo, nodeTopUp)
	require.Equal(t, expectedPower, power)
}

func TestNewRewardsCreatorV2_computeNodesPowerInShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)
	_, _ = setDummyValuesInNodesRewardInfo(nodesRewardInfo, nodesPerShard, tuStake, 0)

	nodePowerInShard := computeNodesPowerInShard(nodesRewardInfo)
	require.NotNil(t, nodePowerInShard)

	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			blocks := nodeInfo.valInfo.NumSelectedInSuccessBlocks
			topUp := nodeInfo.topUpStake
			require.Equal(t, big.NewInt(0).Mul(big.NewInt(int64(blocks)), topUp), nodeInfo.powerInShard)
		}
	}
}

func TestNewRewardsCreatorV2_computeShardsPower(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	topUpPerShard := make(map[uint32]*big.Int)
	topUpPerShard[0] = big.NewInt(1000000)
	topUpPerShard[1] = big.NewInt(1200000)
	topUpPerShard[core.MetachainShardId] = big.NewInt(1500000)

	blocksPerShard := make(map[uint32]uint64)
	blocksPerShard[0] = 1000
	blocksPerShard[1] = 800
	blocksPerShard[core.MetachainShardId] = 1300

	totalPower, powerPerShard := computeShardsPower(topUpPerShard, blocksPerShard)
	expectedTotalPower := big.NewInt(1000000*1000 + 1200000*800 + 1500000*1300)
	require.Equal(t, expectedTotalPower, totalPower)

	for shardID, power := range powerPerShard {
		expectedShardPower := big.NewInt(topUpPerShard[shardID].Int64() * int64(blocksPerShard[shardID]))
		require.Equal(t, expectedShardPower, power)
	}
}

func TestNewRewardsCreatorV2_computeRewardsForPowerPerShardZeroTotalPower(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	powerPerShards := make(map[uint32]*big.Int)
	totalPower := big.NewInt(0)
	topUpRewards := big.NewInt(3000)

	rewardsPerShard := computeRewardsForPowerPerShard(powerPerShards, totalPower, topUpRewards)

	for _, reward := range rewardsPerShard {
		require.Equal(t, big.NewInt(0), reward)
	}
}

func TestNewRewardsCreatorV2_computeRewardsForPowerPerShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	powerPerShards := make(map[uint32]*big.Int)
	powerPerShards[0] = big.NewInt(1000)
	powerPerShards[1] = big.NewInt(1200)
	powerPerShards[core.MetachainShardId] = big.NewInt(1500)

	totalPower := big.NewInt(3700)
	topUpRewards := big.NewInt(3000)

	rewardsPerShard := computeRewardsForPowerPerShard(powerPerShards, totalPower, topUpRewards)

	for shardID, reward := range rewardsPerShard {
		expectedReward := big.NewInt(0).Mul(powerPerShards[shardID], topUpRewards)
		expectedReward.Div(expectedReward, totalPower)
		require.Equal(t, expectedReward, reward)
	}
}

func TestNewRewardsCreatorV2_computeTopUpRewardsPerShardNoBlockPerShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)
	_, _ = setDummyValuesInNodesRewardInfo(nodesRewardInfo, nodesPerShard, tuStake, 0)

	topUpRewards := big.NewInt(3000)
	rwd.economicsDataProvider.SetNumberOfBlocksPerShard(nil)

	topUpRewardsPerShard := rwd.computeTopUpRewardsPerShard(topUpRewards, nodesRewardInfo)

	for _, shardTopUpRewards := range topUpRewardsPerShard {
		require.Equal(t, big.NewInt(0), shardTopUpRewards)
	}
}

func TestNewRewardsCreatorV2_computeTopUpRewardsPerShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)
	topUpPerShard, _ := setDummyValuesInNodesRewardInfo(nodesRewardInfo, nodesPerShard, tuStake, 0)

	blocksPerShard := make(map[uint32]uint64)
	blocksPerShard[0] = 2000
	blocksPerShard[1] = 1200
	blocksPerShard[core.MetachainShardId] = 1500

	topUpRewards := big.NewInt(3000)
	rwd.economicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)

	topUpRewardsPerShard := rwd.computeTopUpRewardsPerShard(topUpRewards, nodesRewardInfo)

	totalPower := big.NewInt(0)
	expectedRewardsPerShard := make(map[uint32]*big.Int)

	for shardID := range topUpRewardsPerShard {
		expectedPower := big.NewInt(0).Mul(topUpPerShard[shardID], big.NewInt(int64(blocksPerShard[shardID])))
		totalPower.Add(totalPower, expectedPower)
		expectedRewardsPerShard[shardID] = big.NewInt(0).Mul(expectedPower, topUpRewards)
	}

	for shardID, shardTopUpRewards := range topUpRewardsPerShard {
		expectedRewardsPerShard[shardID].Div(expectedRewardsPerShard[shardID], totalPower)
		require.Equal(t, expectedRewardsPerShard[shardID], shardTopUpRewards)
	}
}

func TestNewRewardsCreatorV2_computeTopUpRewardsZeroTopup(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	totalToDistribute, _ := big.NewInt(-1).SetString("3000000000000000000000", 10)
	totalTopUpEligible := big.NewInt(0)

	topUpRewards := rwd.computeTopUpRewards(totalToDistribute, totalTopUpEligible)
	require.NotNil(t, topUpRewards)
	require.Equal(t, big.NewInt(0), topUpRewards)
}

func TestNewRewardsCreatorV2_computeTopUpRewardsNegativeToDistribute(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	totalToDistribute := big.NewInt(-1)
	totalTopUpEligible := big.NewInt(10000)

	topUpRewards := rwd.computeTopUpRewards(totalToDistribute, totalTopUpEligible)
	require.NotNil(t, topUpRewards)
	require.Equal(t, big.NewInt(0), topUpRewards)
}

func TestNewRewardsCreatorV2_computeTopUpRewards(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	totalToDistribute, _ := big.NewInt(0).SetString("3000000000000000000000", 10)
	topUpRewardsLimit := core.GetApproximatePercentageOfValue(totalToDistribute, rwd.rewardsHandler.RewardsTopUpFactor())

	totalTopUpEligible, _ := big.NewInt(0).SetString("2000000000000000000000000", 10)
	topUpRewards := rwd.computeTopUpRewards(totalToDistribute, totalTopUpEligible)
	require.NotNil(t, topUpRewards)
	require.True(t, topUpRewards.Cmp(topUpRewardsLimit) < 0)

	totalTopUpEligible, _ = big.NewInt(0).SetString("3000000000000000000000000", 10)
	topUpRewards = rwd.computeTopUpRewards(totalToDistribute, totalTopUpEligible)

	require.NotNil(t, topUpRewards)
	require.Equal(t, topUpRewards, big.NewInt(0).Div(topUpRewardsLimit, big.NewInt(2)))

	totalTopUpEligible, _ = big.NewInt(0).SetString("3500000000000000000000000", 10)
	topUpRewards = rwd.computeTopUpRewards(totalToDistribute, totalTopUpEligible)

	require.NotNil(t, topUpRewards)
	require.True(t, topUpRewards.Cmp(big.NewInt(0).Div(topUpRewardsLimit, big.NewInt(2))) > 0)

	totalTopUpEligible, _ = big.NewInt(0).SetString("9000000000000000000000000000000", 10)
	topUpRewards = rwd.computeTopUpRewards(totalToDistribute, totalTopUpEligible)

	require.NotNil(t, topUpRewards)
	require.True(t, topUpRewards.Cmp(big.NewInt(0).Div(topUpRewardsLimit, big.NewInt(2))) > 0)
	require.True(t, topUpRewards.Cmp(topUpRewardsLimit) < 0)
	ninetyNinePercentLimit := core.GetApproximatePercentageOfValue(topUpRewardsLimit, 0.99)
	require.True(t, topUpRewards.Cmp(ninetyNinePercentLimit) > 0)
}

func TestNewRewardsCreatorV2_computeTopUpRewardsPerNode(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	dummyRwd, _ := NewRewardsCreatorV2(args)
	nodesRewardInfo := dummyRwd.initNodesRewardsInfo(vInfo)
	_, _ = setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, tuStake, 0)

	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo {
				for i, v := range vList {
					if bytes.Equal(v.PublicKey, blsKey) {
						return nodesRewardInfo[shardID][i].topUpStake, nil
					}
				}
			}
			return nil, fmt.Errorf("not found")
		},
	}
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	topUpRewards, _ := big.NewInt(0).SetString("700000000000000000000", 10)
	blocksPerShard := make(map[uint32]uint64)
	blocksPerShard[0] = 2000
	blocksPerShard[1] = 1200
	blocksPerShard[core.MetachainShardId] = 1500
	rwd.economicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)

	dust := rwd.computeTopUpRewardsPerNode(nodesRewardInfo, topUpRewards)

	sumTopUpRewards := big.NewInt(0)
	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			sumTopUpRewards.Add(sumTopUpRewards, nodeInfo.topUpReward)
		}
	}

	// dust should be really small, checking against  topupRewards/1mil
	limit := core.GetApproximatePercentageOfValue(topUpRewards, 0.000001)
	require.True(t, limit.Cmp(dust) > 0)

	sumTopUpRewards.Add(sumTopUpRewards, dust)
	require.Equal(t, topUpRewards, sumTopUpRewards)
}

func TestNewRewardsCreatorV2_computeTopUpRewardsPerNodeNotFoundBLSKeys(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesRewardInfo := rwd.initNodesRewardsInfo(vInfo)

	topUpRewards, _ := big.NewInt(0).SetString("700000000000000000000", 10)
	blocksPerShard := make(map[uint32]uint64)
	blocksPerShard[0] = 2000
	blocksPerShard[1] = 1200
	blocksPerShard[core.MetachainShardId] = 1500
	rwd.economicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)

	dust := rwd.computeTopUpRewardsPerNode(nodesRewardInfo, topUpRewards)

	sumTopUpRewards := big.NewInt(0)
	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			sumTopUpRewards.Add(sumTopUpRewards, nodeInfo.topUpReward)
		}
	}

	// in this case, all rewards accumulated in dust, when all retrieved BLS keys are invalid, or not found
	require.True(t, topUpRewards.Cmp(dust) == 0)

	sumTopUpRewards.Add(sumTopUpRewards, dust)
	require.Equal(t, topUpRewards, sumTopUpRewards)
}

func TestNewRewardsCreatorV2_computeBaseRewardsPerNode(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)

	baseRewards, _ := big.NewInt(0).SetString("3000000000000000000000", 10)
	shardMap := createShardsMap(args.ShardCoordinator)
	nbBlocksPerShard := big.NewInt(14400)
	nbBlocks := big.NewInt(0).Mul(nbBlocksPerShard, big.NewInt(3))
	baseRewardPerBlock := big.NewInt(0).Div(baseRewards, nbBlocks)

	for shardID := range shardMap {
		rwd.mapBaseRewardsPerBlockPerValidator[shardID] = big.NewInt(0).Set(baseRewardPerBlock)
		cnsSize := big.NewInt(0).SetInt64(int64(args.NodesConfigProvider.ConsensusGroupSize(shardID)))
		rwd.mapBaseRewardsPerBlockPerValidator[shardID].Div(rwd.mapBaseRewardsPerBlockPerValidator[shardID], cnsSize)
	}

	dust := rwd.computeBaseRewardsPerNode(nodesRewardInfo, baseRewards)

	// dust should be really small, checking against  baseRewards/1mil
	limit := core.GetApproximatePercentageOfValue(baseRewards, 0.000001)
	require.True(t, limit.Cmp(dust) > 0)

	sumRwds := big.NewInt(0)
	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			sumRwds.Add(sumRwds, nodeInfo.baseReward)
		}
	}

	require.Equal(t, big.NewInt(0).Add(sumRwds, dust), baseRewards)
}

func TestNewRewardsCreatorV2_computeRewardsPerNode(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	dummyRwd, _ := NewRewardsCreatorV2(args)
	nodesRewardInfo := dummyRwd.initNodesRewardsInfo(vInfo)
	_, totalTopUpStake := setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, tuStake, 0)

	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
			topUpStake := big.NewInt(0).Set(totalTopUpStake)
			return topUpStake
		},
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo {
				for i, v := range vList {
					if bytes.Equal(v.PublicKey, blsKey) {
						return nodesRewardInfo[shardID][i].topUpStake, nil
					}
				}
			}
			return nil, fmt.Errorf("not found")
		},
	}
	nbBlocksPerShard := uint64(14400)
	blocksPerShard := make(map[uint32]uint64)
	shardMap := createShardsMap(args.ShardCoordinator)
	for shardID := range shardMap {
		blocksPerShard[shardID] = nbBlocksPerShard
	}

	args.EconomicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)
	rewardsForBlocks, _ := big.NewInt(0).SetString("5000000000000000000000", 10)
	args.EconomicsDataProvider.SetRewardsToBeDistributedForBlocks(rewardsForBlocks)

	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesRewardInfo, accumulatedDust := rwd.computeRewardsPerNode(vInfo)

	// dust should be really small, checking against  baseRewards/1mil
	limit := core.GetApproximatePercentageOfValue(rewardsForBlocks, 0.000001)
	require.True(t, limit.Cmp(accumulatedDust) > 0)

	sumRwds := big.NewInt(0)
	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			sumRwds.Add(sumRwds, nodeInfo.fullRewards)
		}
	}

	require.Equal(t, rewardsForBlocks, big.NewInt(0).Add(sumRwds, accumulatedDust))
}

func TestNewRewardsCreatorV2_computeAverageRewardsPer2169Nodes(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	setupResult := setUpRewards(args)

	tests := []struct {
		topupStake  *big.Int
		expectedROI float64
	}{
		{
			topupStake:  big.NewInt(0),
			expectedROI: 36.007,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(500), setupResult.smallestDivisionEGLD),
			expectedROI: 30.01,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(1500), setupResult.smallestDivisionEGLD),
			expectedROI: 22.51,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(2500), setupResult.smallestDivisionEGLD),
			expectedROI: 18.00,
		},
	}

	nbEligiblePerShard := uint32(400)
	nbWaitingPerShard := uint32(142)
	eligiblePercentage := float64(nbEligiblePerShard) / float64(nbWaitingPerShard+nbEligiblePerShard)

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			verifyRewardsComputation(
				t,
				nbEligiblePerShard,
				setupResult,
				tt.topupStake,
				tt.expectedROI,
				eligiblePercentage,
			)
		})
	}
}

func TestNewRewardsCreatorV2_computeAverageRewardsPer1920Nodes(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	setupResult := setUpRewards(args)

	tests := []struct {
		topupStake  *big.Int
		expectedROI float64
	}{
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(25), big.NewInt(0).Div(setupResult.smallestDivisionEGLD, big.NewInt(100))),
			expectedROI: 40.67,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(500), setupResult.smallestDivisionEGLD),
			expectedROI: 33.89,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(1500), setupResult.smallestDivisionEGLD),
			expectedROI: 25.41,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(2500), setupResult.smallestDivisionEGLD),
			expectedROI: 20.33,
		},
	}

	nbEligiblePerShard := uint32(400)
	nbWaitingPerShard := uint32(80)
	eligiblePercentage := float64(nbEligiblePerShard) / float64(nbWaitingPerShard+nbEligiblePerShard)

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			verifyRewardsComputation(
				t,
				nbEligiblePerShard,
				setupResult,
				tt.topupStake,
				tt.expectedROI,
				eligiblePercentage,
			)
		})
	}
}

func TestNewRewardsCreatorV2_computeAverageRewardsPer3200Nodes(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	setupResult := setUpRewards(args)

	tests := []struct {
		topupStake  *big.Int
		expectedROI float64
	}{
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(25), big.NewInt(0).Div(setupResult.smallestDivisionEGLD, big.NewInt(100))),
			expectedROI: 24.40,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(500), setupResult.smallestDivisionEGLD),
			expectedROI: 20.33,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(1500), setupResult.smallestDivisionEGLD),
			expectedROI: 15.25,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(2500), setupResult.smallestDivisionEGLD),
			expectedROI: 12.20,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(2*2500), setupResult.smallestDivisionEGLD),
			expectedROI: 8.13,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(3*2500), setupResult.smallestDivisionEGLD),
			expectedROI: 6.10,
		},
	}

	nbEligiblePerShard := uint32(400)
	nbWaitingPerShard := uint32(400)
	eligiblePercentage := float64(nbEligiblePerShard) / float64(nbWaitingPerShard+nbEligiblePerShard)

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			verifyRewardsComputation(
				t,
				nbEligiblePerShard,
				setupResult,
				tt.topupStake,
				tt.expectedROI,
				eligiblePercentage,
			)
		})
	}
}

func TestNewRewardsCreatorV35_computeAverageRewardsPer3200Nodes(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV35Arguments()
	setupResult := setUpRewards(args)

	tests := []struct {
		topupStake  *big.Int
		expectedROI float64
	}{
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(25), big.NewInt(0).Div(setupResult.smallestDivisionEGLD, big.NewInt(100))),
			expectedROI: 24.40,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(500), setupResult.smallestDivisionEGLD),
			expectedROI: 20.33,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(1500), setupResult.smallestDivisionEGLD),
			expectedROI: 15.25,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(2500), setupResult.smallestDivisionEGLD),
			expectedROI: 12.20,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(2*2500), setupResult.smallestDivisionEGLD),
			expectedROI: 8.13,
		},
		{
			topupStake:  big.NewInt(0).Mul(big.NewInt(3*2500), setupResult.smallestDivisionEGLD),
			expectedROI: 6.10,
		},
	}

	nbEligiblePerShard := uint32(400)
	nbWaitingPerShard := uint32(400)
	eligiblePercentage := float64(nbEligiblePerShard) / float64(nbWaitingPerShard+nbEligiblePerShard)

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			verifyRewardsComputation(
				t,
				nbEligiblePerShard,
				setupResult,
				tt.topupStake,
				tt.expectedROI,
				eligiblePercentage,
			)
		})
	}
}

// Test for 10.4 staked in the network with v3.5
func TestNewRewardsCreatorV35_computeRewardsPer3200NodesWithDifferentTopups(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV35Arguments()
	setupResult := setUpRewards(args)

	baseStakePerNode := big.NewInt(0).Mul(big.NewInt(2500), setupResult.smallestDivisionEGLD)
	topupStakePerNode := big.NewInt(0).Mul(big.NewInt(750), setupResult.smallestDivisionEGLD)

	baseEligibleStake := big.NewInt(0).Mul(big.NewInt(1600), baseStakePerNode)
	topupEligibleStake := big.NewInt(0).Mul(big.NewInt(1600), topupStakePerNode)

	baseAPR := 20.20
	topupAPR := 13.99

	tests := []struct {
		validatorTopupStake      *big.Int
		validatorExpectedFullROI float64
	}{
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(25), big.NewInt(0).Div(setupResult.smallestDivisionEGLD, big.NewInt(100))),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*0.25) / 2500.25,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*500) / 3000,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(1500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*1500) / 4000,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(2500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*2500) / 5000,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(2*2500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*5000) / 7500,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(3*2500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*7500) / 10000,
		},
	}

	nbEligiblePerShard := uint32(400)
	nbWaitingPerShard := uint32(400)

	eligiblePercentage := float64(nbEligiblePerShard) / float64(nbWaitingPerShard+nbEligiblePerShard)

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {

			vInfo := createDefaultValidatorInfo(nbEligiblePerShard, setupResult.ShardCoordinator, setupResult.NodesConfigProvider, 100, defaultBlocksPerShard)
			nodesRewardInfo, _ := setupNodeRewardInfo(setupResult, vInfo, topupStakePerNode, tt.validatorTopupStake)

			setupResult.EconomicsDataProvider.SetRewardsToBeDistributedForBlocks(setupResult.rewardsForBlocks)
			setupResult.RewardsCreatorArgsV2.StakingDataProvider = &mock.StakingDataProviderStub{
				GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
					return topupEligibleStake
				},
				GetTotalStakeEligibleNodesCalled: func() *big.Int {
					return baseEligibleStake
				},
				GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
					for shardID, vList := range vInfo {
						for i, v := range vList {
							if bytes.Equal(v.PublicKey, blsKey) {
								return nodesRewardInfo[shardID][i].topUpStake, nil
							}
						}
					}
					return nil, fmt.Errorf("not found")
				},
			}

			rwd, _ := NewRewardsCreatorV2(setupResult.RewardsCreatorArgsV2)
			nodesRewardInfo, _ = rwd.computeRewardsPerNode(vInfo)

			currentRewardInfo := nodesRewardInfo[0][0]

			log.Info("baseReward", "amount", currentRewardInfo.baseReward)

			verifyValidatorRewards(
				t,
				setupResult.epochsInYear,
				currentRewardInfo.baseReward,
				baseStakePerNode,
				baseAPR,
				eligiblePercentage,
			)
			verifyValidatorRewards(
				t,
				setupResult.epochsInYear,
				currentRewardInfo.topUpReward,
				currentRewardInfo.topUpStake,
				topupAPR,
				eligiblePercentage,
			)
			verifyValidatorRewards(t, setupResult.epochsInYear,
				currentRewardInfo.fullRewards,
				big.NewInt(0).Add(baseStakePerNode, currentRewardInfo.topUpStake),
				tt.validatorExpectedFullROI, eligiblePercentage)
		})
	}
}

// Test for 10.4 staked in the network with v2
func TestNewRewardsCreatorV2_computeRewardsPer3200NodesWithDifferentTopups(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	setupResult := setUpRewards(args)

	baseStakePerNode := big.NewInt(0).Mul(big.NewInt(2500), setupResult.smallestDivisionEGLD)
	topupStakePerNode := big.NewInt(0).Mul(big.NewInt(750), setupResult.smallestDivisionEGLD)

	baseEligibleStake := big.NewInt(0).Mul(big.NewInt(1600), baseStakePerNode)
	topupEligibleStake := big.NewInt(0).Mul(big.NewInt(1600), topupStakePerNode)

	baseAPR := 22.92
	topupAPR := 4.93

	tests := []struct {
		validatorTopupStake      *big.Int
		validatorExpectedFullROI float64
	}{
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(25), big.NewInt(0).Div(setupResult.smallestDivisionEGLD, big.NewInt(100))),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*0.25) / 2500.25,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*500) / 3000,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(1500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*1500) / 4000,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(2500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*2500) / 5000,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(2*2500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*5000) / 7500,
		},
		{
			validatorTopupStake:      big.NewInt(0).Mul(big.NewInt(3*2500), setupResult.smallestDivisionEGLD),
			validatorExpectedFullROI: (baseAPR*2500 + topupAPR*7500) / 10000,
		},
	}

	nbEligiblePerShard := uint32(400)
	nbWaitingPerShard := uint32(400)
	eligiblePercentage := float64(nbEligiblePerShard) / float64(nbWaitingPerShard+nbEligiblePerShard)

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {

			vInfo := createDefaultValidatorInfo(nbEligiblePerShard, setupResult.ShardCoordinator, setupResult.NodesConfigProvider, 100, defaultBlocksPerShard)
			nodesRewardInfo, _ := setupNodeRewardInfo(setupResult, vInfo, topupStakePerNode, tt.validatorTopupStake)

			setupResult.EconomicsDataProvider.SetRewardsToBeDistributedForBlocks(setupResult.rewardsForBlocks)
			setupResult.RewardsCreatorArgsV2.StakingDataProvider = &mock.StakingDataProviderStub{
				GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
					return topupEligibleStake
				},
				GetTotalStakeEligibleNodesCalled: func() *big.Int {
					return baseEligibleStake
				},
				GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
					for shardID, vList := range vInfo {
						for i, v := range vList {
							if bytes.Equal(v.PublicKey, blsKey) {
								return nodesRewardInfo[shardID][i].topUpStake, nil
							}
						}
					}
					return nil, fmt.Errorf("not found")
				},
			}

			rwd, _ := NewRewardsCreatorV2(setupResult.RewardsCreatorArgsV2)
			nodesRewardInfo, _ = rwd.computeRewardsPerNode(vInfo)

			currentRewardInfo := nodesRewardInfo[0][0]

			verifyValidatorRewards(
				t,
				setupResult.epochsInYear,
				currentRewardInfo.baseReward,
				baseStakePerNode,
				baseAPR,
				eligiblePercentage)
			verifyValidatorRewards(
				t,
				setupResult.epochsInYear,
				currentRewardInfo.topUpReward,
				currentRewardInfo.topUpStake,
				topupAPR,
				eligiblePercentage)
			verifyValidatorRewards(
				t,
				setupResult.epochsInYear,
				currentRewardInfo.fullRewards,
				big.NewInt(0).Add(baseStakePerNode, currentRewardInfo.topUpStake),
				tt.validatorExpectedFullROI,
				eligiblePercentage)
		})
	}
}

func setupNodeRewardInfo(
	setupResult SetupRewardsResult,
	vInfo map[uint32][]*state.ValidatorInfo,
	topupStakePerNode *big.Int,
	validatorTopupStake *big.Int,
) (map[uint32][]*nodeRewardsData, error) {
	dummyRwd, _ := NewRewardsCreatorV2(setupResult.RewardsCreatorArgsV2)
	nodesRewardInfo := dummyRwd.initNodesRewardsInfo(vInfo)
	_, _ = setValuesInNodesRewardInfo(nodesRewardInfo, topupStakePerNode, tuStake)

	nodesRewardInfo[0][0].topUpStake = validatorTopupStake

	nrValidatorsToRemoveTopup := big.NewInt(0).Div(validatorTopupStake, topupStakePerNode)

	//remove the newly added topup from some other nodes
	for i := int64(1); i < nrValidatorsToRemoveTopup.Int64(); i++ {
		nodesRewardInfo[0][i].topUpStake = big.NewInt(0)
	}

	return nodesRewardInfo, nil
}

func verifyValidatorRewards(t *testing.T, epochsInYear *big.Int, reward *big.Int, amount *big.Int, expectedROI float64, eligiblePercentage float64) {
	factor := int64(1000 * 1000)
	tolerance := 0.1
	yearRewardsTimesFactor := big.NewInt(0).Mul(reward, big.NewInt(epochsInYear.Int64()*factor))
	roiInt := big.NewInt(0).Div(yearRewardsTimesFactor, amount)
	computedROI := float64(roiInt.Int64()) * 100 / float64(factor) * 0.9 * eligiblePercentage

	diff := math.Abs(expectedROI - computedROI)
	require.True(t, diff < tolerance, fmt.Sprintf("expected %f, actual %f", expectedROI, computedROI))
}

func setUpRewards(args RewardsCreatorArgsV2) SetupRewardsResult {
	nbBlocksPerShard := uint64(14400)
	blocksPerShard := make(map[uint32]uint64)

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	shardCoordinator.CurrentShard = core.MetachainShardId
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return 0
	}
	args.ShardCoordinator = shardCoordinator

	shardMap := createShardsMap(args.ShardCoordinator)
	for shardID := range shardMap {
		blocksPerShard[shardID] = nbBlocksPerShard
	}

	args.EconomicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)
	smallestDivisionGLD, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	totalRewardsFirstYear := big.NewInt(0).Mul(big.NewInt(2169000), smallestDivisionGLD)
	epochsInYear := big.NewInt(365)

	rewardsForBlocks := big.NewInt(0).Div(totalRewardsFirstYear, epochsInYear)
	baseStake := big.NewInt(0).Mul(big.NewInt(2500), smallestDivisionGLD)
	return SetupRewardsResult{
		args, smallestDivisionGLD, totalRewardsFirstYear, epochsInYear, rewardsForBlocks, baseStake,
	}
}

func computeRewardsAndDust(nbEligiblePerShard uint32, args SetupRewardsResult, topupStake *big.Int) (map[uint32][]*nodeRewardsData, *big.Int) {
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	dummyRwd, _ := NewRewardsCreatorV2(args.RewardsCreatorArgsV2)
	nodesRewardInfo := dummyRwd.initNodesRewardsInfo(vInfo)
	_, totalTopUpStake := setValuesInNodesRewardInfo(nodesRewardInfo, topupStake, tuStake)

	totalEligibleStake, _ := big.NewInt(0).SetString("4000000"+"000000000000000000", 10)

	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
			return totalTopUpStake
		},
		GetTotalStakeEligibleNodesCalled: func() *big.Int {
			return totalEligibleStake
		},
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo {
				for i, v := range vList {
					if bytes.Equal(v.PublicKey, blsKey) {
						return nodesRewardInfo[shardID][i].topUpStake, nil
					}
				}
			}
			return nil, fmt.Errorf("not found")
		},
	}

	args.EconomicsDataProvider.SetRewardsToBeDistributedForBlocks(args.rewardsForBlocks)

	rwd, _ := NewRewardsCreatorV2(args.RewardsCreatorArgsV2)

	var dust *big.Int
	nodesRewardInfo, dust = rwd.computeRewardsPerNode(vInfo)
	return nodesRewardInfo, dust
}

func verifyRewardsComputation(
	t *testing.T,
	nbEligiblePerShard uint32,
	srr SetupRewardsResult,
	topupStake *big.Int,
	expectedROI float64,
	eligiblePercentage float64,
) {
	nodesRewardInfo, _ := computeRewardsAndDust(nbEligiblePerShard, srr, topupStake)

	for _, nodeInfoList := range nodesRewardInfo {
		for _, nodeInfo := range nodeInfoList {
			verifyValidatorRewards(t, srr.epochsInYear, nodeInfo.fullRewards, big.NewInt(0).Add(srr.baseStake, topupStake), expectedROI, eligiblePercentage)
		}
	}
}

func TestNewRewardsCreatorV2_computeValidatorInfoPerRewardAddress(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	proposerFee := uint32(100)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, proposerFee, defaultBlocksPerShard)
	dummyRwd, _ := NewRewardsCreatorV2(args)
	nodesRewardInfo := dummyRwd.initNodesRewardsInfo(valInfo)
	_, totalRwd := setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, fRewards, 0)

	rewardsInfo, unassigned := rwd.computeValidatorInfoPerRewardAddress(nodesRewardInfo)
	require.Equal(t, big.NewInt(0), unassigned)

	sumRwds := big.NewInt(0)
	sumFees := big.NewInt(0)
	for _, rwInfo := range rewardsInfo {
		sumRwds.Add(sumRwds, rwInfo.rewardsFromProtocol)
		sumFees.Add(sumFees, rwInfo.accumulatedFees)
	}

	expectedSumFees := big.NewInt(int64(proposerFee))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(nbEligiblePerShard)))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(args.ShardCoordinator.NumberOfShards())+1))

	require.Equal(t, totalRwd, sumRwds)
	require.Equal(t, expectedSumFees, sumFees)
}

func TestNewRewardsCreatorV2_computeValidatorInfoPerRewardAddressWithOfflineValidators(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	proposerFee := uint32(100)
	nbOfflinePerShard := uint32(20)

	nbShards := int64(args.ShardCoordinator.NumberOfShards()) + 1
	args.EconomicsDataProvider.SetLeadersFees(big.NewInt(0).Mul(big.NewInt(int64(proposerFee)), big.NewInt(int64(nbEligiblePerShard-nbOfflinePerShard)*nbShards)))
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, proposerFee, defaultBlocksPerShard)
	for _, valList := range valInfo {
		for i := 0; i < int(nbOfflinePerShard); i++ {
			valList[i].LeaderSuccess = 0
			valList[i].ValidatorSuccess = 0
			valList[i].AccumulatedFees = big.NewInt(0)
		}
	}

	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)
	_, totalRwd := setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, fRewards, 0)
	rewardsInfo, unassigned := rwd.computeValidatorInfoPerRewardAddress(nodesRewardInfo)

	expectedUnassigned := big.NewInt(0)

	shardMap := createShardsMap(args.ShardCoordinator)
	for shardID := range shardMap {
		for i := 0; i < int(nbOfflinePerShard); i++ {
			expectedUnassigned.Add(expectedUnassigned, nodesRewardInfo[shardID][i].fullRewards)
		}
	}
	require.Equal(t, expectedUnassigned, unassigned)

	sumRwds := big.NewInt(0)
	sumFees := big.NewInt(0)
	for _, rwInfo := range rewardsInfo {
		sumRwds.Add(sumRwds, rwInfo.rewardsFromProtocol)
		sumFees.Add(sumFees, rwInfo.accumulatedFees)
	}

	expectedSumFees := big.NewInt(int64(proposerFee))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(nbEligiblePerShard-nbOfflinePerShard)))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(args.ShardCoordinator.NumberOfShards())+1))

	require.Equal(t, big.NewInt(0).Sub(totalRwd, unassigned), sumRwds)
	require.Equal(t, expectedSumFees, sumFees)
}

func TestNewRewardsCreatorV2_computeValidatorInfoPerRewardAddressWithLeavingValidators(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	proposerFee := uint32(100)
	nbLeavingPerShard := uint32(10)
	nbEligiblePerShard := uint32(400)

	nbShards := int64(args.ShardCoordinator.NumberOfShards()) + 1
	args.EconomicsDataProvider.SetLeadersFees(big.NewInt(0).Mul(big.NewInt(int64(proposerFee)), big.NewInt(int64(nbEligiblePerShard)*nbShards)))
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, proposerFee, defaultBlocksPerShard)
	for _, valList := range valInfo {
		for i := 0; i < int(nbLeavingPerShard); i++ {
			valList[i].List = string(common.LeavingList)
		}
	}

	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)
	_, totalRwd := setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, fRewards, 0)
	rewardsInfo, unassigned := rwd.computeValidatorInfoPerRewardAddress(nodesRewardInfo)

	expectedUnassigned := big.NewInt(0)
	require.Equal(t, expectedUnassigned, unassigned)

	sumRwds := big.NewInt(0)
	sumFees := big.NewInt(0)
	for _, rwInfo := range rewardsInfo {
		sumRwds.Add(sumRwds, rwInfo.rewardsFromProtocol)
		sumFees.Add(sumFees, rwInfo.accumulatedFees)
	}

	expectedSumFees := big.NewInt(int64(proposerFee))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(nbEligiblePerShard)))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(args.ShardCoordinator.NumberOfShards())+1))

	require.Equal(t, big.NewInt(0).Sub(totalRwd, unassigned), sumRwds)
	require.Equal(t, expectedSumFees, sumFees)
}

func TestNewRewardsCreatorV2_computeValidatorInfoPerRewardAddressWithDustLeaderFees(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	proposerFee := uint32(100)
	nbEligiblePerShard := uint32(400)
	dustLeaderFees := big.NewInt(100)

	nbShards := int64(args.ShardCoordinator.NumberOfShards()) + 1
	leaderFeesNoDust := big.NewInt(0).Mul(big.NewInt(int64(proposerFee)), big.NewInt(int64(nbEligiblePerShard)*nbShards))
	leaderFeesWithDust := leaderFeesNoDust.Add(leaderFeesNoDust, dustLeaderFees)
	args.EconomicsDataProvider.SetLeadersFees(leaderFeesWithDust)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, proposerFee, defaultBlocksPerShard)

	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)
	dustLeaderFeesUint32 := uint32(dustLeaderFees.Uint64())
	_, totalRwd := setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, fRewards, dustLeaderFeesUint32)
	rewardsInfo, unassigned := rwd.computeValidatorInfoPerRewardAddress(nodesRewardInfo)

	expectedUnassigned := big.NewInt(0).Set(dustLeaderFees)
	require.Equal(t, expectedUnassigned, unassigned)

	sumRwds := big.NewInt(0)
	sumFees := big.NewInt(0)
	for _, rwInfo := range rewardsInfo {
		sumRwds.Add(sumRwds, rwInfo.rewardsFromProtocol)
		sumFees.Add(sumFees, rwInfo.accumulatedFees)
	}

	expectedSumFees := big.NewInt(int64(proposerFee))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(nbEligiblePerShard)))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(args.ShardCoordinator.NumberOfShards())+1))

	require.Equal(t, big.NewInt(0).Sub(totalRwd, unassigned), sumRwds)
	require.Equal(t, expectedSumFees, sumFees)
}

func TestNewRewardsCreatorV2_addValidatorRewardsToMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	miniBlocks := rwd.initializeRewardsMiniBlocks()
	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)
	_, totalRwd := setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, fRewards, 0)

	metaBlock := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}
	sumFees := big.NewInt(0)
	for _, vInfoList := range valInfo {
		for _, vInfo := range vInfoList {
			sumFees.Add(sumFees, vInfo.AccumulatedFees)
		}
	}

	accumulatedDust, err := rwd.addValidatorRewardsToMiniBlocks(metaBlock, miniBlocks, nodesRewardInfo)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0), accumulatedDust)

	sumRewards := big.NewInt(0)

	var tx data.TransactionHandler
	for _, mb := range miniBlocks {
		for _, txHash := range mb.TxHashes {
			tx, err = rwd.currTxs.GetTx(txHash)
			require.Nil(t, err)
			sumRewards.Add(sumRewards, tx.GetValue())
		}
	}

	require.Equal(t, sumRewards, big.NewInt(0).Add(sumFees, totalRwd))
}

func TestNewRewardsCreatorV2_addValidatorRewardsToMiniBlocksAddressInMetaChainDelegationDisabled(t *testing.T) {
	t.Parallel()

	addrInMeta := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255}
	args := getRewardsCreatorV2Arguments()
	args.ShardCoordinator, _ = sharding.NewMultiShardCoordinator(2, 0)
	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	miniBlocks := rwd.initializeRewardsMiniBlocks()
	nodesRewardInfo := rwd.initNodesRewardsInfo(valInfo)
	_, totalRwd := setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, fRewards, 0)

	metaBlock := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}

	nbAddrInMetachainPerShard := 2

	sumFees := big.NewInt(0)
	for _, vInfoList := range valInfo {
		for i, vInfo := range vInfoList {
			if i < nbAddrInMetachainPerShard {
				vInfo.RewardAddress = addrInMeta
			}
			sumFees.Add(sumFees, vInfo.AccumulatedFees)
		}
	}

	accumulatedDust, err := rwd.addValidatorRewardsToMiniBlocks(metaBlock, miniBlocks, nodesRewardInfo)
	require.Nil(t, err)
	require.True(t, big.NewInt(0).Cmp(accumulatedDust) < 0)

	sumRewards := big.NewInt(0)

	var tx data.TransactionHandler
	for _, mb := range miniBlocks {
		for _, txHash := range mb.TxHashes {
			tx, err = rwd.currTxs.GetTx(txHash)
			require.Nil(t, err)
			sumRewards.Add(sumRewards, tx.GetValue())
		}
	}

	require.Equal(t, big.NewInt(0).Add(sumRewards, accumulatedDust), big.NewInt(0).Add(sumFees, totalRwd))
}

func TestNewRewardsCreatorV2_CreateRewardsMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	dummyRwd, _ := NewRewardsCreatorV2(args)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	nodesRewardInfo := dummyRwd.initNodesRewardsInfo(vInfo)
	_, _ = setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, tuStake, 0)

	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
			totalTopUpStake, _ := big.NewInt(0).SetString("3000000000000000000000000", 10)
			return totalTopUpStake
		},
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo {
				for i, v := range vList {
					if bytes.Equal(v.PublicKey, blsKey) {
						return nodesRewardInfo[shardID][i].topUpStake, nil
					}
				}
			}
			return nil, fmt.Errorf("not found")
		},
	}
	nbBlocksPerShard := uint64(14400)
	blocksPerShard := make(map[uint32]uint64)
	shardMap := createShardsMap(args.ShardCoordinator)
	for shardID := range shardMap {
		blocksPerShard[shardID] = nbBlocksPerShard
	}

	args.EconomicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)
	rewardsForBlocks, _ := big.NewInt(0).SetString("5000000000000000000000", 10)
	args.EconomicsDataProvider.SetRewardsToBeDistributedForBlocks(rewardsForBlocks)

	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	metaBlock := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}

	var miniBlocks block.MiniBlockSlice
	miniBlocks, err = rwd.CreateRewardsMiniBlocks(metaBlock, vInfo, &metaBlock.EpochStart.Economics)
	require.Nil(t, err)
	require.NotNil(t, miniBlocks)

	sumRewards := big.NewInt(0)
	var tx data.TransactionHandler
	for _, mb := range miniBlocks {
		for _, txHash := range mb.TxHashes {
			tx, err = rwd.currTxs.GetTx(txHash)
			require.Nil(t, err)
			sumRewards.Add(sumRewards, tx.GetValue())
		}
	}

	sumFees := big.NewInt(0)
	for _, vInfoList := range vInfo {
		for _, v := range vInfoList {
			sumFees.Add(sumFees, v.AccumulatedFees)
		}
	}

	totalRws := rwd.economicsDataProvider.RewardsToBeDistributedForBlocks()
	rewardsForProtocolSustainability := big.NewInt(0).Set(metaBlock.EpochStart.Economics.RewardsForProtocolSustainability)
	expectedRewards := big.NewInt(0).Add(sumFees, totalRws)
	expectedRewards.Add(expectedRewards, rewardsForProtocolSustainability)
	require.Equal(t, expectedRewards, sumRewards)

	// now verification
	metaBlock.MiniBlockHeaders = make([]block.MiniBlockHeader, len(miniBlocks))

	var mbHash []byte
	for i, mb := range miniBlocks {
		mbHash, err = core.CalculateHash(args.Marshalizer, args.Hasher, mb)
		require.Nil(t, err)
		metaBlock.MiniBlockHeaders[i] = block.MiniBlockHeader{
			Hash:            mbHash,
			SenderShardID:   mb.SenderShardID,
			ReceiverShardID: mb.ReceiverShardID,
			TxCount:         uint32(len(mb.TxHashes)),
			Type:            mb.Type,
			Reserved:        nil,
		}
	}

	err = rwd.VerifyRewardsMiniBlocks(metaBlock, vInfo, &metaBlock.EpochStart.Economics)
	require.Nil(t, err)
}

func TestNewRewardsCreatorV2_CreateRewardsMiniBlocks2169Nodes(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	dummyRwd, _ := NewRewardsCreatorV2(args)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100, defaultBlocksPerShard)
	nodesRewardInfo := dummyRwd.initNodesRewardsInfo(vInfo)
	multiplier, _ := big.NewInt(0).SetString("1000000000000000000", 10)
	topupValue := big.NewInt(2500)
	topupValue.Mul(topupValue, multiplier)
	_, totalTopupStake := setValuesInNodesRewardInfo(nodesRewardInfo, topupValue, tuStake)

	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
			return totalTopupStake
		},
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo {
				for i, v := range vList {
					if bytes.Equal(v.PublicKey, blsKey) {
						return nodesRewardInfo[shardID][i].topUpStake, nil
					}
				}
			}
			return nil, fmt.Errorf("not found")
		},
	}
	nbBlocksPerShard := uint64(14400)
	blocksPerShard := make(map[uint32]uint64)
	shardMap := createShardsMap(args.ShardCoordinator)
	for shardID := range shardMap {
		blocksPerShard[shardID] = nbBlocksPerShard
	}

	args.EconomicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)
	rewardsForBlocks := big.NewInt(0).Mul(big.NewInt(5000), multiplier)
	args.EconomicsDataProvider.SetRewardsToBeDistributedForBlocks(rewardsForBlocks)

	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	metaBlock := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}

	var miniBlocks block.MiniBlockSlice
	miniBlocks, err = rwd.CreateRewardsMiniBlocks(metaBlock, vInfo, &metaBlock.EpochStart.Economics)
	require.Nil(t, err)
	require.NotNil(t, miniBlocks)

	sumRewards := big.NewInt(0)
	var tx data.TransactionHandler
	for _, mb := range miniBlocks {
		for _, txHash := range mb.TxHashes {
			tx, err = rwd.currTxs.GetTx(txHash)
			require.Nil(t, err)
			sumRewards.Add(sumRewards, tx.GetValue())
		}
	}

	sumFees := big.NewInt(0)
	for _, vInfoList := range vInfo {
		for _, v := range vInfoList {
			sumFees.Add(sumFees, v.AccumulatedFees)
		}
	}

	totalRws := rwd.economicsDataProvider.RewardsToBeDistributedForBlocks()
	rewardsForProtocolSustainability := big.NewInt(0).Set(metaBlock.EpochStart.Economics.RewardsForProtocolSustainability)
	expectedRewards := big.NewInt(0).Add(sumFees, totalRws)
	expectedRewards.Add(expectedRewards, rewardsForProtocolSustainability)
	require.Equal(t, expectedRewards, sumRewards)

	// now verification
	metaBlock.MiniBlockHeaders = make([]block.MiniBlockHeader, len(miniBlocks))

	var mbHash []byte
	for i, mb := range miniBlocks {
		mbHash, err = core.CalculateHash(args.Marshalizer, args.Hasher, mb)
		require.Nil(t, err)
		metaBlock.MiniBlockHeaders[i] = block.MiniBlockHeader{
			Hash:            mbHash,
			SenderShardID:   mb.SenderShardID,
			ReceiverShardID: mb.ReceiverShardID,
			TxCount:         uint32(len(mb.TxHashes)),
			Type:            mb.Type,
			Reserved:        nil,
		}
	}

	err = rwd.VerifyRewardsMiniBlocks(metaBlock, vInfo, &metaBlock.EpochStart.Economics)
	require.Nil(t, err)
}

func getRewardsCreatorV2Arguments() RewardsCreatorArgsV2 {
	rewardsTopUpGradientPoint, _ := big.NewInt(0).SetString("3000000000000000000000000", 10)
	topUpRewardFactor := 0.25

	rewardsHandler := &economicsmocks.EconomicsHandlerStub{
		RewardsTopUpGradientPointCalled: func() *big.Int {
			return big.NewInt(0).Set(rewardsTopUpGradientPoint)
		},
		RewardsTopUpFactorCalled: func() float64 {
			return topUpRewardFactor
		},
	}
	return RewardsCreatorArgsV2{
		BaseRewardsCreatorArgs: getBaseRewardsArguments(),
		StakingDataProvider:    &mock.StakingDataProviderStub{},
		EconomicsDataProvider:  NewEpochEconomicsStatistics(),
		RewardsHandler:         rewardsHandler,
	}
}

func getRewardsCreatorV35Arguments() RewardsCreatorArgsV2 {
	rewardsTopUpGradientPoint, _ := big.NewInt(0).SetString("2000000000000000000000000", 10)
	topUpRewardFactor := 0.5

	rewardsHandler := &economicsmocks.EconomicsHandlerStub{
		RewardsTopUpGradientPointCalled: func() *big.Int {
			return big.NewInt(0).Set(rewardsTopUpGradientPoint)
		},
		RewardsTopUpFactorCalled: func() float64 {
			return topUpRewardFactor
		},
	}
	return RewardsCreatorArgsV2{
		BaseRewardsCreatorArgs: getBaseRewardsArguments(),
		StakingDataProvider:    &mock.StakingDataProviderStub{},
		EconomicsDataProvider:  NewEpochEconomicsStatistics(),
		RewardsHandler:         rewardsHandler,
	}
}

func setDummyValuesInNodesRewardInfo(
	nodesRewardInfo map[uint32][]*nodeRewardsData,
	eligibleNodesPerShard uint32,
	field string,
	dustLeaderFees uint32,
) (map[uint32]*big.Int, *big.Int) {
	cumulatedValuePerShard := make(map[uint32]*big.Int)
	totalValue := big.NewInt(0)

	valTopUp := 1000 + eligibleNodesPerShard
	for shardID, nodeInfoList := range nodesRewardInfo {
		cumulatedValuePerShard[shardID] = big.NewInt(0)
		for i, nodeInfo := range nodeInfoList {
			v := big.NewInt(int64(int(valTopUp) - i))
			multiplier, _ := big.NewInt(0).SetString("10000000000000000", 10)
			v.Mul(v, multiplier)
			switch field {
			case tuStake:
				nodeInfo.topUpStake = v
			case tuRewards:
				nodeInfo.topUpReward = v
			case bRewards:
				nodeInfo.baseReward = v
			case fRewards:
				nodeInfo.fullRewards = v
			}

			cumulatedValuePerShard[shardID].Add(cumulatedValuePerShard[shardID], v)
		}
		totalValue.Add(totalValue, cumulatedValuePerShard[shardID])
	}

	return cumulatedValuePerShard, totalValue.Add(totalValue, big.NewInt(0).SetInt64(int64(dustLeaderFees)))
}

func setValuesInNodesRewardInfo(
	nodesRewardInfo map[uint32][]*nodeRewardsData,
	v *big.Int,
	field string,
) (map[uint32]*big.Int, *big.Int) {
	cumulatedValuePerShard := make(map[uint32]*big.Int)
	totalValue := big.NewInt(0)
	for shardID, nodeInfoList := range nodesRewardInfo {
		cumulatedValuePerShard[shardID] = big.NewInt(0)
		for _, nodeInfo := range nodeInfoList {
			switch field {
			case tuStake:
				nodeInfo.topUpStake = v
			case tuRewards:
				nodeInfo.topUpReward = v
			case bRewards:
				nodeInfo.baseReward = v
			case fRewards:
				nodeInfo.fullRewards = v
			}

			cumulatedValuePerShard[shardID].Add(cumulatedValuePerShard[shardID], v)
		}
		totalValue.Add(totalValue, cumulatedValuePerShard[shardID])
	}

	return cumulatedValuePerShard, totalValue
}

func createDefaultValidatorInfo(
	eligibleNodesPerShard uint32,
	shardCoordinator sharding.Coordinator,
	nodesConfigProvider epochStart.NodesConfigProvider,
	proposerFeesPerNode uint32,
	nbBlocksPerShard uint32,
) map[uint32][]*state.ValidatorInfo {
	cGrShard := uint32(nodesConfigProvider.ConsensusGroupSize(0))
	cGrMeta := uint32(nodesConfigProvider.ConsensusGroupSize(core.MetachainShardId))
	nbBlocksSelectedNodeInShard := nbBlocksPerShard * cGrShard / eligibleNodesPerShard
	nbBlocksSelectedNodeInMeta := nbBlocksPerShard * cGrMeta / eligibleNodesPerShard

	shardsMap := createShardsMap(shardCoordinator)

	var nbBlocksSelected uint32
	validators := make(map[uint32][]*state.ValidatorInfo)
	for shardID := range shardsMap {
		validators[shardID] = make([]*state.ValidatorInfo, eligibleNodesPerShard)
		nbBlocksSelected = nbBlocksSelectedNodeInShard
		if shardID == core.MetachainShardId {
			nbBlocksSelected = nbBlocksSelectedNodeInMeta
		}

		for i := uint32(0); i < eligibleNodesPerShard; i++ {
			str := fmt.Sprintf("rewardAddr%d_%d", shardID, i)
			addrHex := make([]byte, len(str)*2)
			_ = hex.Encode(addrHex, []byte(str))

			leaderSuccess := uint32(20)
			validators[shardID][i] = &state.ValidatorInfo{
				PublicKey:                  []byte(fmt.Sprintf("pubKeyBLS%d%d", shardID, i)),
				ShardId:                    shardID,
				RewardAddress:              addrHex,
				LeaderSuccess:              leaderSuccess,
				ValidatorSuccess:           nbBlocksSelected - leaderSuccess,
				NumSelectedInSuccessBlocks: nbBlocksSelected,
				AccumulatedFees:            big.NewInt(int64(proposerFeesPerNode)),
				List:                       string(common.EligibleList),
			}
		}
	}

	return validators
}

func addNonEligibleValidatorInfo(
	nonEligiblePerShard uint32,
	validatorsInfo map[uint32][]*state.ValidatorInfo,
	list string,
) map[uint32][]*state.ValidatorInfo {
	resultedValidatorsInfo := make(map[uint32][]*state.ValidatorInfo)
	for shardID, valInfoList := range validatorsInfo {
		for i := uint32(0); i < nonEligiblePerShard; i++ {
			vInfo := &state.ValidatorInfo{
				PublicKey:                  []byte(fmt.Sprintf("pubKeyBLSExtra%d", i)),
				ShardId:                    shardID,
				RewardAddress:              []byte(fmt.Sprintf("addrRewardsExtra%d", i)),
				LeaderSuccess:              1,
				ValidatorSuccess:           2,
				NumSelectedInSuccessBlocks: 1,
				AccumulatedFees:            big.NewInt(int64(10)),
				List:                       list,
			}
			resultedValidatorsInfo[shardID] = append(valInfoList, vInfo)
		}
	}

	return resultedValidatorsInfo
}
