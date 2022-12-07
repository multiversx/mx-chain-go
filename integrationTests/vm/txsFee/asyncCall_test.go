//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Arwen fix

package txsFee

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestAsyncCallShouldWork(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(100000000)
	senderAddr := []byte("12345678901234567890123456789011")
	ownerAddr := []byte("12345678901234567890123456789010")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, senderAddr, 0, egldBalance)

	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(50000)

	pathToContract := "testdata/first/first.wasm"
	firstScAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(50))

	gasLimit := uint64(5000000)
	args := [][]byte{[]byte(hex.EncodeToString(firstScAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	pathToContract = "testdata/second/output/async.wasm"
	secondSCAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(50))

	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
	testContext.TxFeeHandler.CreateBlockStarted(gasAndFees)

	tx := vm.CreateTransaction(0, big.NewInt(0), senderAddr, secondSCAddress, gasPrice, gasLimit, []byte("doSomething"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.NotNil(t, intermediateTxs)

	res := vm.GetIntValueFromSC(nil, testContext.Accounts, firstScAddress, "numCalled")
	require.Equal(t, big.NewInt(1), res)

	require.Equal(t, big.NewInt(50000000), testContext.TxFeeHandler.GetAccumulatedFees())
	require.Equal(t, big.NewInt(4999988), testContext.TxFeeHandler.GetDeveloperFees())
}

func TestMinterContractWithAsyncCalls(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	egldBalance := big.NewInt(1000000000000)
	ownerAddr := []byte("12345678901234567890123456789011")
	_, _ = vm.CreateAccount(testContext.Accounts, ownerAddr, 0, egldBalance)

	token := []byte("miiutoken")
	roles := [][]byte{[]byte(core.ESDTRoleNFTCreate)}

	gasPrice := uint64(10)
	ownerAccount, _ := testContext.Accounts.LoadAccount(ownerAddr)
	deployGasLimit := uint64(500000)
	pathToContract := "testdata/minter/minter.wasm"
	firstScAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	secondContractAddress := utils.DoDeploySecond(t, testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(0))

	utils.SetESDTRoles(t, testContext.Accounts, firstScAddress, token, roles)
	utils.SetESDTRoles(t, testContext.Accounts, secondContractAddress, token, roles)

	// DO call
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	dataField := []byte(fmt.Sprintf("setTokenID@%s", hex.EncodeToString(token)))
	tx := vm.CreateTransaction(ownerAccount.GetNonce(), big.NewInt(0), ownerAddr, firstScAddress, gasPrice, deployGasLimit, dataField)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	tx = vm.CreateTransaction(ownerAccount.GetNonce(), big.NewInt(0), ownerAddr, secondContractAddress, gasPrice, deployGasLimit, dataField)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	dataField = []byte(fmt.Sprintf("tryMore@%s", hex.EncodeToString(firstScAddress)))
	gasLimit := 600000000
	tx = vm.CreateTransaction(ownerAccount.GetNonce(), big.NewInt(0), ownerAddr, secondContractAddress, gasPrice, uint64(gasLimit), dataField)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, retCode)
	require.Nil(t, err)

	logs := testContext.TxsLogsProcessor.GetAllCurrentLogs()
	event := logs[4].GetLogEvents()[1]
	require.Equal(t, "internalVMErrors", string(event.GetIdentifier()))
	require.Contains(t, string(event.GetData()), process.ErrMaxBuiltInCallsReached.Error())
}
