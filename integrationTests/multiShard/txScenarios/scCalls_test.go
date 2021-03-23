package txScenarios

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/require"
)

func TestTransaction_TransactionSCScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	initialBalance := uint64(1000000000000)
	net := createGeneralTestnetForTxTest(t, initialBalance)
	defer net.Close()

	net.Increment()

	scPath := "./../../vm/arwen/testdata/counter/counter.wasm"
	scCode, err := ioutil.ReadFile(scPath)

	if err != nil {
		panic(fmt.Sprintf("cannotReadContractCode: %s", err))
	}

	players := net.Wallets

	// TODO rewrite the following lines with the TxDataBuilder when
	//	merged into development
	scCodeString := hex.EncodeToString(scCode)
	scCodeMetadataString := "0000"
	txData := []byte(scCodeString + "@" + hex.EncodeToString(factory.ArwenVirtualMachine) + "@" + scCodeMetadataString)

	// deploy contract insufficient gas limit
	player0 := players[0]
	tx := net.CreateSignedTxUint64(player0, net.DeploymentAddress, 0, txData)

	net.BypassErrorsOnce = true
	net.SendTx(tx)
	time.Sleep(time.Second)

	// Nonce decremented because the previous tx was deliberately wrong
	player0.Nonce--
	net.RequireWalletNoncesInSyncWithState()

	// contract data is not ok
	invalidContractData := make([]byte, len(txData))
	_ = copy(invalidContractData, txData)
	invalidContractData[123] = 50
	invalidContractData[124] = 51

	player2 := players[2]
	tx2 := net.CreateTxUint64(player2, net.DeploymentAddress, 0, txData)
	tx2.GasLimit = net.MaxGasLimit

	net.SignAndSendTx(player2, tx2)
	time.Sleep(time.Second)

	// contract deploy should work
	player4 := players[4]
	scAddress := net.NewAddress(player4)
	tx3 := net.CreateTxUint64(player4, net.DeploymentAddress, 0, txData)
	tx3.GasLimit = net.MaxGasLimit

	net.SignAndSendTx(player4, tx3)
	time.Sleep(time.Second)

	// delay for processing
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		net.Step()
		integrationTests.AddSelfNotarizedHeaderByMetachain(net.Nodes)
		time.Sleep(time.Second)
	}

	//check balance address that tried but failed to deploy contract should not be modified
	senderAccount := net.GetAccountHandler(player0.Address)
	require.Equal(t, initialBalance, senderAccount.GetBalance().Uint64())

	// check balance address that try to deploy contract should consume all the
	// gas provided
	txFee := net.ComputeTxFeeUint64(tx2)
	expectedBalance := initialBalance - txFee
	senderAccount = net.GetAccountHandler(player2.Address)
	require.Equal(t, expectedBalance, senderAccount.GetBalance().Uint64())

	// deploy should have worked: gas used should be greater than the gas limit computed from the transaction alone
	gasUsed := net.ComputeGasLimit(tx3)
	txFee = gasUsed * integrationTests.MinTxGasPrice
	expectedBalance = initialBalance - txFee

	senderAccount = net.GetAccountHandler(player4.Address)
	require.Greater(t, expectedBalance, senderAccount.GetBalance().Uint64())

	// do some smart contract call from shard 1 to shard 0
	// address is in shard 1
	sender := players[1]
	txData = []byte("increment")
	numIncrement := 10
	for i := 0; i < numIncrement; i++ {
		tx := net.CreateTxUint64(sender, scAddress, 0, txData)
		tx.GasLimit = net.MaxGasLimit

		net.SignTx(sender, tx)
		net.SendTxFromNode(tx, net.Nodes[1])
		time.Sleep(time.Millisecond)
	}

	// wait transaction to propagate cross shard
	for i := int64(0); i < 15; i++ {
		net.Step()
		integrationTests.AddSelfNotarizedHeaderByMetachain(net.Nodes)
		time.Sleep(time.Second)
	}

	// check value of counter from contract
	counterValue := vm.GetIntValueFromSC(nil, net.DefaultNode.AccntState, scAddress, "get", nil)
	require.Equal(t, int64(numIncrement+1), counterValue.Int64())

	net.RequireWalletNoncesInSyncWithState()
}
