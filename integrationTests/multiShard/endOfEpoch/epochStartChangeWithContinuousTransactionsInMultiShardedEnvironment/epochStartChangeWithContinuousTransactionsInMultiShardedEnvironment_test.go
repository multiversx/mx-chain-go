package epochStartChangeWithContinuousTransactionsInMultiShardedEnvironment

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
)

func TestEpochStartChangeWithContinuousTransactionsInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 3

	enableEpochsConfig := config.EnableEpochs{
		StakingV2EnableEpoch:                 integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:       integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch: integrationTests.UnreachableEpoch,
	}

	nodes := integrationTests.CreateNodesWithEnableEpochs(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochsConfig,
	)

	roundsPerEpoch := uint64(10)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000)
	sendValue := big.NewInt(5)
	integrationTests.MintAllNodes(nodes, initialVal)
	receiverAddress := []byte("12345678901234567890123456789012")

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	// ----- wait for epoch end period
	epoch := uint32(2)
	nrRoundsToPropagateMultiShard := uint64(5)
	for i := uint64(0); i <= (uint64(epoch)*roundsPerEpoch)+nrRoundsToPropagateMultiShard; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		for _, node := range nodes {
			integrationTests.CreateAndSendTransaction(node, nodes, sendValue, receiverAddress, "", integrationTests.AdditionalGasLimit)
		}

		time.Sleep(integrationTests.StepDelay)
	}

	time.Sleep(time.Second)

	endOfEpoch.VerifyThatNodesHaveCorrectEpoch(t, epoch, nodes)
	endOfEpoch.VerifyIfAddedShardHeadersAreWithNewEpoch(t, nodes)
}
