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
	rwd.protocolSustainability = big.NewInt(10)
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
	require.Equal(t, big.NewInt(0), rwd.protocolSustainability)
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

func TestNewEpochStartRewardsCreatorV2_computeNodePowerInShard(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	rwd, err := NewEpochStartRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	nodesPerShard := uint32(400)
	valInfo := createDefaultValidatorInfo(nodesPerShard, args.ShardCoordinator, args.NodesConfigProvider, 100)
	topUpPerNode := createDefaultTopUpPerNode(400, args.ShardCoordinator)

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

func createDefaultTopUpPerNode(
	eligibleNodesPerShard uint32,
	shardCoordinator sharding.Coordinator,
) map[uint32][]*big.Int {
	topUpPerNode := make(map[uint32][]*big.Int)
	shardMap := createShardsMap(shardCoordinator)

	valTopUp := 1000 + eligibleNodesPerShard
	for shardID := range shardMap {
		topUpPerNode[shardID] = make([]*big.Int, eligibleNodesPerShard)
		for i := uint32(0); i < eligibleNodesPerShard; i++ {
			topUpPerNode[shardID][i] = big.NewInt(int64(valTopUp - i))
		}
	}

	return topUpPerNode
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
			}
		}
	}

	return validators
}

func createShardsMap(shardCoordinator sharding.Coordinator) map[uint32]struct{} {
	shardsMap := make(map[uint32]struct{}, shardCoordinator.NumberOfShards())
	for shardID := uint32(0); shardID < shardCoordinator.NumberOfShards(); shardID++ {
		shardsMap[shardID] = struct{}{}
	}
	shardsMap[core.MetachainShardId] = struct{}{}
	return shardsMap
}
