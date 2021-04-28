package txScenarios

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_TransactionBuiltinFunctionsScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	initialBalance := big.NewInt(1000000000000)
	nodes, idxProposers, players := createGeneralSetupForTxTest(initialBalance)
	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scPath := "./../../vm/arwen/testdata/counter/counter.wasm"
	scCode, err := ioutil.ReadFile(scPath)

	if err != nil {
		panic(fmt.Sprintf("cannotReadContractCode: %s", err))
	}

	scCodeString := hex.EncodeToString(scCode)
	scCodeMetadataString := "0000"

	txData := []byte(scCodeString + "@" + hex.EncodeToString(factory.ArwenVirtualMachine) + "@" + scCodeMetadataString)
	// contract deploy should work
	scAddressBytes, _ := nodes[0].BlockchainHook.NewAddress(players[0].Address, 0, factory.ArwenVirtualMachine)
	_ = createAndSendTransaction(nodes[0], players[0], make([]byte, 32), big.NewInt(0), txData, integrationTests.MinTxGasPrice, nodes[0].EconomicsData.MaxGasLimitPerBlock(0)-1)
	time.Sleep(time.Second)

	nrRoundsToTest := int64(5)

	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	// invalid function intra-shard should consume gas
	sender := players[2]
	txData = []byte("invalidFunction")
	gasLimit := nodes[0].EconomicsData.MaxGasLimitPerBlock(0) - 1
	tx2 := createAndSendTransaction(nodes[0], sender, scAddressBytes, big.NewInt(0), txData, integrationTests.MinTxGasPrice, gasLimit)
	time.Sleep(time.Millisecond)

	// invalid function cross-shard should consume gas
	_ = createAndSendTransaction(nodes[1], players[1], scAddressBytes, big.NewInt(0), txData, integrationTests.MinTxGasPrice, gasLimit)
	time.Sleep(time.Millisecond)

	// change owner address wrong caller should not work
	newOwnerAddress := []byte("12345678123456781234567812345678")
	txData = []byte("ChangeOwnerAddress" + "@" + hex.EncodeToString(newOwnerAddress))
	_ = createAndSendTransaction(nodes[0], players[4], scAddressBytes, big.NewInt(0), txData, integrationTests.MinTxGasPrice, gasLimit)
	time.Sleep(time.Millisecond)

	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	// check balance after a call of invalid function intra-shard
	senderAccount := getUserAccount(nodes, players[2].Address)
	assert.Equal(t, players[2].Nonce, senderAccount.GetNonce())
	txFee := nodes[0].EconomicsData.ComputeTxFee(tx2)
	assert.Equal(t, big.NewInt(0).Sub(initialBalance, txFee), senderAccount.GetBalance())

	// check balance after a call of invalid function cross-shard
	senderAccount = getUserAccount(nodes, players[1].Address)
	assert.Equal(t, players[1].Nonce, senderAccount.GetNonce())
	txFee = nodes[0].EconomicsData.ComputeTxFee(tx2)
	assert.Equal(t, big.NewInt(0).Sub(initialBalance, txFee), senderAccount.GetBalance())

	// check owner address should not change
	account := getUserAccount(nodes, scAddressBytes)
	assert.NotEqual(t, newOwnerAddress, account.GetOwnerAddress())

	// change reward address from correct address
	newOwnerAddress = []byte("12345678123456781234567812345678")
	txData = []byte("ChangeOwnerAddress" + "@" + hex.EncodeToString(newOwnerAddress))
	_ = createAndSendTransaction(nodes[0], players[0], scAddressBytes, big.NewInt(0), txData, integrationTests.MinTxGasPrice, gasLimit)
	time.Sleep(time.Millisecond)

	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		time.Sleep(time.Second)
	}

	// check owner address should work
	account = getUserAccount(nodes, scAddressBytes)
	initialOwnerAccount := getUserAccount(nodes, players[0].Address)
	assert.Equal(t, players[0].Nonce, initialOwnerAccount.GetNonce())
	assert.Equal(t, newOwnerAddress, account.GetOwnerAddress())
}
