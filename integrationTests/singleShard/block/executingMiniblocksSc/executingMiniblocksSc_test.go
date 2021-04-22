package executingMiniblocksSc

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/singleShard/block"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestShouldProcessMultipleERC20ContractsInSingleShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	scCode, err := ioutil.ReadFile("../../../vm/arwen/testdata/erc20-c-03/wrc20_arwen.wasm")
	assert.Nil(t, err)

	maxShards := uint32(1)
	numOfNodes := 2

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	connectableNodes := make([]integrationTests.Connectable, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(
			maxShards,
			0,
			0,
		)
		connectableNodes[i] = nodes[i]
	}

	integrationTests.ConnectNodes(connectableNodes)

	idxProposer := 0
	numPlayers := 10
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = integrationTests.CreateTestWalletAccount(nodes[idxProposer].ShardCoordinator, 0)
	}

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000005006c560111a94e434413c1cdaafbc3e1348947d1d5b3a1")
	nodes[idxProposer].LoadTxSignSkBytes(hardCodedSk)

	initialVal := big.NewInt(100000000000)
	integrationTests.MintAllNodes(nodes, initialVal)
	integrationTests.MintAllPlayers(nodes, players, initialVal)

	integrationTests.DeployScTx(nodes, idxProposer, hex.EncodeToString(scCode), factory.ArwenVirtualMachine, "001000000000")
	time.Sleep(block.StepDelay)
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, []int{idxProposer}, round, nonce)

	playersDoTopUp(nodes[idxProposer], players, hardCodedScResultingAddress, big.NewInt(10000000))
	time.Sleep(block.StepDelay)
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, []int{idxProposer}, round, nonce)

	for i := 0; i < 100; i++ {
		playersDoTransfer(nodes[idxProposer], players, hardCodedScResultingAddress, big.NewInt(100))
	}

	for i := 0; i < 10; i++ {
		time.Sleep(block.StepDelay)
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, []int{idxProposer}, round, nonce)
	}
	integrationTests.CheckRootHashes(t, nodes, []int{idxProposer})

	time.Sleep(1 * time.Second)
}

func playersDoTopUp(
	node *integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount,
	scAddress []byte,
	txValue *big.Int,
) {
	for _, player := range players {
		createAndSendTx(node, player, txValue, 20000, scAddress, []byte("topUp"))
	}
}

func playersDoTransfer(
	node *integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount,
	scAddress []byte,
	txValue *big.Int,
) {
	for _, playerToTransfer := range players {
		for _, player := range players {
			createAndSendTx(node, player, big.NewInt(0), 20000, scAddress,
				[]byte("transfer@"+hex.EncodeToString(playerToTransfer.Address)+"@"+hex.EncodeToString(txValue.Bytes())))
		}
	}
}

func createAndSendTx(
	node *integrationTests.TestProcessorNode,
	player *integrationTests.TestWalletAccount,
	txValue *big.Int,
	gasLimit uint64,
	rcvAddress []byte,
	txData []byte,
) {
	tx := &transaction.Transaction{
		Nonce:    player.Nonce,
		Value:    txValue,
		SndAddr:  player.Address,
		RcvAddr:  rcvAddress,
		Data:     txData,
		GasPrice: node.EconomicsData.GetMinGasPrice(),
		GasLimit: gasLimit,
	}

	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = player.SingleSigner.Sign(player.SkTxSign, txBuff)

	_, _ = node.SendTransaction(tx)
	player.Nonce++
}
