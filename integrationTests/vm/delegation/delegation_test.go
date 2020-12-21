package delegation

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
	integrationTestsVm "github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
)

func TestDelegationSystemSCWithValidatorStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 1
	shardConsensusGroupSize := 1
	metaConsensusGroupSize := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorAndTxKeys(
		nodesPerShard,
		numMetachainNodes,
		numOfShards,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	nodes := make([]*integrationTests.TestProcessorNode, 0)
	idxProposers := make([]int, numOfShards+1)

	for _, nds := range nodesMap {
		nodes = append(nodes, nds...)
	}

	for _, nds := range nodesMap {
		idx, err := integrationTestsVm.GetNodeIndex(nodes, nds[0])
		assert.Nil(t, err)

		idxProposers = append(idxProposers, idx)
	}
	integrationTests.DisplayAndStartNodes(nodes)

	roundsPerEpoch := uint64(5)
	for _, node := range nodes {
		node.InitDelegationManager()
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, node := range nodesMap {
		fmt.Println(integrationTests.MakeDisplayTable(node))
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	rewardAddress := createNewDelegationSystemSC(nodes[0], nodes)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	round, nonce = processBlocks(t, round, nonce, 1, nodesMap)

	for _, node := range nodes {
		txData := "changeRewardAddress" + "@" + hex.EncodeToString(rewardAddress)
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), vm.ValidatorSCAddress, txData, core.MinMetaTxExtraGasCost)
		delegateToSystemSC(node, nodes, rewardAddress, big.NewInt(1000000))
	}
	time.Sleep(time.Second)

	epochs := uint32(2)
	nbBlocksToProduce := (roundsPerEpoch+1)*uint64(epochs) + 1

	round, nonce = processBlocks(t, round, nonce, nbBlocksToProduce, nodesMap)

	checkRewardsUpdatedInDelegationSC(t, nodes, rewardAddress, epochs)

	balancesBeforeClaimRewards := getNodesBalances(nodes)
	balanceToConsumeForGas := core.SafeMul(integrationTests.MinTxGasPrice, core.MinMetaTxExtraGasCost)
	for i, node := range nodes {
		txData := "claimRewards"
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), rewardAddress, txData, core.MinMetaTxExtraGasCost)
		balancesBeforeClaimRewards[i].Sub(balancesBeforeClaimRewards[i], balanceToConsumeForGas)
	}
	time.Sleep(time.Second)

	round, nonce = processBlocks(t, round, nonce, 15, nodesMap)
	balancesAfterClaimRewards := getNodesBalances(nodes)

	for i := 0; i < len(balancesAfterClaimRewards); i++ {
		assert.True(t, balancesAfterClaimRewards[i].Cmp(balancesBeforeClaimRewards[i]) > 0)
	}
}

func getNodesBalances(nodes []*integrationTests.TestProcessorNode) []*big.Int {
	balances := make([]*big.Int, 0, len(nodes))
	for _, node := range nodes {
		shardIDForAddress := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)
		nodeInShard := getNodeWithShardID(nodes, shardIDForAddress)

		acc, _ := nodeInShard.AccntState.GetExistingAccount(node.OwnAccount.Address)
		userAcc, _ := acc.(state.UserAccountHandler)
		balances = append(balances, userAcc.GetBalance())
	}
	return balances
}

func processBlocks(
	t *testing.T,
	round, nonce uint64,
	blockToProduce uint64,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
) (uint64, uint64) {
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode

	for i := uint64(0); i < blockToProduce; i++ {
		for _, nodesSlice := range nodesMap {
			integrationTests.UpdateRound(nodesSlice, round)
			integrationTests.AddSelfNotarizedHeaderByMetachain(nodesSlice)
		}

		_, _, consensusNodes = integrationTests.AllShardsProposeBlock(round, nonce, nodesMap)
		indexesProposers := endOfEpoch.GetBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++

		time.Sleep(time.Second)
	}

	return round, nonce
}

func getNodeWithShardID(nodes []*integrationTests.TestProcessorNode, shardId uint32) *integrationTests.TestProcessorNode {
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == shardId {
			return node
		}
	}
	return nil
}

func checkRewardsUpdatedInDelegationSC(t *testing.T, nodes []*integrationTests.TestProcessorNode, address []byte, lastEpoch uint32) {
	node := getNodeWithShardID(nodes, core.MetachainShardId)

	systemVM, _ := node.VMContainer.Get(factory.SystemVirtualMachine)
	for i := uint32(1); i <= lastEpoch; i++ {
		vmInput := &vmcommon.ContractCallInput{
			VMInput: vmcommon.VMInput{
				CallerAddr:  vm.EndOfEpochAddress,
				Arguments:   [][]byte{big.NewInt(int64(i)).Bytes()},
				CallValue:   big.NewInt(0),
				GasProvided: 1000000,
			},
			RecipientAddr: address,
			Function:      "getRewardData",
		}

		vmOutput, err := systemVM.RunSmartContractCall(vmInput)
		assert.Nil(t, err)
		assert.NotNil(t, vmOutput)

		assert.Equal(t, len(vmOutput.ReturnData), 3)
		rwdInBigInt := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
		assert.True(t, rwdInBigInt.Cmp(big.NewInt(0)) > 0)
	}
}

func createNewDelegationSystemSC(
	node *integrationTests.TestProcessorNode,
	nodes []*integrationTests.TestProcessorNode,
) []byte {
	txData := "createNewDelegationContract" + "@00@00"
	integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(10000), vm.DelegationManagerSCAddress, txData, core.MinMetaTxExtraGasCost)

	rewardAddress := make([]byte, len(vm.FirstDelegationSCAddress))
	copy(rewardAddress, vm.FirstDelegationSCAddress)
	rewardAddress[28] = 2
	time.Sleep(time.Second)
	return rewardAddress
}

func delegateToSystemSC(
	node *integrationTests.TestProcessorNode,
	nodes []*integrationTests.TestProcessorNode,
	address []byte,
	value *big.Int,
) {
	txData := "delegate"
	integrationTests.CreateAndSendTransaction(node, nodes, value, address, txData, core.MinMetaTxExtraGasCost)
}
