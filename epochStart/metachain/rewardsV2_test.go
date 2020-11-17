package metachain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

func TestNewEpochStartRewardsCreator_InvalidBaseRewardsCreatorArgs(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.BaseRewardsCreatorArgs.NodesConfigProvider = nil

	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrNilNodesConfigProvider, err)
}

func TestNewEpochStartRewardsCreator_NilStakingDataProvider(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.StakingDataProvider = nil

	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrNilStakingDataProvider, err)
}

func TestNewEpochStartRewardsCreator_NilEconomicsDataProvider(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.EconomicsDataProvider = nil

	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrNilEconomicsDataProvider, err)
}

func TestNewEpochStartRewardsCreator_NegativeGradientPointShouldErr(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.TopUpGradientPoint = big.NewInt(-1)

	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrInvalidRewardsTopUpGradientPoint, err)
}

func TestNewEpochStartRewardsCreator_NegativeTopUpRewardFactorShouldErr(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.TopUpRewardFactor = -1

	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrInvalidRewardsTopUpFactor, err)
}

func TestNewEpochStartRewardsCreator_SupraUnitaryTopUpRewardFactorShouldErr(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	args.TopUpRewardFactor = 1.5

	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.True(t, check.IfNil(rwd))
	require.Equal(t, epochStart.ErrInvalidRewardsTopUpFactor, err)
}

func TestNewEpochStartRewardsCreatorOK(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)
}

func TestRewardsCreator_CreateRewardsMiniBlocksV2ComputeErrorsShouldErr(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	numComputeRewards := 0
	expectedErr := errors.New("expected error")
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		PrepareStakingDataCalled: func(keys map[uint32][][]byte) error {
			numComputeRewards += len(keys)
			return expectedErr
		},
	}
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)

	mb := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}
	valInfo := make(map[uint32][]*state.ValidatorInfo)
	valInfo[0] = []*state.ValidatorInfo{
		{
			PublicKey:       []byte("pubkey"),
			ShardId:         0,
			AccumulatedFees: big.NewInt(100),
		},
	}
	bdy, err := rwd.CreateRewardsMiniBlocks(mb, valInfo)
	require.NotNil(t, err)
	require.Equal(t, expectedErr, err)
	require.Nil(t, bdy)
	require.Equal(t, 1, numComputeRewards)
}

func TestNewEpochStartRewardsCreatorV2_getEligibleNodesKeyMap(t *testing.T) {
	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	valInfo := createDefaultValidatorInfo(400, args.ShardCoordinator, args.NodesConfigProvider, 100)
	nodesKeyMap := rwd.getEligibleNodesKeyMap(valInfo)

	for shardID, nodesListInfo := range valInfo {
		for i, nodeInfo := range nodesListInfo {
			require.Equal(t, nodesKeyMap[shardID][i], nodeInfo.PublicKey)
		}
	}
}

func TestNewEpochStartRewardsCreatorV2_prepareRewardsDataCleansLocalData(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	rwd.accumulatedRewards = big.NewInt(10)
	rwd.protocolSustainabilityValue = big.NewInt(10)
	rwd.mapBaseRewardsPerBlockPerValidator = make(map[uint32]*big.Int)
	rwd.mapBaseRewardsPerBlockPerValidator[0] = big.NewInt(10)
	rwd.mapBaseRewardsPerBlockPerValidator[1] = big.NewInt(10)
	rwd.mapBaseRewardsPerBlockPerValidator[core.MetachainShardId] = big.NewInt(10)

	valInfo := createDefaultValidatorInfo(400, args.ShardCoordinator, args.NodesConfigProvider, 100)
	metaBlk := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}

	err = rwd.prepareRewardsData(metaBlk, valInfo)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0), rwd.accumulatedRewards)
	require.Equal(t, big.NewInt(0), rwd.protocolSustainabilityValue)
	require.Equal(t, 0, len(rwd.mapBaseRewardsPerBlockPerValidator))
}

func TestNewEpochStartRewardsCreatorV2_getTopUpForAllEligibleNodes(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	topUpVal, _ := big.NewInt(0).SetString("100000000000000000000", 10)
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			t := big.NewInt(0).Set(topUpVal)
			return t, nil
		},
	}
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)

	topUp := rwd.getTopUpForAllEligibleNodes(valInfo)
	for _, nodesTopUpList := range topUp {
		require.Equal(t, nodesPerShard, uint32(len(nodesTopUpList)))
		for _, nodeTopUp := range nodesTopUpList {
			require.Equal(t, topUpVal, nodeTopUp)
		}
	}
}

func TestNewEpochStartRewardsCreatorV2_getTopUpForAllEligibleSomeBLSKeysNotFoundZeroed(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	topUpVal, _ := big.NewInt(0).SetString("100000000000000000000", 10)
	notFoundKey := []byte("notFound")
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			if bytes.Compare(blsKey, notFoundKey) == 0 {
				return nil, fmt.Errorf("not found")
			}
			t := big.NewInt(0).Set(topUpVal)
			return t, nil
		},
	}
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(10)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)
	for _, valList := range valInfo {
		valList[0].PublicKey = notFoundKey
		valList[1].PublicKey = notFoundKey
	}

	topUp := rwd.getTopUpForAllEligibleNodes(valInfo)
	for _, nodesTopUpList := range topUp {
		require.Equal(t, nodesPerShard, uint32(len(nodesTopUpList)))
		for i, nodeTopUp := range nodesTopUpList {
			if i < 2 {
				require.Equal(t, big.NewInt(0), nodeTopUp)
			} else {
				require.Equal(t, topUpVal, nodeTopUp)
			}
		}
	}
}

func TestNewEpochStartRewardsCreatorV2_aggregateBaseAndTopUpRewardsPerNode(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := 10
	baseRewards := make(map[uint32][]*big.Int)
	topUpRewards := make(map[uint32][]*big.Int)

	value := 1000 + nodesPerShard
	shardMap := createShardsMap(args.ShardCoordinator)
	for shardID := range shardMap {
		baseRewards[shardID] = make([]*big.Int, nodesPerShard)
		topUpRewards[shardID] = make([]*big.Int, nodesPerShard)
		for i := 0; i < nodesPerShard; i++ {
			baseRewards[shardID][i] = big.NewInt(int64(value - i))
			topUpRewards[shardID][i] = big.NewInt(int64(value + i))
		}
	}

	aggRwdPerNode := aggregateBaseAndTopUpRewardsPerNode(baseRewards, topUpRewards)

	for _, rwdPerNodeList := range aggRwdPerNode {
		for _, reward := range rwdPerNodeList {
			require.Equal(t, big.NewInt(2*int64(value)), reward)
		}
	}
}

func TestNewEpochStartRewardsCreatorV2_computeNodePowerInShardOfflineNodeZeroPower(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
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

func TestNewEpochStartRewardsCreatorV2_computeNodePowerInShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
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

func TestNewEpochStartRewardsCreatorV2_computeNodesPowerInShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)
	topUpPerNode, _, _ := createMapWithBigIntValuesPerNode(400, args.ShardCoordinator)

	totalNodesInShardPower, nodePowerInShard := computeNodesPowerInShard(valInfo, topUpPerNode)
	require.NotNil(t, totalNodesInShardPower)
	require.NotNil(t, nodePowerInShard)

	for shardID, powerNodeList := range nodePowerInShard {
		for i, powerNode := range powerNodeList {
			blocks := valInfo[shardID][i].NumSelectedInSuccessBlocks
			topUp := topUpPerNode[shardID][i]
			require.Equal(t, big.NewInt(0).Mul(big.NewInt(int64(blocks)), topUp), powerNode)
		}
	}
}

func TestNewEpochStartRewardsCreatorV2_computeShardsPower(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
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

func TestNewEpochStartRewardsCreatorV2_computeRewardsForPowerPerShardZeroTotalPower(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
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

func TestNewEpochStartRewardsCreatorV2_computeRewardsForPowerPerShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
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

func TestNewEpochStartRewardsCreatorV2_computeTopUpRewardsPerShardNoBlockPerShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	topUpPerNode, _, _ := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)

	topUpRewards := big.NewInt(3000)
	rwd.economicsDataProvider.SetNumberOfBlocksPerShard(nil)

	topUpRewardsPerShard := rwd.computeTopUpRewardsPerShard(topUpRewards, topUpPerNode)

	for _, shardTopUpRewards := range topUpRewardsPerShard {
		require.Equal(t, big.NewInt(0), shardTopUpRewards)
	}
}

func TestNewEpochStartRewardsCreatorV2_computeTopUpRewardsPerShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	topUpPerNode, topUpPerShard, _ := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)

	blocksPerShard := make(map[uint32]uint64)
	blocksPerShard[0] = 2000
	blocksPerShard[1] = 1200
	blocksPerShard[core.MetachainShardId] = 1500

	topUpRewards := big.NewInt(3000)
	rwd.economicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)

	topUpRewardsPerShard := rwd.computeTopUpRewardsPerShard(topUpRewards, topUpPerNode)

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

func TestNewEpochStartRewardsCreatorV2_computeTopUpRewardsZeroTopup(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	totalToDistribute, _ := big.NewInt(-1).SetString("3000000000000000000000", 10)
	totalTopUpEligible := big.NewInt(0)

	topUpRewards := rwd.computeTopUpRewards(totalToDistribute, totalTopUpEligible)
	require.NotNil(t, topUpRewards)
	require.Equal(t, big.NewInt(0), topUpRewards)
}

func TestNewEpochStartRewardsCreatorV2_computeTopUpRewardsNegativeToDistribute(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	totalToDistribute := big.NewInt(-1)
	totalTopUpEligible := big.NewInt(10000)

	topUpRewards := rwd.computeTopUpRewards(totalToDistribute, totalTopUpEligible)
	require.NotNil(t, topUpRewards)
	require.Equal(t, big.NewInt(0), topUpRewards)
}

func TestNewEpochStartRewardsCreatorV2_computeTopUpRewards(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	totalToDistribute, _ := big.NewInt(0).SetString("3000000000000000000000", 10)
	topUpRewardsLimit := core.GetPercentageOfValue(totalToDistribute, rwd.topUpRewardFactor)

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
	ninetyNinePercentLimit := core.GetPercentageOfValue(topUpRewardsLimit, 0.99)
	require.True(t, topUpRewards.Cmp(ninetyNinePercentLimit) > 0)
}

func TestNewEpochStartRewardsCreatorV2_computeTopUpRewardsPerNode(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	dummyTopUp, _, _ := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo {
				for i, v := range vList {
					if bytes.Compare(v.PublicKey, blsKey) == 0 {
						return dummyTopUp[shardID][i], nil
					}
				}
			}
			return nil, fmt.Errorf("not found")
		},
	}
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	topUpRewards, _ := big.NewInt(0).SetString("700000000000000000000", 10)
	blocksPerShard := make(map[uint32]uint64)
	blocksPerShard[0] = 2000
	blocksPerShard[1] = 1200
	blocksPerShard[core.MetachainShardId] = 1500
	rwd.economicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)

	topUpRewardsPerNode, dust := rwd.computeTopUpRewardsPerNode(topUpRewards, vInfo)

	sumTopUpRewards := big.NewInt(0)
	for _, rewardList := range topUpRewardsPerNode {
		for _, reward := range rewardList {
			sumTopUpRewards.Add(sumTopUpRewards, reward)
		}
	}

	// dust should be really small, checking against  topupRewards/1mil
	limit := core.GetPercentageOfValue(topUpRewards, 0.000001)
	require.True(t, limit.Cmp(dust) > 0)

	sumTopUpRewards.Add(sumTopUpRewards, dust)
	require.Equal(t, topUpRewards, sumTopUpRewards)
}

func TestNewEpochStartRewardsCreatorV2_computeTopUpRewardsPerNodeNotFoundBLSKeys(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)
	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	topUpRewards, _ := big.NewInt(0).SetString("700000000000000000000", 10)
	blocksPerShard := make(map[uint32]uint64)
	blocksPerShard[0] = 2000
	blocksPerShard[1] = 1200
	blocksPerShard[core.MetachainShardId] = 1500
	rwd.economicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)

	topUpRewardsPerNode, dust := rwd.computeTopUpRewardsPerNode(topUpRewards, vInfo)

	sumTopUpRewards := big.NewInt(0)
	for _, rewardList := range topUpRewardsPerNode {
		for _, reward := range rewardList {
			sumTopUpRewards.Add(sumTopUpRewards, reward)
		}
	}

	// in this case, all rewards accumulated in dust, when all retrieved BLS keys are invalid, or not found
	require.True(t, topUpRewards.Cmp(dust) == 0)

	sumTopUpRewards.Add(sumTopUpRewards, dust)
	require.Equal(t, topUpRewards, sumTopUpRewards)
}

func TestNewEpochStartRewardsCreatorV2_computeBaseRewardsPerNode(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)
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

	baseRwdPerNode, dust := rwd.computeBaseRewardsPerNode(valInfo, baseRewards)

	// dust should be really small, checking against  baseRewards/1mil
	limit := core.GetPercentageOfValue(baseRewards, 0.000001)
	require.True(t, limit.Cmp(dust) > 0)

	sumRwds := big.NewInt(0)
	for _, rwdList := range baseRwdPerNode {
		for _, nodeRwd := range rwdList {
			sumRwds.Add(sumRwds, nodeRwd)
		}
	}

	require.Equal(t, big.NewInt(0).Add(sumRwds, dust), baseRewards)
}

func TestNewEpochStartRewardsCreatorV2_computeRewardsPerNode(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	dummyTopUp, _, _ := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)

	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
			totalTopUpStake, _ := big.NewInt(0).SetString("3000000000000000000000000", 10)
			return totalTopUpStake
		},
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo {
				for i, v := range vList {
					if bytes.Compare(v.PublicKey, blsKey) == 0 {
						return dummyTopUp[shardID][i], nil
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

	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	rewardsPerNode, accumulatedDust := rwd.computeRewardsPerNode(vInfo)

	// dust should be really small, checking against  baseRewards/1mil
	limit := core.GetPercentageOfValue(rewardsForBlocks, 0.000001)
	require.True(t, limit.Cmp(accumulatedDust) > 0)

	sumRwds := big.NewInt(0)
	for _, rwdList := range rewardsPerNode {
		for _, nodeRwd := range rwdList {
			sumRwds.Add(sumRwds, nodeRwd)
		}
	}

	require.Equal(t, rewardsForBlocks, big.NewInt(0).Add(sumRwds, accumulatedDust))
}

func TestNewEpochStartRewardsCreatorV2_computeValidatorInfoPerRewardAddress(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	proposerFee := uint32(100)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, proposerFee)
	rwdPerNode, _, totalRwd := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)

	rewardsInfo, unassigned := rwd.computeValidatorInfoPerRewardAddress(valInfo, rwdPerNode)
	require.Equal(t, big.NewInt(0), unassigned)

	sumRwds := big.NewInt(0)
	sumFees := big.NewInt(0)
	for _, rwInfo := range rewardsInfo {
		sumRwds.Add(sumRwds, rwInfo.protocolRewards)
		sumFees.Add(sumFees, rwInfo.accumulatedFees)
	}

	expectedSumFees := big.NewInt(int64(proposerFee))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(nbEligiblePerShard)))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(args.ShardCoordinator.NumberOfShards())+1))

	require.Equal(t, totalRwd, sumRwds)
	require.Equal(t, expectedSumFees, sumFees)
}

func TestNewEpochStartRewardsCreatorV2_computeValidatorInfoPerRewardAddressWithOfflineValidators(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	proposerFee := uint32(100)
	nbOfflinePerShard := 20

	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, proposerFee)
	for _, valList := range valInfo {
		for i := 0; i < nbOfflinePerShard; i++ {
			valList[i].LeaderSuccess = 0
			valList[i].ValidatorSuccess = 0
			valList[i].AccumulatedFees = big.NewInt(0)
		}
	}

	rwdPerNode, _, totalRwd := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)
	rewardsInfo, unassigned := rwd.computeValidatorInfoPerRewardAddress(valInfo, rwdPerNode)

	expectedUnassigned := big.NewInt(0)

	shardMap := createShardsMap(args.ShardCoordinator)
	for shardID := range shardMap {
		for i := 0; i < nbOfflinePerShard; i++ {
			expectedUnassigned.Add(expectedUnassigned, rwdPerNode[shardID][i])
		}
	}
	require.Equal(t, expectedUnassigned, unassigned)

	sumRwds := big.NewInt(0)
	sumFees := big.NewInt(0)
	for _, rwInfo := range rewardsInfo {
		sumRwds.Add(sumRwds, rwInfo.protocolRewards)
		sumFees.Add(sumFees, rwInfo.accumulatedFees)
	}

	expectedSumFees := big.NewInt(int64(proposerFee))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(nbEligiblePerShard-uint32(nbOfflinePerShard))))
	expectedSumFees.Mul(expectedSumFees, big.NewInt(int64(args.ShardCoordinator.NumberOfShards())+1))

	require.Equal(t, big.NewInt(0).Sub(totalRwd, unassigned), sumRwds)
	require.Equal(t, expectedSumFees, sumFees)
}

func TestNewEpochStartRewardsCreatorV2_addValidatorRewardsToMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)
	miniBlocks := rwd.initializeRewardsMiniBlocks()
	rwdPerNode, _, totalRwd := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)
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

	accumulatedDust, err := rwd.addValidatorRewardsToMiniBlocks(valInfo, metaBlock, miniBlocks, rwdPerNode)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0), accumulatedDust)

	sumRewards := big.NewInt(0)
	for _, mb := range miniBlocks {
		for _, txHash := range mb.TxHashes {
			tx, err := rwd.currTxs.GetTx(txHash)
			require.Nil(t, err)
			sumRewards.Add(sumRewards, tx.GetValue())
		}
	}

	require.Equal(t, sumRewards, big.NewInt(0).Add(sumFees, totalRwd))
}

func TestNewEpochStartRewardsCreatorV2_addValidatorRewardsToMiniBlocksAddressInMetaChainDelegationDisabled(t *testing.T) {
	t.Parallel()

	addrInMeta := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255}
	args := getRewardsCreatorV2Arguments()
	args.ShardCoordinator, _ = sharding.NewMultiShardCoordinator(2, 0)
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nbEligiblePerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)
	miniBlocks := rwd.initializeRewardsMiniBlocks()
	rwdPerNode, _, totalRwd := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)
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

	accumulatedDust, err := rwd.addValidatorRewardsToMiniBlocks(valInfo, metaBlock, miniBlocks, rwdPerNode)
	require.Nil(t, err)
	require.True(t, big.NewInt(0).Cmp(accumulatedDust) < 0)

	sumRewards := big.NewInt(0)
	for _, mb := range miniBlocks {
		for _, txHash := range mb.TxHashes {
			tx, err := rwd.currTxs.GetTx(txHash)
			require.Nil(t, err)
			sumRewards.Add(sumRewards, tx.GetValue())
		}
	}

	require.Equal(t, big.NewInt(0).Add(sumRewards, accumulatedDust), big.NewInt(0).Add(sumFees, totalRwd))
}

func TestNewEpochStartRewardsCreatorV2_CreateRewardsMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	dummyTopUp, _, _ := createMapWithBigIntValuesPerNode(nbEligiblePerShard, args.ShardCoordinator)
	vInfo := createDefaultValidatorInfo(nbEligiblePerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)

	args.StakingDataProvider = &mock.StakingDataProviderStub{
		GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
			totalTopUpStake, _ := big.NewInt(0).SetString("3000000000000000000000000", 10)
			return totalTopUpStake
		},
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo {
				for i, v := range vList {
					if bytes.Compare(v.PublicKey, blsKey) == 0 {
						return dummyTopUp[shardID][i], nil
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

	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	metaBlock := &block.MetaBlock{
		EpochStart:     getDefaultEpochStart(),
		DevFeesInEpoch: big.NewInt(0),
	}

	miniBlocks, err := rwd.CreateRewardsMiniBlocks(metaBlock, vInfo)
	require.Nil(t, err)
	require.NotNil(t, miniBlocks)

	sumRewards := big.NewInt(0)
	for _, mb := range miniBlocks {
		for _, txHash := range mb.TxHashes {
			tx, err := rwd.currTxs.GetTx(txHash)
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
	for i, mb := range miniBlocks {
		mbHash, err := core.CalculateHash(args.Marshalizer, args.Hasher, mb)
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

	err = rwd.VerifyRewardsMiniBlocks(metaBlock, vInfo)
	require.Nil(t, err)
}

func getRewardsCreatorV2Arguments() RewardsCreatorArgsV2 {
	rewardsTopUpGradientPoint, _ := big.NewInt(0).SetString("3000000000000000000000000", 10)
	return RewardsCreatorArgsV2{
		BaseRewardsCreatorArgs: getBaseRewardsArguments(),
		StakingDataProvider:    &mock.StakingDataProviderStub{},
		EconomicsDataProvider:  NewEpochEconomicsStatistics(),
		TopUpRewardFactor:      0.25,
		TopUpGradientPoint:     rewardsTopUpGradientPoint,
	}
}

func createMapWithBigIntValuesPerNode(
	eligibleNodesPerShard uint32,
	shardCoordinator sharding.Coordinator,
) (map[uint32][]*big.Int, map[uint32]*big.Int, *big.Int) {
	cumulatedValuePerShard := make(map[uint32]*big.Int)
	valuePerNode := make(map[uint32][]*big.Int)
	totalValue := big.NewInt(0)
	shardMap := createShardsMap(shardCoordinator)

	valTopUp := 1000 + eligibleNodesPerShard
	for shardID := range shardMap {
		cumulatedValuePerShard[shardID] = big.NewInt(0)
		valuePerNode[shardID] = make([]*big.Int, eligibleNodesPerShard)
		for i := uint32(0); i < eligibleNodesPerShard; i++ {
			valuePerNode[shardID][i] = big.NewInt(int64(valTopUp - i))
			cumulatedValuePerShard[shardID].Add(cumulatedValuePerShard[shardID], valuePerNode[shardID][i])
		}
		totalValue.Add(totalValue, cumulatedValuePerShard[shardID])
	}

	return valuePerNode, cumulatedValuePerShard, totalValue
}

func createDefaultValidatorInfo(
	eligibleNodesPerShard uint32,
	shardCoordinator sharding.Coordinator,
	nodesConfigProvider epochStart.NodesConfigProvider,
	proposerFeesPerNode uint32,
) map[uint32][]*state.ValidatorInfo {
	nbBlocksPerShard := uint32(14400)
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
				PublicKey:                  nil,
				ShardId:                    shardID,
				RewardAddress:              addrHex,
				LeaderSuccess:              leaderSuccess,
				ValidatorSuccess:           nbBlocksSelected - leaderSuccess,
				NumSelectedInSuccessBlocks: nbBlocksSelected,
				AccumulatedFees:            big.NewInt(int64(proposerFeesPerNode)),
				List:                       string(core.EligibleList),
			}
		}
	}

	return validators
}
