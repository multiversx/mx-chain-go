//go:build !race
// +build !race

package delegation

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/multiShard/endOfEpoch"
	integrationTestsVm "github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelegationSystemSCWithValidatorStatisticsAndStakingPhase3p5(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 1
	shardConsensusGroupSize := 1
	metaConsensusGroupSize := 1

	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorAndTxKeys(
		nodesPerShard,
		numMetachainNodes,
		numOfShards,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
	)

	nodes := make([]*integrationTests.TestProcessorNode, 0)

	for _, nds := range nodesMap {
		nodes = append(nodes, nds...)
	}

	for _, nds := range nodesMap {
		_, err := integrationTestsVm.GetNodeIndex(nodes, nds[0])
		assert.Nil(t, err)
	}
	integrationTests.DisplayAndStartNodes(nodes)

	roundsPerEpoch := uint64(5)
	for _, node := range nodes {
		node.InitDelegationManager()
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	round, nonce = processBlocks(t, round, nonce, roundsPerEpoch, nodesMap)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	for _, node := range nodesMap {
		fmt.Println(integrationTests.MakeDisplayTable(node))
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	nodeIndexForDelegationOwner := 0
	delegationAddress := createNewDelegationSystemSC(nodes[nodeIndexForDelegationOwner], nodes)

	round, nonce = processBlocks(t, round, nonce, 1, nodesMap)

	for index, node := range nodes {
		if index == nodeIndexForDelegationOwner {
			round, nonce = doMergeValidatorToDelegationSameOwner(t, delegationAddress, node, nodes, nodesMap, round, nonce)
			continue
		}

		round, nonce = doMergeValidatorToDelegationWithWhitelist(t, delegationAddress, node, nodes, nodesMap, nodeIndexForDelegationOwner, round, nonce)
	}
	time.Sleep(time.Second)

	epochs := uint32(2)
	nbBlocksToProduce := (roundsPerEpoch+1)*uint64(epochs) + 1

	round, nonce = processBlocks(t, round, nonce, nbBlocksToProduce, nodesMap)

	lastEpoch := round / (roundsPerEpoch + 1)
	checkRewardsUpdatedInDelegationSC(t, nodes, delegationAddress, uint32(lastEpoch))

	balancesBeforeClaimRewards := getNodesBalances(nodes)
	balanceToConsumeForGas := core.SafeMul(integrationTests.MinTxGasPrice, core.MinMetaTxExtraGasCost)
	for i, node := range nodes {
		txData := "claimRewards"
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), delegationAddress, txData, core.MinMetaTxExtraGasCost)
		balancesBeforeClaimRewards[i].Sub(balancesBeforeClaimRewards[i], balanceToConsumeForGas)
	}
	time.Sleep(time.Second)

	_, _ = processBlocks(t, round, nonce, 15, nodesMap)
	balancesAfterClaimRewards := getNodesBalances(nodes)

	for i := 0; i < len(balancesAfterClaimRewards); i++ {
		assert.True(t, balancesAfterClaimRewards[i].Cmp(balancesBeforeClaimRewards[i]) > 0)
	}

	delegationMgr := getUserAccount(nodes, vm.DelegationManagerSCAddress)
	assert.Equal(t, delegationMgr.GetBalance(), big.NewInt(0))
}

func TestDelegationSystemSCWithValidatorStatisticsAndStakingPhase3p5_CreateStakingProviderFromExistingValidator(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 1
	shardConsensusGroupSize := 1
	metaConsensusGroupSize := 1

	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorAndTxKeys(
		nodesPerShard,
		numMetachainNodes,
		numOfShards,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
	)

	nodes := make([]*integrationTests.TestProcessorNode, 0)

	for _, nds := range nodesMap {
		nodes = append(nodes, nds...)
	}

	for _, nds := range nodesMap {
		_, err := integrationTestsVm.GetNodeIndex(nodes, nds[0])
		assert.Nil(t, err)
	}
	integrationTests.DisplayAndStartNodes(nodes)

	roundsPerEpoch := uint64(5)
	for _, node := range nodes {
		node.InitDelegationManager()
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	round, nonce = processBlocks(t, round, nonce, roundsPerEpoch, nodesMap)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	for _, node := range nodesMap {
		fmt.Println(integrationTests.MakeDisplayTable(node))
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	nodeIndexForDelegationOwner := 0
	nodeToBeConverted := nodes[nodeIndexForDelegationOwner]

	delegationAddress := convertValidatorToDelegationSystemSC(nodeToBeConverted, nodes)
	time.Sleep(time.Second)

	epochs := uint32(2)
	nbBlocksToProduce := (roundsPerEpoch+1)*uint64(epochs) + 1

	_, _ = processBlocks(t, round, nonce, nbBlocksToProduce, nodesMap)

	delegationMgr := getUserAccount(nodes, vm.DelegationManagerSCAddress)
	assert.Equal(t, delegationMgr.GetBalance(), big.NewInt(0))

	delegationAccount := getUserAccount(nodes, delegationAddress)
	assert.Equal(t, nodeToBeConverted.OwnAccount.Address, delegationAccount.GetOwnerAddress())
}

func doMergeValidatorToDelegationSameOwner(
	t *testing.T,
	delegationAddress []byte,
	node *integrationTests.TestProcessorNode,
	nodes []*integrationTests.TestProcessorNode,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
	round, nonce uint64,
) (uint64, uint64) {
	numBlocksToProduce := uint64(3)
	txDataFieldBuilder := txDataBuilder.NewBuilder()
	txDataFieldBuilder.Func("mergeValidatorToDelegationSameOwner").Bytes(delegationAddress)
	integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), vm.DelegationManagerSCAddress, txDataFieldBuilder.ToString(), core.MinMetaTxExtraGasCost)

	return processBlocks(t, round, nonce, numBlocksToProduce, nodesMap)
}

func doMergeValidatorToDelegationWithWhitelist(
	t *testing.T,
	delegationAddress []byte,
	node *integrationTests.TestProcessorNode,
	nodes []*integrationTests.TestProcessorNode,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
	nodeIndexForDelegationOwner int,
	round, nonce uint64,
) (uint64, uint64) {
	numBlocksToProduce := uint64(3)
	txDataFieldBuilder := txDataBuilder.NewBuilder()
	txDataFieldBuilder.Func("whitelistForMerge").Bytes(node.OwnAccount.PkTxSignBytes)
	integrationTests.CreateAndSendTransaction(nodes[nodeIndexForDelegationOwner], nodes, big.NewInt(0), delegationAddress, txDataFieldBuilder.ToString(), core.MinMetaTxExtraGasCost)

	round, nonce = processBlocks(t, round, nonce, numBlocksToProduce, nodesMap)

	txDataFieldBuilder.Clear()
	txDataFieldBuilder.Func("mergeValidatorToDelegationWithWhitelist").Bytes(delegationAddress)
	integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), vm.DelegationManagerSCAddress, txDataFieldBuilder.ToString(), core.MinMetaTxExtraGasCost)

	round, nonce = processBlocks(t, round, nonce, numBlocksToProduce, nodesMap)

	delegateToSystemSC(node, nodes, delegationAddress, big.NewInt(1000000))

	return round, nonce
}

func getUserAccount(nodes []*integrationTests.TestProcessorNode, address []byte) state.UserAccountHandler {
	shardIDForAddress := nodes[0].ShardCoordinator.ComputeId(address)
	nodeInShard := getNodeWithShardID(nodes, shardIDForAddress)

	acc, _ := nodeInShard.AccntState.GetExistingAccount(address)
	userAcc, _ := acc.(state.UserAccountHandler)
	return userAcc
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
	for i := uint32(2); i <= lastEpoch; i++ {
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
		require.Nil(t, err)
		require.NotNil(t, vmOutput)

		require.Equal(t, 3, len(vmOutput.ReturnData))
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

	time.Sleep(time.Second)
	return generateSecondDelegationAddress()
}

// generateSecondDelegationAddress will generate the address of the second delegation address (the exact next one after vm.FirstDelegationSCAddress)
// by copying the vm.FirstDelegationSCAddress bytes and altering the corresponding position (28-th) to 2
func generateSecondDelegationAddress() []byte {
	address := make([]byte, len(vm.FirstDelegationSCAddress))
	copy(address, vm.FirstDelegationSCAddress)
	address[28] = 2

	return address
}

func convertValidatorToDelegationSystemSC(
	node *integrationTests.TestProcessorNode,
	nodes []*integrationTests.TestProcessorNode,
) []byte {
	txData := "makeNewContractFromValidatorData" + "@00@00"
	integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), vm.DelegationManagerSCAddress, txData, core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	return generateSecondDelegationAddress()
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
