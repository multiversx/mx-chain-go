package txsFee

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/require"
)

func TestMultiESDTTransferShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	egldBalance := big.NewInt(100000000)
	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)
	secondToken := []byte("second")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, big.NewInt(0), secondToken, 0, esdtBalance)

	gasLimit := uint64(4000)
	tx := utils.CreateMultiTransferTX(0, sndAddr, rcvAddr, gasPrice, gasLimit, &utils.TransferESDTData{
		Token: token,
		Value: big.NewInt(100),
	}, &utils.TransferESDTData{
		Token: secondToken,
		Value: big.NewInt(200),
	})
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
	require.Nil(t, testContext.GetCompositeTestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalanceSnd := big.NewInt(99999900)
	utils.CheckESDTBalance(t, testContext, sndAddr, token, expectedBalanceSnd)

	expectedReceiverBalance := big.NewInt(100)
	utils.CheckESDTBalance(t, testContext, rcvAddr, token, expectedReceiverBalance)

	expectedBalanceSndSecondToken := big.NewInt(99999800)
	utils.CheckESDTBalance(t, testContext, sndAddr, secondToken, expectedBalanceSndSecondToken)

	expectedReceiverBalanceSecondToken := big.NewInt(200)
	utils.CheckESDTBalance(t, testContext, rcvAddr, secondToken, expectedReceiverBalanceSecondToken)

	expectedEGLDBalance := big.NewInt(99960000)
	utils.TestAccount(t, testContext.Accounts, sndAddr, 1, expectedEGLDBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(40000), accumulatedFees)

	allLogs := testContext.TxsLogsProcessor.GetAllCurrentLogs()
	require.NotNil(t, allLogs)
}

func TestMultiESDTTransferFailsBecauseOfMaxLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testContext, err := vm.CreatePreparedTxProcessorWithVMsAndCustomGasSchedule(config.EnableEpochs{},
		func(gasMap wasmConfig.GasScheduleMap) {
			gasMap[common.MaxPerTransaction]["MaxNumberOfTransfersPerTx"] = 1
		})
	require.Nil(t, err)
	defer testContext.Close()

	sndAddr := []byte("12345678901234567890123456789012")
	rcvAddr := []byte("12345678901234567890123456789022")

	egldBalance := big.NewInt(100000000)
	esdtBalance := big.NewInt(100000000)
	token := []byte("miiutoken")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, egldBalance, token, 0, esdtBalance)
	secondToken := []byte("second")
	utils.CreateAccountWithESDTBalance(t, testContext.Accounts, sndAddr, big.NewInt(0), secondToken, 0, esdtBalance)

	gasLimit := uint64(4000)
	tx := utils.CreateMultiTransferTX(0, sndAddr, rcvAddr, gasPrice, gasLimit, &utils.TransferESDTData{
		Token: token,
		Value: big.NewInt(100),
	}, &utils.TransferESDTData{
		Token: secondToken,
		Value: big.NewInt(200),
	})
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.NotNil(t, err)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Contains(t, testContext.GetCompositeTestError().Error(), process.ErrMaxCallsReached.Error())
}
