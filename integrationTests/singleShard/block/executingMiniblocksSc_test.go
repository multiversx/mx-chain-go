package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "../../agar_v1_min.hex"
var stepDelay = time.Second
var p2pBootstrapDelay = time.Second * 12

func TestShouldProcessWithScTxsJoinAndRewardOneRound(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	maxShards := uint32(1)
	numOfNodes := 4
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(
			maxShards,
			0,
			0,
			advertiserAddr,
		)
		nodes[i].EconomicsData.SetMinGasPrice(0)
	}

	idxProposer := 0
	numPlayers := 10
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = integrationTests.CreateTestWalletAccount(nodes[idxProposer].ShardCoordinator, 0)
	}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000001006c560111a94e434413c1cdaafbc3e1348947d1d5b3a1")
	nodes[idxProposer].LoadTxSignSkBytes(hardCodedSk)

	initialVal := big.NewInt(10000000000)
	topUpValue := big.NewInt(500)
	integrationTests.MintAllNodes(nodes, initialVal)
	integrationTests.MintAllPlayers(nodes, players, initialVal)

	integrationTests.DeployScTx(nodes, idxProposer, string(scCode), factory.IELEVirtualMachine)
	time.Sleep(stepDelay)
	integrationTests.ProposeBlock(nodes, []int{idxProposer}, round, nonce)
	time.Sleep(stepDelay)
	integrationTests.SyncBlock(t, nodes, []int{idxProposer}, round)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	numRounds := 1
	runMultipleRoundsOfTheGame(
		t,
		numRounds,
		numPlayers,
		nodes,
		players,
		topUpValue,
		hardCodedScResultingAddress,
		round,
		nonce,
		[]int{idxProposer},
	)

	integrationTests.CheckRootHashes(t, nodes, []int{idxProposer})

	time.Sleep(1 * time.Second)
}

func getPercentageOfValue(value *big.Int, percentage float64) *big.Int {
	x := new(big.Float).SetInt(value)
	y := big.NewFloat(percentage)

	z := new(big.Float).Mul(x, y)

	op := big.NewInt(0)
	result, _ := z.Int(op)

	return result
}

func runMultipleRoundsOfTheGame(
	t *testing.T,
	nrRounds, numPlayers int,
	nodes []*integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount,
	topUpValue *big.Int,
	hardCodedScResultingAddress []byte,
	round uint64,
	nonce uint64,
	idxProposers []int,
) {
	rMonitor := &statistics.ResourceMonitor{}
	numRewardedPlayers := 10
	if numRewardedPlayers > numPlayers {
		numRewardedPlayers = numPlayers
	}

	totalWithdrawValue := big.NewInt(0).SetUint64(topUpValue.Uint64() * uint64(len(players)))
	withdrawValues := make([]*big.Int, numRewardedPlayers)
	winnerRate := 1.0 - 0.05*float64(numRewardedPlayers-1)
	withdrawValues[0] = big.NewInt(0).Set(getPercentageOfValue(totalWithdrawValue, winnerRate))
	for i := 1; i < numRewardedPlayers; i++ {
		withdrawValues[i] = big.NewInt(0).Set(getPercentageOfValue(totalWithdrawValue, 0.05))
	}

	for currentRound := 0; currentRound < nrRounds; currentRound++ {
		for _, player := range players {
			integrationTests.PlayerJoinsGame(
				nodes,
				player,
				topUpValue,
				int32(currentRound),
				hardCodedScResultingAddress,
			)
		}

		// waiting to disseminate transactions
		time.Sleep(stepDelay)
		round, nonce = integrationTests.ProposeAndSyncBlocks(t, nodes, idxProposers, round, nonce)

		integrationTests.CheckJoinGame(t, nodes, players, topUpValue, idxProposers[0], hardCodedScResultingAddress)

		for i := 0; i < numRewardedPlayers; i++ {
			integrationTests.NodeCallsRewardAndSend(nodes, idxProposers[0], players[i], withdrawValues[i], int32(currentRound), hardCodedScResultingAddress)
		}

		// waiting to disseminate transactions
		time.Sleep(stepDelay)
		round, nonce = integrationTests.ProposeAndSyncBlocks(t, nodes, idxProposers, round, nonce)

		time.Sleep(stepDelay)
		integrationTests.CheckRewardsDistribution(t, nodes, players, topUpValue, totalWithdrawValue,
			hardCodedScResultingAddress, idxProposers[0])

		fmt.Println(rMonitor.GenerateStatistics())
	}
}

func TestShouldProcessMultipleERC20ContractsInSingleShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:DEBUG")

	scCode, err := ioutil.ReadFile("./wrc20_arwen_01.wasm")
	assert.Nil(t, err)

	maxShards := uint32(1)
	numOfNodes := 2
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(
			maxShards,
			0,
			0,
			advertiserAddr,
		)
	}

	idxProposer := 0
	numPlayers := 10
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = integrationTests.CreateTestWalletAccount(nodes[idxProposer].ShardCoordinator, 0)
	}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

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

	integrationTests.DeployScTx(nodes, idxProposer, hex.EncodeToString(scCode), factory.ArwenVirtualMachine)
	time.Sleep(stepDelay)
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, []int{idxProposer}, round, nonce)

	playersDoTopUp(nodes[idxProposer], players, hardCodedScResultingAddress, big.NewInt(10000000))
	time.Sleep(stepDelay)
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, []int{idxProposer}, round, nonce)

	for i := 0; i < 100; i++ {
		playersDoTransfer(nodes[idxProposer], players, hardCodedScResultingAddress, big.NewInt(100))
	}

	for i := 0; i < 10; i++ {
		time.Sleep(stepDelay)
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
				[]byte("transfer@"+hex.EncodeToString(playerToTransfer.Address.Bytes())+"@"+hex.EncodeToString(txValue.Bytes())))
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
		SndAddr:  player.Address.Bytes(),
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
