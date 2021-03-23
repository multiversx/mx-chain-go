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
	"github.com/stretchr/testify/require"
)

func TestTransactionSigning(t *testing.T) {
	initialBalance := uint64(1000000000000)
	net := createGeneralTestnetForTxTest(t, initialBalance)
	defer net.Close()

	net.Increment()

	sender := net.Wallets[0]
	receiver := sender
	tx := net.CreateTxUint64(sender, receiver.Address, 0, []byte("some data"))
	tx.GasLimit = net.MaxGasLimit

	fmt.Println(tx)

	err := net.DefaultNode.Node.ValidateTransaction(tx)
	require.Nil(t, err)
}

func TestTransactionSigning2(t *testing.T) {
	initialBalance := big.NewInt(1000000000000)
	nodes, _, players, advertiser := createGeneralSetupForTxTest(initialBalance)
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	txData := []byte("some data")
	tx := createUserTx(
		players[0],
		players[0].Address,
		big.NewInt(0),
		txData,
		integrationTests.MinTxGasPrice,
		nodes[0].EconomicsData.MaxGasLimitPerBlock(0)-1)

	fmt.Println(tx)
	err := nodes[0].Node.ValidateTransaction(tx)
	require.Nil(t, err)
}

func TestTransactionSigning3(t *testing.T) {
	txData := []byte("some data")
	initialBalance := big.NewInt(1000000000000)

	// init old network ===================
	nodes, _, players, advertiser := createGeneralSetupForTxTest(initialBalance)
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// init TestNetwork ===================
	net := createGeneralTestnetForTxTest(t, initialBalance.Uint64())
	defer net.Close()

	net.Increment()

	// create old-style tx ===================
	oldTxByOldWallet := createUserTx(players[0], players[0].Address, big.NewInt(0), txData, integrationTests.MinTxGasPrice, nodes[0].EconomicsData.MaxGasLimitPerBlock(0)-1)
	players[0].Nonce--
	oldTxByNewWallet := createUserTx(net.Wallets[0], net.Wallets[0].Address, big.NewInt(0), txData, integrationTests.MinTxGasPrice, nodes[0].EconomicsData.MaxGasLimitPerBlock(0)-1)
	net.Wallets[0].Nonce--

	// create new-style tx ==============
	newTxByOldWallet := net.CreateTxUint64(players[0], players[0].Address, 0, txData)
	players[0].Nonce--
	newTxByOldWallet.GasLimit = net.MaxGasLimit

	newTxByNewWallet := net.CreateTxUint64(net.Wallets[0], net.Wallets[0].Address, 0, txData)
	newTxByNewWallet.GasLimit = net.MaxGasLimit
	net.Wallets[0].Nonce--

	// transaction comparations: tx using old vs new wallets, created by old network
	expectedForSigning, err := oldTxByOldWallet.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	require.Nil(t, err)
	actualForSigning, err := newTxByOldWallet.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	require.Nil(t, err)
	require.Equal(t, expectedForSigning, actualForSigning)

	signatureByOldWallet, err := players[0].SingleSigner.Sign(players[0].SkTxSign, actualForSigning)

	// transaction comparations: tx using old vs new wallets, created by new network
	expectedForSigning, err = oldTxByNewWallet.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	require.Nil(t, err)
	actualForSigning, err = newTxByNewWallet.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	require.Nil(t, err)
	require.Equal(t, expectedForSigning, actualForSigning)

	signatureByNewWallet, err := net.Wallets[0].SingleSigner.Sign(net.Wallets[0].SkTxSign, actualForSigning)

	// transaction comparations: signatures
	require.Equal(t, signatureByOldWallet, oldTxByOldWallet.Signature)
	require.Equal(t, signatureByOldWallet, newTxByOldWallet.Signature)
	require.Equal(t, signatureByNewWallet, oldTxByNewWallet.Signature)
	require.Equal(t, signatureByNewWallet, newTxByNewWallet.Signature)

	// old net validates old tx
	err = nodes[0].Node.ValidateTransaction(oldTxByOldWallet)
	assert.Nil(t, err)

	// new net validates old tx
	err = net.DefaultNode.Node.ValidateTransaction(oldTxByNewWallet)
	assert.Nil(t, err)

	// old net validates new tx
	err = nodes[0].Node.ValidateTransaction(newTxByOldWallet)
	assert.Nil(t, err)

	// new net validates new tx
	err = net.DefaultNode.Node.ValidateTransaction(newTxByNewWallet)
	assert.Nil(t, err)
}

func TestTransaction_TransactionSCScenarios_Rewritten(t *testing.T) {
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

	// // deploy contract insufficient gas limit
	player0 := players[0]
	// tx := net.CreateTxUint64(player0, net.DeploymentAddress, 0, txData)

	// net.BypassErrorsOnce = true
	// net.SendTx(tx)
	// time.Sleep(time.Second)

	// // Nonce decremented because the previous tx was deliberately wrong
	// player0.Nonce--
	// net.RequireWalletNoncesInSyncWithState()

	// contract data is not ok
	invalidContractData := make([]byte, len(txData))
	_ = copy(invalidContractData, txData)
	invalidContractData[123] = 50
	invalidContractData[124] = 51

	player2 := players[2]
	tx2 := net.CreateTxUint64(player2, net.DeploymentAddress, 0, txData)
	tx2.GasLimit = net.MaxGasLimit

	err = net.DefaultNode.Node.ValidateTransaction(tx2)
	require.Nil(t, err)

	net.SendTx(tx2)
	time.Sleep(time.Second)

	// contract deploy should work
	player4 := players[4]
	scAddress := net.NewAddress(player4)
	tx3 := net.CreateTxUint64(player4, net.DeploymentAddress, 0, txData)
	tx3.GasLimit = net.MaxGasLimit

	net.SendTx(tx3)
	time.Sleep(time.Second)

	// delay for processing
	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		net.Step()
		integrationTests.AddSelfNotarizedHeaderByMetachain(net.Nodes)
		time.Sleep(time.Second)
	}

	//check balance address that try to deploy contract should not be modified
	senderAccount := net.GetAccountHandler(player0.Address)
	require.Equal(t, uint64(0), senderAccount.GetNonce())
	require.Equal(t, initialBalance, senderAccount.GetBalance())

	//check balance address that try to deploy contract should consume all the
	//gas provided
	txFee := net.ComputeTxFeeUint64(tx2)
	expectedBalance := initialBalance - txFee
	senderAccount = net.GetAccountHandler(player2.Address)
	require.Equal(t, player2.Nonce, senderAccount.GetNonce())
	require.Equal(t, expectedBalance, senderAccount.GetBalance().Int64())

	//deploy should work gas used should be greater than estimation and small that all gas provided
	gasUsed := net.ComputeGasLimit(tx3)
	txFee = gasUsed * integrationTests.MinTxGasPrice
	expectedBalance = initialBalance - txFee

	senderAccount = net.GetAccountHandler(player4.Address)
	require.Equal(t, player4.Nonce, senderAccount.GetNonce())
	require.Equal(t, expectedBalance, senderAccount.GetBalance().Uint64())

	// do some smart contract call from shard 1 to shard 0
	// address is in shard 1
	sender := players[1]
	txData = []byte("increment")
	numIncrement := 10
	for i := 0; i < numIncrement; i++ {
		tx := net.CreateTxUint64(sender, scAddress, 0, txData)
		tx.GasLimit = net.MaxGasLimit

		net.SendTxFromNode(tx, net.Nodes[1])
		time.Sleep(time.Millisecond)
	}

	// wait transaction to propagate cross shard
	for i := int64(0); i < 15; i++ {
		net.Step()
		integrationTests.AddSelfNotarizedHeaderByMetachain(net.Nodes)
		time.Sleep(time.Second)
	}

	// check account nonce that called increment method
	senderAccount = net.GetAccountHandler(players[1].Address)
	require.Equal(t, players[1].Nonce, senderAccount.GetNonce())

	// check value of counter from contract
	counterValue := vm.GetIntValueFromSC(nil, net.DefaultNode.AccntState, scAddress, "get", nil)
	require.Equal(t, int64(numIncrement+1), counterValue.Int64())
}

func TestTransaction_TransactionSCScenarios(t *testing.T) {
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
