package smartContract

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	factory2 "github.com/ElrondNetwork/elrond-go/vm/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScDeployAndChangeScOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(1000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	firstSCOwner := nodes[0].OwnAccount.Address.Bytes()

	// deploy the smart contracts
	firstSCAddress := putDeploySCToDataPool("./testdata/counter.wasm", firstSCOwner, 0, big.NewInt(50), nodes)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
	integrationTests.SyncBlock(t, nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// make smart contract call to shard 1 which will do in shard 0
	for _, node := range nodes {
		txData := "increment"
		for i := 0; i < 10; i++ {
			integrationTests.CreateAndSendTransaction(node, big.NewInt(0), firstSCAddress, txData)
		}
	}

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 15
	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	address, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(firstSCAddress)
	shId := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != shId {
			continue
		}

		numCalled := vm.GetIntValueFromSC(nil, node.AccntState, firstSCAddress, "get", nil)
		require.NotNil(t, numCalled)
	}

	account := getAccountFromAddrBytes(nodes[0].AccntState, nodes[0].OwnAccount.Address.Bytes())
	require.Equal(t, big.NewInt(0), account.DeveloperReward)

	newOwnerAddress := []byte("12345678123456781234567812345678")
	txData := "ChangeOwnerAddress" + "@" + hex.EncodeToString(newOwnerAddress)
	integrationTests.CreateAndSendTransaction(nodes[0], big.NewInt(0), firstSCAddress, txData)

	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	// check new owner address is set
	account = getAccountFromAddrBytes(nodes[0].AccntState, firstSCAddress)
	require.Equal(t, newOwnerAddress, account.OwnerAddress)
	require.True(t, account.DeveloperReward.Cmp(big.NewInt(0)) == 1)
}

func TestScDeployAndClaimSmartContractDeveloperRewards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(1000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	firstSCOwner := nodes[0].OwnAccount.Address.Bytes()

	// deploy the smart contracts
	firstSCAddress := putDeploySCToDataPool("./testdata/counter.wasm", firstSCOwner, 0, big.NewInt(50), nodes)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
	integrationTests.SyncBlock(t, nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// make smart contract call to shard 1 which will do in shard 0
	for _, node := range nodes {
		txData := "increment"
		for i := 0; i < 10; i++ {
			integrationTests.CreateAndSendTransaction(node, big.NewInt(0), firstSCAddress, txData)
		}
	}

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 15
	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	address, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(firstSCAddress)
	shId := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != shId {
			continue
		}

		numCalled := vm.GetIntValueFromSC(nil, node.AccntState, firstSCAddress, "get", nil)
		require.NotNil(t, numCalled)
	}

	account := getAccountFromAddrBytes(nodes[0].AccntState, nodes[0].OwnAccount.Address.Bytes())
	require.Equal(t, big.NewInt(0), account.DeveloperReward)
	fmt.Println("smart contract owner before claim", account.Balance)
	oldOwnerBalance := big.NewInt(0).Set(account.Balance)

	account = getAccountFromAddrBytes(nodes[0].AccntState, firstSCAddress)
	fmt.Println("smart contract rewards balance", account.DeveloperReward)

	for _, node := range nodes {
		node.EconomicsData.SetGasPerDataByte(0)
		node.EconomicsData.SetMinGasLimit(0)
		node.EconomicsData.SetMinGasPrice(0)
	}

	txData := "ClaimDeveloperRewards"
	integrationTests.CreateAndSendTransaction(nodes[0], big.NewInt(0), firstSCAddress, txData)

	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	account = getAccountFromAddrBytes(nodes[0].AccntState, nodes[0].OwnAccount.Address.Bytes())
	fmt.Println("smart contract owner after claim", account.Balance)
	require.True(t, account.Balance.Cmp(oldOwnerBalance) == 1)
}

func getAccountFromAddrBytes(accState state.AccountsAdapter, address []byte) *state.Account {
	addrCont, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(address)
	sndrAcc, _ := accState.GetExistingAccount(addrCont)

	sndAccSt, _ := sndrAcc.(*state.Account)

	return sndAccSt
}

func TestSCCallingInCrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(1000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// mint smart contract holders
	firstSCOwner := []byte("12345678901234567890123456789000")
	secondSCOwner := []byte("99945678901234567890123456789001")

	mintPubKey(firstSCOwner, initialVal, nodes)
	mintPubKey(secondSCOwner, initialVal, nodes)

	// deploy the smart contracts
	firstSCAddress := putDeploySCToDataPool("./testdata/first/first.wasm", firstSCOwner, 0, big.NewInt(50), nodes)
	//000000000000000005005d3d53b5d0fcf07d222170978932166ee9f3972d3030
	secondSCAddress := putDeploySCToDataPool("./testdata/second/second.wasm", secondSCOwner, 0, big.NewInt(50), nodes)
	//00000000000000000500017cc09151c48b99e2a1522fb70a5118ad4cb26c3031

	integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
	integrationTests.SyncBlock(t, nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// make smart contract call to shard 1 which will do in shard 0
	for _, node := range nodes {
		txData := "doSomething"
		integrationTests.CreateAndSendTransaction(node, big.NewInt(50), secondSCAddress, txData)
	}

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	// verify how many times was shard 0 and shard 1 called
	address, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(firstSCAddress)
	shId := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != shId {
			continue
		}

		numCalled := vm.GetIntValueFromSC(nil, node.AccntState, firstSCAddress, "numCalled", nil)
		assert.Equal(t, uint64(len(nodes)), numCalled.Uint64())
	}
}

func TestSCCallingInCrossShardDelegation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:INFO,process/smartcontract:DEBUG")

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 3
	shardConsensusGroupSize := 2
	metaConsensusGroupSize := 2

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		numMetachainNodes,
		numOfShards,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	nodes := make([]*integrationTests.TestProcessorNode, 0)

	for _, nds := range nodesMap {
		nodes = append(nodes, nds...)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(1000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// mint smart contract holders
	delegateSCOwner := []byte("12345678901234567890123456789002")
	stakerBLSKey, _ := hex.DecodeString(strings.Repeat("a", 256))

	mintPubKey(delegateSCOwner, initialVal, nodes)

	// deploy the smart contracts
	delegateSCAddress := putDeploySCToDataPool("./testdata/delegate/delegate.wasm", delegateSCOwner, 0, big.NewInt(50), nodes)

	integrationTests.UpdateRound(nodes, round)
	integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
	integrationTests.SyncBlock(t, nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// one node calls to stake all the money from the delegation - that's how the contract is :D
	node := nodes[0]
	txData := "sendToStaking"
	integrationTests.CreateAndSendTransaction(node, node.EconomicsData.GenesisNodePrice(), delegateSCAddress, txData)

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	time.Sleep(time.Second)
	// verify system smart contract has the value
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != sharding.MetachainShardId {
			continue
		}
		scQuery := &process.SCQuery{
			ScAddress: factory2.StakingSCAddress,
			FuncName:  "isStaked",
			Arguments: [][]byte{stakerBLSKey},
		}
		vmOutput, _ := node.SCQueryService.ExecuteQuery(scQuery)

		assert.NotNil(t, vmOutput)
		if vmOutput != nil {
			assert.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
		}
	}
}

func putDeploySCToDataPool(
	fileName string,
	pubkey []byte,
	nonce uint64,
	transferOnDeploy *big.Int,
	nodes []*integrationTests.TestProcessorNode,
) []byte {
	scCode, _ := ioutil.ReadFile(fileName)
	scCodeString := hex.EncodeToString(scCode)

	blockChainHook := nodes[0].BlockchainHook

	scAddressBytes, _ := blockChainHook.NewAddress(pubkey, nonce, factory.ArwenVirtualMachine)

	tx := &transaction.Transaction{
		Nonce:    nonce,
		Value:    transferOnDeploy,
		RcvAddr:  make([]byte, 32),
		SndAddr:  pubkey,
		GasPrice: nodes[0].EconomicsData.GetMinGasPrice(),
		GasLimit: nodes[0].EconomicsData.MaxGasLimitPerBlock() - 1,
		Data:     []byte(scCodeString + "@" + hex.EncodeToString(factory.ArwenVirtualMachine)),
	}
	txHash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, tx)

	address, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(pubkey)
	shId := nodes[0].ShardCoordinator.ComputeId(address)

	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != shId {
			continue
		}
		strCache := process.ShardCacherIdentifier(shId, shId)
		node.DataPool.Transactions().AddData(txHash, tx, strCache)
	}

	return scAddressBytes
}

func mintPubKey(
	pubkey []byte,
	initialVal *big.Int,
	nodes []*integrationTests.TestProcessorNode,
) {
	address, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(pubkey)
	shId := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != shId {
			continue
		}
		integrationTests.MintAddress(node.AccntState, pubkey, initialVal)
	}
}
