package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "../../agarioV3.hex"
var stepDelay = time.Second

func TestShouldProcessWithScTxsJoinAndRewardOneRound(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	log := logger.DefaultLogger()
	log.SetLevel(logger.LogDebug)

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	maxShards := uint32(1)
	numOfNodes := 4
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
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
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	nodes[idxProposer].LoadTxSignSkBytes(hardCodedSk)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	integrationTests.MintAllNodes(nodes, initialVal)
	integrationTests.MintAllPlayers(nodes, players, initialVal)

	integrationTests.DeployScTx(nodes, idxProposer, string(scCode))
	time.Sleep(stepDelay)
	integrationTests.ProposeBlock(nodes, []int{idxProposer}, round, nonce)
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
				strconv.Itoa(currentRound),
				hardCodedScResultingAddress,
			)
		}

		// waiting to disseminate transactions
		time.Sleep(stepDelay)

		round, nonce = integrationTests.ProposeAndSyncBlocks(t, nodes, idxProposers, round, nonce)

		integrationTests.CheckJoinGame(t, nodes, players, topUpValue, idxProposers[0], hardCodedScResultingAddress)

		for i := 0; i < numRewardedPlayers; i++ {
			integrationTests.NodeCallsRewardAndSend(nodes, idxProposers[0], players[i], withdrawValues[i], strconv.Itoa(currentRound), hardCodedScResultingAddress)
		}

		// waiting to disseminate transactions
		time.Sleep(stepDelay)

		round, nonce = integrationTests.ProposeAndSyncBlocks(t, nodes, idxProposers, round, nonce)

		integrationTests.CheckRewardsDistribution(t, nodes, players, topUpValue, totalWithdrawValue,
			hardCodedScResultingAddress, idxProposers[0])

		fmt.Println(rMonitor.GenerateStatistics())
	}
}
