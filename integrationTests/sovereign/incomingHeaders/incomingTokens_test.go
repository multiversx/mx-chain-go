package incomingHeaders

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/sovereign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSovereign_IncomingHeaderHandler(t *testing.T) {
	initialBalance := big.NewInt(1000000000000)
	nodes, idxProposers, players := sovereign.CreateGeneralSovereignSetup(initialBalance)
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	require.Equal(t, len(nodes), 3)

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// shard 1
	bechAddrShard1 := "erd1qhmhf5grwtep3n6ynkpz5u5lxw8n2s38yuq9ge8950lc0zqlwkfs3cus7a"
	bech32, err := pubkeyConverter.NewBech32PubkeyConverter(32, integrationTests.AddressHrp)
	assert.Nil(t, err)
	receiverAddress, _ := bech32.Decode(bechAddrShard1)

	senderShardID := nodes[0].ShardCoordinator.ComputeId(players[0].Address)
	assert.Equal(t, uint32(0), senderShardID)

	receiverShardID := nodes[0].ShardCoordinator.ComputeId(receiverAddress)
	assert.Equal(t, uint32(0), receiverShardID)

	integrationTests.CreateAndSendTransaction(nodes[0], nodes, sendValue, receiverAddress, "", integrationTests.MinTxGasLimit)
	time.Sleep(100 * time.Millisecond)

	//nrRoundsToTest := int64(4)
	//for i := int64(0); i < nrRoundsToTest; i++ {
	//	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	//	//integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	//
	//	time.Sleep(integrationTests.StepDelay)
	//}
	_ = idxProposers
}
