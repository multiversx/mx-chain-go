package epochStartChangeWithoutTransactionInMultiShardedEnvironment

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/multiShard/endOfEpoch"
)

func TestEpochStartChangeWithoutTransactionInMultiShardedEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	enableEpochsConfig := config.EnableEpochs{
		StakingV2EnableEpoch:                   integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:         integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:   integrationTests.UnreachableEpoch,
		ConsensusPropagationChangesEnableEpoch: integrationTests.UnreachableEpoch,
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

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	time.Sleep(time.Second)

	// ----- wait for epoch end period
	round, nonce = endOfEpoch.CreateAndPropagateBlocks(t, roundsPerEpoch, round, nonce, nodes, idxProposers)

	nrRoundsToPropagateMultiShard := uint64(5)
	_, _ = endOfEpoch.CreateAndPropagateBlocks(t, nrRoundsToPropagateMultiShard, round, nonce, nodes, idxProposers)

	epoch := uint32(1)
	endOfEpoch.VerifyThatNodesHaveCorrectEpoch(t, epoch, nodes)
	endOfEpoch.VerifyIfAddedShardHeadersAreWithNewEpoch(t, nodes)
}
