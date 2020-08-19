package txScenarios

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_TransactionMoveBalanceScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	initialBalance := big.NewInt(1000000000000)
	nodes, idxProposers, players, advertiser := createGeneralSetupForTxTest(initialBalance)
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

	// shard 1
	bechAddrShard1 := "erd1qhmhf5grwtep3n6ynkpz5u5lxw8n2s38yuq9ge8950lc0zqlwkfs3cus7a"
	bech32, _ := pubkeyConverter.NewBech32PubkeyConverter(32)
	receiverAddress, _ := bech32.Decode(bechAddrShard1)

	senderShardID := nodes[0].ShardCoordinator.ComputeId(players[0].Address)
	assert.Equal(t, uint32(0), senderShardID)

	receiverShardID := nodes[0].ShardCoordinator.ComputeId(receiverAddress)
	assert.Equal(t, uint32(1), receiverShardID)

	// move balance cross shard
	tx1 := createAndSendTransaction(nodes[0], players[0], receiverAddress, sendValue, []byte(""), integrationTests.MinTxGasPrice, integrationTests.MinTxGasLimit)
	time.Sleep(100 * time.Millisecond)
	// move balance self should consume gas
	tx2 := createAndSendTransaction(nodes[0], players[2], players[2].Address, sendValue, []byte(""), integrationTests.MinTxGasPrice, integrationTests.MinTxGasLimit)
	time.Sleep(100 * time.Millisecond)
	// move balance insufficient gas
	_ = createAndSendTransaction(nodes[1], players[1], players[1].Address, sendValue, []byte(""), integrationTests.MinTxGasPrice, integrationTests.MinTxGasLimit-1)
	time.Sleep(100 * time.Millisecond)
	// move balance wrong gas price
	_ = createAndSendTransaction(nodes[1], players[3], players[3].Address, sendValue, []byte(""), integrationTests.MinTxGasPrice-1, integrationTests.MinTxGasLimit)
	time.Sleep(100 * time.Millisecond)
	// send value to staking contract without data should consume gas
	tx3 := createAndSendTransaction(nodes[0], players[4], vm.AuctionSCAddress, sendValue, []byte(""), integrationTests.MinTxGasPrice, integrationTests.MinTxGasLimit)
	time.Sleep(100 * time.Millisecond)
	// send value to staking contract with data should consume gas
	txData := []byte("contract@qwt")
	gasLimitTxWithData := nodes[0].EconomicsData.MaxGasLimitPerBlock(0) - 1
	_ = createAndSendTransaction(nodes[0], players[6], vm.AuctionSCAddress, sendValue, txData, integrationTests.MinTxGasPrice, gasLimitTxWithData)
	time.Sleep(100 * time.Millisecond)

	nrRoundsToTest := int64(7)
	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	// check sender balance intra shard transaction (receiver is self)
	gasUsed := nodes[0].EconomicsData.ComputeGasLimit(tx2)
	txFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(gasUsed), big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	expectedBalance := big.NewInt(0).Sub(initialBalance, txFee)

	senderAccount := getUserAccount(nodes, players[2].Address)
	assert.Equal(t, players[2].Nonce, senderAccount.GetNonce())
	assert.Equal(t, expectedBalance, senderAccount.GetBalance())

	//check balance intra shard tx insufficient gas limit
	senderAccount = getUserAccount(nodes, players[1].Address)
	assert.Equal(t, uint64(0), senderAccount.GetNonce())
	assert.Equal(t, initialBalance, senderAccount.GetBalance())

	// check balance invalid gas price
	senderAccount = getUserAccount(nodes, players[3].Address)
	assert.Equal(t, uint64(0), senderAccount.GetNonce())
	assert.Equal(t, initialBalance, senderAccount.GetBalance())

	// check sender balance cross shard transaction
	gasUsed = nodes[0].EconomicsData.ComputeGasLimit(tx1)
	txFee = big.NewInt(0).Mul(big.NewInt(0).SetUint64(gasUsed), big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	expectedBalance = big.NewInt(0).Sub(initialBalance, sendValue)
	expectedBalance.Sub(expectedBalance, txFee)
	senderAccount = getUserAccount(nodes, players[0].Address)
	assert.Equal(t, players[0].Nonce, senderAccount.GetNonce())
	assert.Equal(t, expectedBalance, senderAccount.GetBalance())

	// check balance account that send money to AuctionSC
	gasUsed = nodes[0].EconomicsData.ComputeGasLimit(tx3)
	txFee = big.NewInt(0).Mul(big.NewInt(0).SetUint64(gasUsed), big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	senderAccount = getUserAccount(nodes, players[4].Address)
	assert.Equal(t, players[4].Nonce, senderAccount.GetNonce())
	expectedBalance = big.NewInt(0).Sub(initialBalance, txFee)
	assert.Equal(t, expectedBalance, senderAccount.GetBalance())

	// check balance account that send money to AuctionSC with data field
	txFee = big.NewInt(0).Mul(big.NewInt(0).SetUint64(gasLimitTxWithData), big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	senderAccount = getUserAccount(nodes, players[6].Address)
	expectedBalance = big.NewInt(0).Sub(initialBalance, txFee)
	assert.Equal(t, players[6].Nonce, senderAccount.GetNonce())
	assert.Equal(t, expectedBalance, senderAccount.GetBalance())

	roundToPropagateMultiShard := int64(15)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)

	// check balance account that send money to AuctionSC with data field should refund money back
	senderAccount = getUserAccount(nodes, players[6].Address)
	expectedBalance = big.NewInt(0).Sub(initialBalance, txFee)
	assert.Equal(t, players[6].Nonce, senderAccount.GetNonce())
	assert.Equal(t, expectedBalance, senderAccount.GetBalance())

	// check balance receiver cross shard transaction
	receiver1 := getUserAccount(nodes, receiverAddress)
	assert.Equal(t, sendValue, receiver1.GetBalance())
}
