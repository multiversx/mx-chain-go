package txsFee

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/txsFee/utils"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
)

func TestAsyncCallShoutWork(t *testing.T) {
	testContext := vm.CreatePreparedTxProcessorWithVMs(t, vm.ArgEnableEpoch{})
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
	firstScAddress := utils.DoDeploySecond(t, &testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, nil, big.NewInt(50))

	gasLimit := uint64(5000000)
	args := [][]byte{[]byte(hex.EncodeToString(firstScAddress))}
	ownerAccount, _ = testContext.Accounts.LoadAccount(ownerAddr)
	pathToContract = "testdata/second/output/async.wasm"
	secondSCAddress := utils.DoDeploySecond(t, &testContext, pathToContract, ownerAccount, gasPrice, deployGasLimit, args, big.NewInt(50))

	tx := vm.CreateTransaction(0, big.NewInt(0), senderAddr, secondSCAddress, gasPrice, gasLimit, []byte("doSomething"))
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
	require.Nil(t, testContext.GetLatestError())

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	// TODO not done yet
}
