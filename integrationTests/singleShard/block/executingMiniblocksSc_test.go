package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/pkg/profile"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "agarioV2.hex"
var stepDelay = time.Second

func TestShouldProcessWithScTxsJoinAndRewardTheOwner(t *testing.T) {
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
	round = integrationTests.IncrementAndPrintRound(round)

	idxProposer := 0
	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	nodes[idxProposer].LoadTxSignSkBytes(hardCodedSk)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	integrationTests.MintAllNodes(nodes, initialVal)

	integrationTests.DeployScTx(nodes, idxProposer, string(scCode))
	integrationTests.ProposeBlock(nodes, []int{idxProposer}, round)
	integrationTests.SyncBlock(t, nodes, []int{idxProposer}, round)
	round = integrationTests.IncrementAndPrintRound(round)

	integrationTests.NodeJoinsGame(nodes, idxProposer, topUpValue, 0, hardCodedScResultingAddress)
	integrationTests.ProposeBlock(nodes, []int{idxProposer}, round)
	integrationTests.SyncBlock(t, nodes, []int{idxProposer}, round)
	round = integrationTests.IncrementAndPrintRound(round)

	integrationTests.CheckJoinGameIsDoneCorrectly(
		t,
		nodes,
		idxProposer,
		idxProposer,
		initialVal,
		topUpValue,
		hardCodedScResultingAddress,
	)

	integrationTests.NodeCallsRewardAndSend(nodes, idxProposer, idxProposer, withdrawValue, 0, hardCodedScResultingAddress)
	integrationTests.ProposeBlock(nodes, []int{idxProposer}, round)
	integrationTests.SyncBlock(t, nodes, []int{idxProposer}, round)
	round = integrationTests.IncrementAndPrintRound(round)

	integrationTests.CheckRewardIsDoneCorrectly(
		t,
		nodes,
		idxProposer,
		idxProposer,
		initialVal,
		topUpValue,
		withdrawValue,
		hardCodedScResultingAddress,
	)

	integrationTests.CheckRootHashes(t, nodes, []int{0})

	time.Sleep(1 * time.Second)
}

func TestProcessesJoinGameOf100PlayersRewardAndEndgame(t *testing.T) {
	t.Skip("this is a stress test for VM and AGAR.IO")

	stepDelay = time.Nanosecond

	p := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	defer p.Stop()

	log := logger.DefaultLogger()
	log.SetLevel(logger.LogDebug)

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	maxShards := uint32(1)
	numOfNodes := 1
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
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
	round = integrationTests.IncrementAndPrintRound(round)

	idxProposer := 0
	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	nodes[idxProposer].LoadTxSignSkBytes(hardCodedSk)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	integrationTests.MintAllNodes(nodes, initialVal)

	integrationTests.DeployScTx(nodes, idxProposer, string(scCode))
	integrationTests.ProposeBlock(nodes, []int{idxProposer}, round)
	integrationTests.SyncBlock(t, nodes, []int{idxProposer}, round)
	round = integrationTests.IncrementAndPrintRound(round)

	rMonitor := &statistics.ResourceMonitor{}
	fmt.Println(rMonitor.GenerateStatistics())

	for rr := 0; rr < 50; rr++ {
		for i := 0; i < 100; i++ {
			integrationTests.NodeJoinsGame(nodes, idxProposer, topUpValue, rr, hardCodedScResultingAddress)
		}
		time.Sleep(time.Second)

		startTime := time.Now()
		integrationTests.ProposeBlock(nodes, []int{idxProposer}, round)
		elapsedTime := time.Since(startTime)
		fmt.Printf("Block Created in %s\n", elapsedTime)
		round = integrationTests.IncrementAndPrintRound(round)

		integrationTests.NodeCallsRewardAndSend(nodes, idxProposer, idxProposer, withdrawValue, rr, hardCodedScResultingAddress)
		integrationTests.NodeEndGame(nodes, idxProposer, rr, hardCodedScResultingAddress)
		time.Sleep(time.Second)

		startTime = time.Now()
		integrationTests.ProposeBlock(nodes, []int{idxProposer}, round)
		elapsedTime = time.Since(startTime)
		fmt.Printf("Block Created in %s\n", elapsedTime)
		round = integrationTests.IncrementAndPrintRound(round)

		fmt.Println(rMonitor.GenerateStatistics())
	}

	integrationTests.CheckRootHashes(t, nodes, []int{0})

	time.Sleep(time.Second)
}
