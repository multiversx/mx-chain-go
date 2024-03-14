package staking

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

// Test scenarios
// Do 3 stake transactions from 3 different wallets - tx value 2499, 2500, 2501
// testcase1 -- staking v3.5 --> tx1 fail, tx2 - node in queue, tx3 - node in queue with topUp 1
// testcase2 -- staking v4 step1 --> tx1 fail, tx2 - node in auction, tx3 - node in auction with topUp 1
// testcase3 -- staking v4 step2 --> tx1 fail, tx2 - node in auction, tx3 - node in auction with topUp 1
// testcase4 -- staking v3.step3 --> tx1 fail, tx2 - node in auction, tx3 - node in auction with topUp 1

// // Internal test scenario #3
func TestChainSimulator_SimpleStake(t *testing.T) {
	t.Run("staking ph 4 is not active", func(t *testing.T) {
		testChainSimulatorSimpleStake(t, 1, "queued")
	})

	t.Run("staking ph 4 step1", func(t *testing.T) {
		testChainSimulatorSimpleStake(t, 2, "auction")
	})

	t.Run("staking ph 4 step2", func(t *testing.T) {
		testChainSimulatorSimpleStake(t, 3, "auction")
	})

	t.Run("staking ph 4 step3", func(t *testing.T) {
		testChainSimulatorSimpleStake(t, 4, "auction")
	})
}

func testChainSimulatorSimpleStake(t *testing.T, targetEpoch int32, nodesStatus string) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	numOfShards := uint32(3)

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   false,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              numOfShards,
		GenesisTimestamp:         startTime,
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  3,
		NumNodesWaitingListShard: 3,
		AlterConfigsFunction: func(cfg *config.Configs) {
			configs.SetStakingV4ActivationEpochs(cfg, 2)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	mintValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(3000))
	wallet1, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)
	wallet2, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)
	wallet3, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)

	_, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(3)
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	dataFieldTx1 := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], mockBLSSignature)
	tx1Value := big.NewInt(0).Mul(big.NewInt(2499), oneEGLD)
	tx1 := generateTransaction(wallet1.Bytes, 0, vm.ValidatorSCAddress, tx1Value, dataFieldTx1, gasLimitForStakeOperation)

	dataFieldTx2 := fmt.Sprintf("stake@01@%s@%s", blsKeys[1], mockBLSSignature)
	tx2 := generateTransaction(wallet3.Bytes, 0, vm.ValidatorSCAddress, minimumStakeValue, dataFieldTx2, gasLimitForStakeOperation)

	dataFieldTx3 := fmt.Sprintf("stake@01@%s@%s", blsKeys[2], mockBLSSignature)
	tx3Value := big.NewInt(0).Mul(big.NewInt(2501), oneEGLD)
	tx3 := generateTransaction(wallet2.Bytes, 0, vm.ValidatorSCAddress, tx3Value, dataFieldTx3, gasLimitForStakeOperation)

	results, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{tx1, tx2, tx3}, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 3, len(results))
	require.NotNil(t, results)

	// tx1 should fail
	require.Equal(t, "insufficient stake value: expected 2500000000000000000000, got 2499000000000000000000", string(results[0].Logs.Events[0].Topics[1]))

	_ = cs.GenerateBlocks(1)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	if targetEpoch < 2 {
		bls1, _ := hex.DecodeString(blsKeys[1])
		bls2, _ := hex.DecodeString(blsKeys[2])

		blsKeyStatus := getBLSKeyStatus(t, metachainNode, bls1)
		require.Equal(t, nodesStatus, blsKeyStatus)

		blsKeyStatus = getBLSKeyStatus(t, metachainNode, bls2)
		require.Equal(t, nodesStatus, blsKeyStatus)
	} else {
		// tx2 -- validator should be in queue
		checkValidatorStatus(t, cs, blsKeys[1], nodesStatus)
		// tx3 -- validator should be in queue
		checkValidatorStatus(t, cs, blsKeys[2], nodesStatus)
	}
}

// Test auction list api calls during stakingV4 step 2 and onwards.
// Nodes configuration at genesis consisting of a total of 32 nodes, distributed on 3 shards + meta:
// - 4 eligible nodes/shard
// - 4 waiting nodes/shard
// - 2 nodes to shuffle per shard
// - max num nodes config for stakingV4 step3 = 24 (being downsized from previously 32 nodes)
// Steps:
// 1. Stake 1 node and check that in stakingV4 step1 it is unstaked
// 2. Re-stake the node to enter the auction list
// 3. From stakingV4 step2 onwards, check that api returns 8 qualified + 1 unqualified nodes
func TestChainSimulator_StakingV4Step2APICalls(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	stakingV4Step1Epoch := uint32(2)
	stakingV4Step2Epoch := uint32(3)
	stakingV4Step3Epoch := uint32(4)

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       time.Now().Unix(),
		RoundDurationInMillis:  uint64(6000),
		RoundsPerEpoch: core.OptionalUint64{
			HasValue: true,
			Value:    30,
		},
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         4,
		MetaChainMinNodes:        4,
		NumNodesWaitingListMeta:  4,
		NumNodesWaitingListShard: 4,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = stakingV4Step1Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = stakingV4Step2Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = stakingV4Step3Epoch

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].MaxNumNodes = 32
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].NodesToShufflePerShard = 2

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = stakingV4Step3Epoch
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].MaxNumNodes = 24
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].NodesToShufflePerShard = 2
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Add(minimumStakeValue, oneEGLD)
	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	// Stake a new validator that should end up in auction in step 1
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], mockBLSSignature)
	txStake := generateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, minimumStakeValue, txDataField, gasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = cs.GenerateBlocksUntilEpochIsReached(int32(stakingV4Step1Epoch))
	require.Nil(t, err)
	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// In step 1, only the previously staked node should be in auction list
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)
	require.Empty(t, auctionList)

	// re-stake the node
	txDataField = fmt.Sprintf("reStakeUnStakedNodes@%s", blsKeys[0])
	txReStake := generateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, big.NewInt(0), txDataField, gasLimitForStakeOperation)
	reStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txReStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, reStakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// after the re-stake process, the node should be in auction list
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err = metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)
	require.Equal(t, []*common.AuctionListValidatorAPIResponse{
		{
			Owner:          validatorOwner.Bech32,
			NumStakedNodes: 1,
			TotalTopUp:     "0",
			TopUpPerNode:   "0",
			QualifiedTopUp: "0",
			Nodes: []*common.AuctionNode{
				{
					BlsKey:    blsKeys[0],
					Qualified: true,
				},
			},
		},
	}, auctionList)

	// For steps 2,3 and onwards, when making API calls, we'll be using the api nodes config provider to mimic the max number of
	// nodes as it will be in step 3. This means we'll see the 8 nodes that were shuffled out from the eligible list,
	// plus the additional node that was staked manually.
	// Since those 8 shuffled out nodes will be replaced only with another 8 nodes, and the auction list size = 9,
	// the outcome should show 8 nodes qualifying and 1 node not qualifying
	for epochToSimulate := int32(stakingV4Step2Epoch); epochToSimulate < int32(stakingV4Step3Epoch)+3; epochToSimulate++ {
		err = cs.GenerateBlocksUntilEpochIsReached(epochToSimulate)
		require.Nil(t, err)
		err = cs.GenerateBlocks(2)
		require.Nil(t, err)

		numQualified, numUnQualified := getNumQualifiedAndUnqualified(t, metachainNode)
		require.Equal(t, 8, numQualified)
		require.Equal(t, 1, numUnQualified)
	}
}

func getNumQualifiedAndUnqualified(t *testing.T, metachainNode process.NodeHandler) (int, int) {
	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	numQualified := 0
	numUnQualified := 0

	for _, auctionOwnerData := range auctionList {
		for _, auctionNode := range auctionOwnerData.Nodes {
			if auctionNode.Qualified {
				numQualified++
			} else {
				numUnQualified++
			}
		}
	}

	return numQualified, numUnQualified
}
