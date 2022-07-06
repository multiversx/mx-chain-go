//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Arwen fix

package txsFee

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestAsyncESDTCallShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)

	// create an address with ESDT token
	sndAddr := []byte("12345678901234567890123456789012")

	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)

	// deploy 2 contracts
	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/second-contract.wasm", ownerAccount, gasPrice, deployGasLimit, argsSecond, big.NewInt(0))

	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	firstSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/first-contract.wasm", ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(0))

	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasLimit := uint64(500000)
	tx := utils.CreateESDTTransferTx(0, sndAddr, firstSCAddress, token, big.NewInt(5000), gasPrice, gasLimit)
	tx.Data = []byte(string(tx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractHalf")))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckESDTBalance(t, testContext, firstSCAddress, token, big.NewInt(2500))
	utils.CheckESDTBalance(t, testContext, secondSCAddress, token, big.NewInt(2500))

	expectedSenderBalance := big.NewInt(95000000)
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedSenderBalance)

	expectedAccumulatedFees := big.NewInt(5000000)
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedAccumulatedFees, accumulatedFees)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true, testContext.TxsLogsProcessor)
	testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

	indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
	require.Equal(t, tx.GasLimit, indexerTx.GasUsed)
	require.Equal(t, "5000000", indexerTx.Fee)
}

func TestAsyncESDTCallSecondScRefusesPayment(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)

	// create an address with ESDT token
	sndAddr := []byte("12345678901234567890123456789012")

	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)

	// deploy 2 contracts
	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/second-contract.wasm", ownerAccount, gasPrice, deployGasLimit, argsSecond, big.NewInt(0))

	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	firstSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/first-contract.wasm", ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(0))

	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)
	require.Equal(t, big.NewInt(0), testContext.TxFeeHandler.GetAccumulatedFees())

	gasLimit := uint64(500000)
	tx := utils.CreateESDTTransferTx(0, sndAddr, firstSCAddress, token, big.NewInt(5000), gasPrice, gasLimit)
	tx.Data = []byte(string(tx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractRejected")))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckESDTBalance(t, testContext, firstSCAddress, token, big.NewInt(5000))
	utils.CheckESDTBalance(t, testContext, secondSCAddress, token, big.NewInt(0))

	expectedSenderBalance := big.NewInt(95999990)
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedSenderBalance)

	expectedAccumulatedFees := big.NewInt(4000010)
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedAccumulatedFees, accumulatedFees)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, true, testContext.TxsLogsProcessor)
	testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

	indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
	require.Equal(t, uint64(400001), indexerTx.GasUsed)
	require.Equal(t, "4000010", indexerTx.Fee)
}

func TestAsyncESDTCallsOutOfGas(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)

	// create an address with ESDT token
	sndAddr := []byte("12345678901234567890123456789012")

	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)

	// deploy 2 contracts
	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	argsSecond := [][]byte{[]byte(hex.EncodeToString(token))}
	secondSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/second-contract.wasm", ownerAccount, gasPrice, deployGasLimit, argsSecond, big.NewInt(0))

	args := [][]byte{[]byte(hex.EncodeToString(token)), []byte(hex.EncodeToString(secondSCAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	firstSCAddress := utils.DoDeploySecond(t, testContext, "../esdt/testdata/first-contract.wasm", ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(0))

	testContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasLimit := uint64(2000)
	tx := utils.CreateESDTTransferTx(0, sndAddr, firstSCAddress, token, big.NewInt(5000), gasPrice, gasLimit)
	tx.Data = []byte(string(tx.Data) + "@" + hex.EncodeToString([]byte("transferToSecondContractRejected")))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckESDTBalance(t, testContext, firstSCAddress, token, big.NewInt(0))
	utils.CheckESDTBalance(t, testContext, secondSCAddress, token, big.NewInt(0))

	expectedSenderBalance := big.NewInt(99980000)
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedSenderBalance)

	expectedAccumulatedFees := big.NewInt(20000)
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, expectedAccumulatedFees, accumulatedFees)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	testIndexer := vm.CreateTestIndexer(t, testContext.ShardCoordinator, testContext.EconomicsData, false, testContext.TxsLogsProcessor)
	testIndexer.SaveTransaction(tx, block.TxBlock, intermediateTxs)

	indexerTx := testIndexer.GetIndexerPreparedTransaction(t)
	require.Equal(t, tx.GasLimit, indexerTx.GasUsed)
	require.Equal(t, "20000", indexerTx.Fee)
}

func TestAsyncMultiTransferOnCallback(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	ownerAddr := []byte("12345678901234567890123456789010")
	sftTokenID := []byte("SFT-123456")
	sftNonce := uint64(1)
	sftBalance := big.NewInt(1000)
	halfBalance := big.NewInt(500)

	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, ownerAddr, big.NewInt(1000000000), sftTokenID, sftNonce, sftBalance)
	utils.CheckESDTNFTBalance(t, testContext, ownerAddr, sftTokenID, sftNonce, sftBalance)

	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(1000000)
	txGasLimit := uint64(1000000)

	// deploy forwarder
	forwarderAddr := utils.DoDeploySecond(t,
		testContext,
		"../esdt/testdata/forwarder-raw-managed-api.wasm",
		ownerAccount,
		gasPrice,
		deployGasLimit,
		nil,
		big.NewInt(0),
	)

	// deploy vault
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	vaultAddr := utils.DoDeploySecond(t,
		testContext,
		"../esdt/testdata/vault-managed-api.wasm",
		ownerAccount,
		gasPrice,
		deployGasLimit,
		nil,
		big.NewInt(0),
	)

	// send the tokens to vault
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	tx := utils.CreateESDTNFTTransferTx(
		ownerAccount.GetNonce(),
		ownerAddr,
		vaultAddr,
		sftTokenID,
		sftNonce,
		sftBalance,
		gasPrice,
		txGasLimit,
		"just_accept_funds",
	)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckESDTNFTBalance(t, testContext, vaultAddr, sftTokenID, sftNonce, sftBalance)

	lenSCRs := len(testContext.GetIntermediateTransactions(t))
	// receive tokens from vault to forwarder on callback
	// receive 500 + 500 of the SFT through multi-transfer
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	tx = utils.CreateSmartContractCall(
		ownerAccount.GetNonce(),
		ownerAddr,
		forwarderAddr,
		gasPrice,
		txGasLimit,
		"forward_async_retrieve_multi_transfer_funds",
		vaultAddr,
		sftTokenID,
		big.NewInt(int64(sftNonce)).Bytes(),
		halfBalance.Bytes(),
		sftTokenID,
		big.NewInt(int64(sftNonce)).Bytes(),
		halfBalance.Bytes(),
	)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
	require.Equal(t, 1, len(testContext.GetIntermediateTransactions(t))-lenSCRs)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckESDTNFTBalance(t, testContext, forwarderAddr, sftTokenID, sftNonce, sftBalance)
}

func TestAsyncMultiTransferOnCallAndOnCallback(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	ownerAddr := []byte("12345678901234567890123456789010")
	sftTokenID := []byte("SFT-123456")
	sftNonce := uint64(1)
	sftBalance := big.NewInt(1000)
	halfBalance := big.NewInt(500)

	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, ownerAddr, big.NewInt(1000000000), sftTokenID, sftNonce, sftBalance)
	utils.CheckESDTNFTBalance(t, testContext, ownerAddr, sftTokenID, sftNonce, sftBalance)

	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(1000000)
	txGasLimit := uint64(1000000)

	// deploy forwarder
	forwarderAddr := utils.DoDeploySecond(t,
		testContext,
		"../esdt/testdata/forwarder-raw-managed-api.wasm",
		ownerAccount,
		gasPrice,
		deployGasLimit,
		nil,
		big.NewInt(0),
	)

	// deploy vault
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	vaultAddr := utils.DoDeploySecond(t,
		testContext,
		"../esdt/testdata/vault-managed-api.wasm",
		ownerAccount,
		gasPrice,
		deployGasLimit,
		nil,
		big.NewInt(0),
	)

	// set vault roles
	utils.SetESDTRoles(t, testContext.Accounts, vaultAddr, sftTokenID, [][]byte{
		[]byte(core.ESDTRoleNFTAddQuantity),
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTBurn),
	})
	// set lastNonce for vault
	utils.SetLastNFTNonce(t, testContext.Accounts, vaultAddr, sftTokenID, 1)

	// send the tokens to forwarder
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	tx := utils.CreateESDTNFTTransferTx(
		ownerAccount.GetNonce(),
		ownerAddr,
		forwarderAddr,
		sftTokenID,
		sftNonce,
		sftBalance,
		gasPrice,
		txGasLimit,
		"deposit",
	)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckESDTNFTBalance(t, testContext, forwarderAddr, sftTokenID, sftNonce, sftBalance)

	// send tokens to vault, vault burns and creates new ones, sending them on forwarder's callback
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	tx = utils.CreateSmartContractCall(
		ownerAccount.GetNonce(),
		ownerAddr,
		forwarderAddr,
		gasPrice,
		txGasLimit,
		"forwarder_async_send_and_retrieve_multi_transfer_funds",
		vaultAddr,
		sftTokenID,
		big.NewInt(int64(sftNonce)).Bytes(),
		halfBalance.Bytes(),
		sftTokenID,
		big.NewInt(int64(sftNonce)).Bytes(),
		halfBalance.Bytes(),
	)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	utils.CheckESDTNFTBalance(t, testContext, forwarderAddr, sftTokenID, 2, halfBalance)
	utils.CheckESDTNFTBalance(t, testContext, forwarderAddr, sftTokenID, 3, halfBalance)
}

func TestSendNFTToContractWith0Function(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	ownerAddr := []byte("12345678901234567890123456789010")
	sftTokenID := []byte("SFT-123456")
	sftNonce := uint64(1)
	sftBalance := big.NewInt(1000)

	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, ownerAddr, big.NewInt(1000000000), sftTokenID, sftNonce, sftBalance)
	utils.CheckESDTNFTBalance(t, testContext, ownerAddr, sftTokenID, sftNonce, sftBalance)

	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(1000000)
	txGasLimit := uint64(1000000)

	vaultAddr := utils.DoDeploySecond(t,
		testContext,
		"../esdt/testdata/vault-managed-api.wasm",
		ownerAccount,
		gasPrice,
		deployGasLimit,
		nil,
		big.NewInt(0),
	)

	// send the tokens to vault
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	tx := utils.CreateESDTNFTTransferTx(
		ownerAccount.GetNonce(),
		ownerAddr,
		vaultAddr,
		sftTokenID,
		sftNonce,
		sftBalance,
		gasPrice,
		txGasLimit,
		"",
	)
	tx.Data = append(tx.Data, []byte("@")...)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)
}
