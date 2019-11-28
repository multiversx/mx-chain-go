package smartContract

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestSCCallingInCrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:INFO,process/smartcontract:DEBUG")

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

	initialVal := big.NewInt(10000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// mint smart contract holders
	firstSCOwner := []byte("12345678901234567890123456789000")
	secondSCOwner := []byte("99945678901234567890123456789001")
	delegateSCOwner := []byte("12345678901234567890123456789002")

	mintPubKey(firstSCOwner, initialVal, nodes)
	mintPubKey(secondSCOwner, initialVal, nodes)
	mintPubKey(delegateSCOwner, initialVal, nodes)

	// deploy the smart contracts
	firstSCAddress := putDeploySCToDataPool("./testdata/first/first.wasm", firstSCOwner, 0, big.NewInt(50), nodes)
	//000000000000000005005d3d53b5d0fcf07d222170978932166ee9f3972d3030
	secondSCAddress := putDeploySCToDataPool("./testdata/second/second.wasm", secondSCOwner, 0, big.NewInt(50), nodes)
	//00000000000000000500017cc09151c48b99e2a1522fb70a5118ad4cb26c3031
	//delegateSCAddress := putDeploySCToDataPool("./testdata/delegate/delegate.wasm", delegateSCOwner, 0, big.NewInt(50), nodes)

	fmt.Println(firstSCAddress)
	fmt.Println(secondSCAddress)
	//fmt.Println(delegateSCAddress)

	integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
	integrationTests.SyncBlock(t, nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// make smart contract call to shard 1 which will do in shard 0
	for _, node := range nodes {
		//pubKey, _ := node.NodeKeys.Pk.ToByteArray()
		txData := "doSomething" //"main" + "@" + hex.EncodeToString(pubKey)
		integrationTests.CreateAndSendTransaction(node, big.NewInt(50), secondSCAddress, txData)
	}

	// make nodes delegate stake to delegateSCAddress
	//for _, node := range nodes {
	//	pubKey, _ := node.NodeKeys.Pk.ToByteArray()
	//	txData := "delegate" + "@" + hex.EncodeToString(pubKey)
	//	integrationTests.CreateAndSendTransaction(node, node.EconomicsData.StakeValue(), delegateSCAddress, txData)
	//}

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 7
	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
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

	// one node calls to stake all the money from the delegation - that's how the contract is :D
	//node := nodes[0]
	//txData := "sendToStaking"
	//integrationTests.CreateAndSendTransaction(node, big.NewInt(100), delegateSCAddress, txData)

	time.Sleep(time.Second)

	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	// verify system smart contract has the value
	/*for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != sharding.MetachainShardId {
			continue
		}
		scQuery := &process.SCQuery{
			ScAddress: factory2.StakingSCAddress,
			FuncName:  "isStaked",
			Arguments: [][]byte{delegateSCAddress},
		}
		vmOutput, _ := node.SCQueryService.ExecuteQuery(scQuery)
		assert.NotNil(t, vmOutput)
		assert.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
	}*/
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
	fmt.Println(hex.EncodeToString(scAddressBytes))
	fmt.Println(scAddressBytes)
	tx := &transaction.Transaction{
		Nonce:    nonce,
		Value:    transferOnDeploy,
		RcvAddr:  make([]byte, 32),
		SndAddr:  pubkey,
		GasPrice: nodes[0].EconomicsData.GetMinGasPrice(),
		GasLimit: nodes[0].EconomicsData.MaxGasLimitPerBlock() - 1,
		Data:     scCodeString + "@" + hex.EncodeToString(factory.ArwenVirtualMachine),
	}
	txHash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, tx)

	address, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(pubkey)
	shId := nodes[0].ShardCoordinator.ComputeId(address)
	fmt.Printf("resulting shard ID %d \n", shId)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != shId {
			continue
		}
		strCache := process.ShardCacherIdentifier(shId, shId)
		node.ShardDataPool.Transactions().AddData(txHash, tx, strCache)
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

//[0 0 0 0 0 0 0 0 5 0 93 61 83 181 208 252 240 125 34 33 112 151 137 50 22 110 233 243 151 45 48 48]
//resulting shard ID 0
//00000000000000000500017cc09151c48b99e2a1522fb70a5118ad4cb26c3031
//[0 0 0 0 0 0 0 0 5 0 1 124 192 145 81 196 139 153 226 161 82 47 183 10 81 24 173 76 178 108 48 49]
