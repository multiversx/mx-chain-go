package edgecases

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/multiShard/relayedTx"
	"github.com/stretchr/testify/assert"
)

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTxButWrongNonceShouldNotIncrementUserAccNonce(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, leaders, players, relayer := relayedTx.CreateGeneralSetupForRelayTxTest(false)
	defer func() {
		for _, n := range nodes {
			n.Close()
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
			player.Nonce += 1
			_, _ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			_, _ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, integrationTests.MinTxGasLimit, []byte(""))
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, leaders, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, leaders, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	receiver1 := relayedTx.GetUserAccount(nodes, receiverAddress1)
	receiver2 := relayedTx.GetUserAccount(nodes, receiverAddress2)

	assert.True(t, check.IfNil(receiver1))
	assert.True(t, check.IfNil(receiver2))

	for _, player := range players {
		account := relayedTx.GetUserAccount(nodes, player.Address)
		assert.True(t, account.GetBalance().Cmp(big.NewInt(0)) == 0)
		assert.Equal(t, uint64(0), account.GetNonce())
	}

	relayerAccount := relayedTx.GetUserAccount(nodes, relayer.Address)
	assert.True(t, relayerAccount.GetBalance().Cmp(relayer.Balance) == 0)
}

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTxButWithTooMuchGas(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, leaders, players, relayer := relayedTx.CreateGeneralSetupForRelayTxTest(false)
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	additionalGasLimit := uint64(100000)
	tooMuchGasLimit := integrationTests.MinTxGasLimit + additionalGasLimit
	nrRoundsToTest := int64(5)

	txsSentEachRound := big.NewInt(2) // 2 relayed txs each round
	txsSentPerPlayer := big.NewInt(0).Mul(txsSentEachRound, big.NewInt(nrRoundsToTest))
	initialPlayerFunds := big.NewInt(0).Mul(sendValue, txsSentPerPlayer)
	integrationTests.MintAllPlayers(nodes, players, initialPlayerFunds)

	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			_, _ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, tooMuchGasLimit, []byte(""))
			_, _ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, tooMuchGasLimit, []byte(""))
		}

		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, leaders, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, leaders, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	receiver1 := relayedTx.GetUserAccount(nodes, receiverAddress1)
	receiver2 := relayedTx.GetUserAccount(nodes, receiverAddress2)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	assert.Equal(t, 0, receiver1.GetBalance().Cmp(finalBalance))
	assert.Equal(t, 0, receiver2.GetBalance().Cmp(finalBalance))

	players = append(players, relayer)
	checkPlayerBalancesWithPenalization(t, nodes, players)
}

func checkPlayerBalancesWithPenalization(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount,
) {

	for i := 0; i < len(players); i++ {
		userAcc := relayedTx.GetUserAccount(nodes, players[i].Address)
		assert.Equal(t, 0, userAcc.GetBalance().Cmp(players[i].Balance))
		assert.Equal(t, userAcc.GetNonce(), players[i].Nonce)
	}
}
