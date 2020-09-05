package edgecases

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/relayedTx"
	"github.com/stretchr/testify/assert"
)

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTxButWrongNonce(t *testing.T) {
	//TODO fix this test: task EN-7532
	t.Skip("fixme")

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

	totalFees := big.NewInt(0)
	relayerInitialValue := big.NewInt(0).Set(relayer.Balance)
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			player.Nonce += 1
			relayerTx := relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			totalFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(relayerTx.GetGasPrice()), big.NewInt(0).SetUint64(relayerTx.GetGasLimit()))
			totalFees.Add(totalFees, totalFee)
			relayerTx = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			totalFee = big.NewInt(0).Mul(big.NewInt(0).SetUint64(relayerTx.GetGasPrice()), big.NewInt(0).SetUint64(relayerTx.GetGasLimit()))
			totalFees.Add(totalFees, totalFee)
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

	assert.True(t, check.IfNil(receiver1))
	assert.True(t, check.IfNil(receiver2))

	for _, player := range players {
		account := relayedTx.GetUserAccount(nodes, player.Address)
		assert.True(t, account.GetBalance().Cmp(big.NewInt(0)) == 0)
		assert.Equal(t, uint64(nrRoundsToTest)*2, account.GetNonce())
	}

	expectedBalance := big.NewInt(0).Sub(relayerInitialValue, totalFees)
	relayerAccount := relayedTx.GetUserAccount(nodes, relayer.Address)
	assert.True(t, relayerAccount.GetBalance().Cmp(expectedBalance) == 0)
}

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTxButWithTooMuchGas(t *testing.T) {
	//TODO fix this test: task EN-7532
	t.Skip("fixme")

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

	additionalGasLimit := uint64(100000)
	tooMuchGasLimit := integrationTests.MinTxGasLimit + additionalGasLimit
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		for _, player := range players {
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, tooMuchGasLimit, []byte(""))
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, tooMuchGasLimit, []byte(""))
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
	additionalCost := big.NewInt(0).Mul(
		big.NewInt(0).SetUint64(additionalGasLimit*uint64(nrRoundsToTest)),
		big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice),
	)
	checkPlayerBalancesWithPenalization(t, nodes, players, additionalCost)
}

func checkPlayerBalancesWithPenalization(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount,
	additionalCost *big.Int,
) {

	for i := 0; i < len(players)-1; i++ {
		userAcc := relayedTx.GetUserAccount(nodes, players[i].Address)
		assert.Equal(t, userAcc.GetBalance().Cmp(players[i].Balance.Add(players[i].Balance, additionalCost)), 0)
		assert.Equal(t, userAcc.GetNonce(), players[i].Nonce)
	}

	if len(players) > 0 {
		relayer := players[len(players)-1]
		userAcc := relayedTx.GetUserAccount(nodes, relayer.Address)
		assert.Equal(t, userAcc.GetBalance().Cmp(relayer.Balance), 0)
		assert.Equal(t, userAcc.GetNonce(), relayer.Nonce)
	}
}
