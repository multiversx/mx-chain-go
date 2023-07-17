//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestDeployDNSContract_TestRegisterAndResolveAndSendTxWithSndAndRcvUserName(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: 10,
	})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeployDNS(t, testContext, "../../multiShard/smartContract/dns/dns.wasm")
	fmt.Println(scAddress)
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(10000000)
	gasLimit := uint64(200000)

	rcvAddr := []byte("12345678901234567890123456789113")

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, rcvAddr, 0, senderBalance)

	userName := utils.GenerateUserNameForMyDNSContract()
	txData := []byte("register@" + hex.EncodeToString(userName))
	// create username for sender
	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, txData)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(9299330))
	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(700670), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(70023), developerFees)
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	// create username for receiver
	rcvUserName := utils.GenerateUserNameForMyDNSContract()
	txData = []byte("register@" + hex.EncodeToString(rcvUserName))
	tx = vm.CreateTransaction(0, big.NewInt(0), rcvAddr, scAddress, gasPrice, gasLimit, txData)
	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	vm.TestAccount(t, testContext.Accounts, rcvAddr, 1, big.NewInt(9299330))
	// check accumulated fees
	accumulatedFees = testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(1401340), accumulatedFees)

	developerFees = testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(140046), developerFees)

	ret := vm.GetVmOutput(nil, testContext.Accounts, scAddress, "resolve", userName)
	dnsUserNameAddr := ret.ReturnData[0]
	require.Equal(t, sndAddr, dnsUserNameAddr)

	ret = vm.GetVmOutput(nil, testContext.Accounts, scAddress, "resolve", rcvUserName)
	dnsUserNameAddr = ret.ReturnData[0]
	require.Equal(t, rcvAddr, dnsUserNameAddr)

	acc, _ := testContext.Accounts.GetExistingAccount(sndAddr)
	account, _ := acc.(state.UserAccountHandler)
	un := account.GetUserName()
	require.Equal(t, userName, un)

	acc, _ = testContext.Accounts.GetExistingAccount(rcvAddr)
	account, _ = acc.(state.UserAccountHandler)
	un = account.GetUserName()
	require.Equal(t, rcvUserName, un)

	gasLimit = 10
	tx = vm.CreateTransaction(1, big.NewInt(0), sndAddr, rcvAddr, gasPrice, gasLimit, nil)
	tx.SndUserName = userName
	tx.RcvUserName = rcvUserName

	retCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)
}
