package normalTxs

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/relayedTx"
	"github.com/stretchr/testify/assert"
)

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := relayedTx.CreateGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, integrationTests.MinTxGasLimit, []byte(""))
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	receiver1 := relayedTx.GetUserAccount(nodes, receiverAddress1)
	receiver2 := relayedTx.GetUserAccount(nodes, receiverAddress2)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	assert.Equal(t, receiver1.GetBalance().Cmp(finalBalance), 0)
	assert.Equal(t, receiver2.GetBalance().Cmp(finalBalance), 0)

	players = append(players, relayer)
	relayedTx.CheckPlayerBalances(t, nodes, players)
}
