package txScenarios

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_TransactionSCScenarios(t *testing.T) {
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

	// deploy contract insufficient gas limit
	_ = createAndSendTransaction(nodes[0], players[0], make([]byte, 32), big.NewInt(0), txData, integrationTests.MinTxGasPrice, integrationTests.MinTxGasLimit)
	time.Sleep(time.Second)

	// contract data is not ok
	invalidContractData := make([]byte, len(txData))
	_ = copy(invalidContractData, txData)
	invalidContractData[123] = 50
	invalidContractData[124] = 51
	tx2 := createAndSendTransaction(nodes[0], players[2], make([]byte, 32), big.NewInt(0), invalidContractData, integrationTests.MinTxGasPrice, nodes[0].EconomicsData.MaxGasLimitPerBlock(0)-1)
	time.Sleep(time.Second)

	// contract deploy should work
	scAddressBytes, _ := nodes[0].BlockchainHook.NewAddress(players[4].Address, 0, factory.ArwenVirtualMachine)
	tx3 := createAndSendTransaction(nodes[0], players[4], make([]byte, 32), big.NewInt(0), txData, integrationTests.MinTxGasPrice, nodes[0].EconomicsData.MaxGasLimitPerBlock(0)-1)
	time.Sleep(time.Second)

	nrRoundsToTest := int64(5)

	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	//check balance address that try to deploy contract should not be modified
	senderAccount := getUserAccount(nodes, players[0].Address)
	assert.Equal(t, uint64(0), senderAccount.GetNonce())
	assert.Equal(t, initialBalance, senderAccount.GetBalance())

	//check balance address that try to deploy contract should consume all the gas provided
	txFee := nodes[0].EconomicsData.ComputeTxFee(tx2)
	expectedBalance := big.NewInt(0).Sub(initialBalance, txFee)
	senderAccount = getUserAccount(nodes, players[2].Address)
	assert.Equal(t, players[2].Nonce, senderAccount.GetNonce())
	assert.Equal(t, expectedBalance, senderAccount.GetBalance())

	//deploy should work gas used should be greater than estimation and small that all gas provided
	gasUsed := nodes[0].EconomicsData.ComputeGasLimit(tx3)
	txFee = big.NewInt(0).Mul(big.NewInt(0).SetUint64(gasUsed), big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice))
	expectedBalance = big.NewInt(0).Sub(initialBalance, txFee)

	senderAccount = getUserAccount(nodes, players[4].Address)
	assert.Equal(t, players[4].Nonce, senderAccount.GetNonce())
	assert.True(t, expectedBalance.Cmp(senderAccount.GetBalance()) == 1)

	// do some smart contract call from shard 1 to shard 0
	// address is in shard 1
	sender := players[1]
	numIncrement := 10
	for i := 0; i < numIncrement; i++ {
		txData = []byte("increment")
		_ = createAndSendTransaction(nodes[1], sender, scAddressBytes, big.NewInt(0), txData, integrationTests.MinTxGasPrice, nodes[0].EconomicsData.MaxGasLimitPerBlock(0)-1)
		time.Sleep(time.Millisecond)
	}

	// wait transaction to propagate cross shard
	for i := int64(0); i < 15; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		time.Sleep(time.Second)
	}

	// check account nonce that called increment method
	senderAccount = getUserAccount(nodes, players[1].Address)
	assert.Equal(t, players[1].Nonce, senderAccount.GetNonce())

	// check value of counter from contract
	counterValue := vm.GetIntValueFromSC(nil, nodes[0].AccntState, scAddressBytes, "get", nil)
	assert.Equal(t, int64(numIncrement+1), counterValue.Int64())
}
