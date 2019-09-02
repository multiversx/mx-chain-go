package longTests

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/pkg/profile"
	"github.com/stretchr/testify/assert"
)

func TestProcessesJoinGameTheSamePlayerMultipleTimesRewardAndEndgameInMultipleRounds_WASMVM(t *testing.T) {
	t.Skip("this is a stress test for VM and AGAR.IO")

	p := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	defer p.Stop()

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
	numPlayers := 100
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

	numRounds := 10
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

	time.Sleep(time.Second)
}
