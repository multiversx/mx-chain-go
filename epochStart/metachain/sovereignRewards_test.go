package metachain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/stakingcommon"
	"github.com/stretchr/testify/require"
)

func createSovereignDefaultValidatorInfo(
	eligibleNodesPerShard uint32,
	nodesConfigProvider epochStart.NodesConfigProvider,
	proposerFeesPerNode uint32,
	nbBlocksPerShard uint32,
) state.ShardValidatorsInfoMapHandler {
	shardID := core.SovereignChainShardId

	cGrShard := uint32(nodesConfigProvider.ConsensusGroupSize(shardID))
	nbBlocksSelectedNodeInShard := nbBlocksPerShard * cGrShard / eligibleNodesPerShard

	var nbBlocksSelected uint32
	validators := state.NewShardValidatorsInfoMap()
	nbBlocksSelected = nbBlocksSelectedNodeInShard
	for i := uint32(0); i < eligibleNodesPerShard; i++ {
		str := fmt.Sprintf("rewardAddr%d_%d", shardID, i)
		addrHex := make([]byte, len(str)*2)
		_ = hex.Encode(addrHex, []byte(str))

		leaderSuccess := uint32(20)
		_ = validators.Add(&state.ValidatorInfo{
			PublicKey:                  []byte(fmt.Sprintf("pubKeyBLS%d%d", shardID, i)),
			ShardId:                    shardID,
			RewardAddress:              addrHex,
			LeaderSuccess:              leaderSuccess,
			ValidatorSuccess:           nbBlocksSelected - leaderSuccess,
			NumSelectedInSuccessBlocks: nbBlocksSelected,
			AccumulatedFees:            big.NewInt(int64(proposerFeesPerNode)),
			List:                       string(common.EligibleList),
		})
	}

	return validators
}

func TestSovereignRewards_CreateRewardsMiniBlocks(t *testing.T) {
	t.Parallel()

	args := getRewardsCreatorV2Arguments()
	nbEligiblePerShard := uint32(400)
	dummyRwd, _ := NewRewardsCreatorV2(args)
	vInfo := createSovereignDefaultValidatorInfo(nbEligiblePerShard, args.NodesConfigProvider, 100, defaultBlocksPerShard)

	nodesRewardInfo := dummyRwd.initNodesRewardsInfo(vInfo)
	_, _ = setDummyValuesInNodesRewardInfo(nodesRewardInfo, nbEligiblePerShard, tuStake, 0)

	args.StakingDataProvider = &stakingcommon.StakingDataProviderStub{
		GetTotalTopUpStakeEligibleNodesCalled: func() *big.Int {
			totalTopUpStake, _ := big.NewInt(0).SetString("3000000000000000000000000", 10)
			return totalTopUpStake
		},
		GetNodeStakedTopUpCalled: func(blsKey []byte) (*big.Int, error) {
			for shardID, vList := range vInfo.GetShardValidatorsInfoMap() {
				for i, v := range vList {
					if bytes.Equal(v.GetPublicKey(), blsKey) {
						return nodesRewardInfo[shardID][i].topUpStake, nil
					}
				}
			}
			return nil, fmt.Errorf("not found")
		},
	}
	rewardsShardedData := testscommon.NewShardedDataCacheNotifierMock()
	args.DataPool = &dataRetriever.PoolsHolderStub{
		RewardTransactionsCalled: func() retriever.ShardedDataCacherNotifier {
			return rewardsShardedData
		},
	}

	blocksPerShard := map[uint32]uint64{
		core.SovereignChainShardId: 14400,
	}

	args.EconomicsDataProvider.SetNumberOfBlocksPerShard(blocksPerShard)
	rewardsForBlocks, _ := big.NewInt(0).SetString("5000000000000000000000", 10)
	args.EconomicsDataProvider.SetRewardsToBeDistributedForBlocks(rewardsForBlocks)

	rwd, err := NewRewardsCreatorV2(args)
	require.Nil(t, err)
	require.NotNil(t, rwd)

	sovRwd, err := NewSovereignRewards(rwd)
	require.Nil(t, err)
	require.False(t, sovRwd.IsInterfaceNil())

	sovBlock := &block.SovereignChainHeader{
		EpochStart: block.EpochStartSovereign{
			Economics: getDefaultEpochStart().Economics,
		},
		Header:         &block.Header{},
		DevFeesInEpoch: big.NewInt(0),
	}

	var miniBlocks block.MiniBlockSlice
	miniBlocks, err = sovRwd.CreateRewardsMiniBlocks(sovBlock, vInfo, &sovBlock.EpochStart.Economics)
	require.Nil(t, err)
	require.Len(t, miniBlocks, 1)
	require.Len(t, miniBlocks[0].TxHashes, int(nbEligiblePerShard+1))

	numTxs := 0
	sumRewards := big.NewInt(0)
	var tx data.TransactionHandler
	for _, txHash := range miniBlocks[0].TxHashes {
		numTxs++
		tx, err = rwd.currTxs.GetTx(txHash)
		require.Nil(t, err)
		sumRewards.Add(sumRewards, tx.GetValue())
	}

	sumFees := big.NewInt(0)
	for _, v := range vInfo.GetAllValidatorsInfo() {
		sumFees.Add(sumFees, v.GetAccumulatedFees())
	}

	totalRws := rwd.economicsDataProvider.RewardsToBeDistributedForBlocks()
	rewardsForProtocolSustainability := big.NewInt(0).Set(sovBlock.EpochStart.Economics.RewardsForProtocolSustainability)
	expectedRewards := big.NewInt(0).Add(sumFees, totalRws)
	expectedRewards.Add(expectedRewards, rewardsForProtocolSustainability)
	require.Equal(t, expectedRewards, sumRewards)
	require.Len(t, rewardsShardedData.Keys(), 0)
	require.Equal(t, int(nbEligiblePerShard+1), numTxs)
}
