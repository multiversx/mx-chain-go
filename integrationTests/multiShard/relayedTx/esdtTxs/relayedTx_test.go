package esdtTxs

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/relayedTx"
	"github.com/ElrondNetwork/elrond-go/vm"
)

func TestRelayedTransactionInMultiShardEnvironmentWithESDTTX(t *testing.T) {
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

	///////////------- send token issue
	tokenName := "robertWhyNot"
	issuePrice := big.NewInt(1000)
	initalSupply := big.NewInt(10000000000)
	tokenIssuer := nodes[0]
	txData := "issue" + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(initalSupply.Bytes())
	integrationTests.CreateAndSendTransaction(tokenIssuer, issuePrice, vm.ESDTSCAddress, txData, integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := int64(10)
	for i := int64(0); i < nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

	relayedTx.CheckAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenName, initalSupply)

	/////////------ send tx to players
	valueToTopUp := big.NewInt(100000000)
	txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(valueToTopUp.Bytes())
	for _, player := range players {
		integrationTests.CreateAndSendTransaction(tokenIssuer, big.NewInt(0), player.Address, txData, integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	for i := int64(0); i < nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

	txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(sendValue.Bytes())
	transferTokenESDTGas := uint64(1)
	transferTokenBaseGas := tokenIssuer.EconomicsData.ComputeGasLimit(&transaction.Transaction{Data: []byte(txData)})
	transferTokenFullGas := transferTokenBaseGas + transferTokenESDTGas
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		for _, player := range players {
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, big.NewInt(0), transferTokenFullGas, []byte(txData))
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, big.NewInt(0), transferTokenFullGas, []byte(txData))
		}

		time.Sleep(time.Second)
	}

	nrRoundsToPropagateMultiShard = int64(20)
	for i := int64(0); i <= nrRoundsToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)
	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	relayedTx.CheckAddressHasESDTTokens(t, receiverAddress1, nodes, tokenName, finalBalance)
	relayedTx.CheckAddressHasESDTTokens(t, receiverAddress2, nodes, tokenName, finalBalance)

	players = append(players, relayer)
	relayedTx.CheckPlayerBalances(t, nodes, players)
}
