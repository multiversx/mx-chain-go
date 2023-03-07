//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeployDNSContract_TestRegisterAndResolveAndSendTxWithSndAndRcvUserName(t *testing.T) {
	testContext, err := vm.CreatePreparedTxProcessorWithVMs(config.EnableEpochs{})
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeployDNS(t, testContext, "../../multiShard/smartContract/dns/dns.wasm")
	fmt.Println(scAddress)
	utils.CleanAccumulatedIntermediateTransactions(t, testContext)

	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(10000000)
	gasPrice := uint64(10)
	gasLimit := uint64(200000)

	rcvAddr := []byte("12345678901234567890123456789113")

	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, 0, senderBalance)
	_, _ = vm.CreateAccount(testContext.Accounts, rcvAddr, 0, senderBalance)

	userName := utils.GenerateUserNameForDefaultDNSContract()
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
	rcvUserName := utils.GenerateUserNameForDefaultDNSContract()
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

func getNonce(testContext *vm.VMTestContext, address []byte) uint64 {
	accnt, _ := testContext.Accounts.LoadAccount(address)
	return accnt.GetNonce()
}

func getBalance(testContext *vm.VMTestContext, address []byte) *big.Int {
	accnt, _ := testContext.Accounts.LoadAccount(address)
	userAccnt, _ := accnt.(state.UserAccountHandler)

	return userAccnt.GetBalance()
}

// relayer address is in shard 2, creates a transaction on the behalf of the user from shard 2, that will call the DNS contract
// from shard 1 that will try to set the username but fails.
func TestDeployDNSContract_TestGasWhenSaveUsernameFailsCrossShard(t *testing.T) {
	testContextForDNSContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextForDNSContract.Close()

	// TODO remove this
	// logger.SetLogLevel("process:TRACE,vm:TRACE")

	testContextForRelayerAndUser, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{})
	require.Nil(t, err)
	defer testContextForRelayerAndUser.Close()

	scAddress, _ := utils.DoDeployDNS(t, testContextForDNSContract, "../../multiShard/smartContract/dns/dns.wasm")
	fmt.Println(scAddress)
	utils.CleanAccumulatedIntermediateTransactions(t, testContextForDNSContract)
	require.Equal(t, uint32(1), testContextForDNSContract.ShardCoordinator.ComputeId(scAddress))

	relayerAddress := []byte("relayer-901234567890123456789112")
	require.Equal(t, uint32(2), testContextForRelayerAndUser.ShardCoordinator.ComputeId(relayerAddress))
	userAddress := []byte("user-678901234567890123456789112")
	require.Equal(t, uint32(2), testContextForRelayerAndUser.ShardCoordinator.ComputeId(userAddress))

	initialBalance := big.NewInt(10000000)
	_, _ = vm.CreateAccount(testContextForRelayerAndUser.Accounts, relayerAddress, 0, initialBalance)

	firstUsername := utils.GenerateUserNameForDNSContract(scAddress)
	args := argsProcessRegister{
		relayerAddress:               relayerAddress,
		userAddress:                  userAddress,
		scAddress:                    scAddress,
		testContextForRelayerAndUser: testContextForRelayerAndUser,
		testContextForDNSContract:    testContextForDNSContract,
		username:                     firstUsername,
		gasPrice:                     10,
	}
	scrs, retCode, err := processRegisterThroughRelayedTxs(t, args)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, retCode)
	assert.Equal(t, 4, len(scrs))

	// check username
	acc, _ := testContextForRelayerAndUser.Accounts.GetExistingAccount(userAddress)
	account, _ := acc.(state.UserAccountHandler)
	require.Equal(t, firstUsername, account.GetUserName())
	checkBalances(t, args, initialBalance)

	secondUsername := utils.GenerateUserNameForDNSContract(scAddress)
	args.username = secondUsername

	scrs, retCode, err = processRegisterThroughRelayedTxs(t, args)
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, retCode)

	// check username hasn't changed
	acc, _ = testContextForRelayerAndUser.Accounts.GetExistingAccount(userAddress)
	account, _ = acc.(state.UserAccountHandler)
	require.Equal(t, firstUsername, account.GetUserName())
	checkBalances(t, args, initialBalance)

	// TODO refactor
	for _, scr := range scrs {
		log.Info("SCR: " + spew.Sdump(scr))
	}
}

type argsProcessRegister struct {
	relayerAddress               []byte
	userAddress                  []byte
	scAddress                    []byte
	testContextForRelayerAndUser *vm.VMTestContext
	testContextForDNSContract    *vm.VMTestContext
	username                     []byte
	gasPrice                     uint64
}

func processRegisterThroughRelayedTxs(tb testing.TB, args argsProcessRegister) ([]*smartContractResult.SmartContractResult, vmcommon.ReturnCode, error) {
	scrs := make([]*smartContractResult.SmartContractResult, 0)

	// generate the user transaction
	userTxData := []byte("register@" + hex.EncodeToString(args.username))
	userTxGasLimit := uint64(200000)
	userTx := vm.CreateTransaction(
		getNonce(args.testContextForRelayerAndUser, args.userAddress),
		big.NewInt(0),
		args.userAddress,
		args.scAddress,
		args.gasPrice,
		userTxGasLimit,
		userTxData,
	)

	// generate the relayed transaction
	relayedTxData := utils.PrepareRelayerTxData(userTx) // v1 will suffice
	relayedTxGasLimit := userTxGasLimit + 1 + uint64(len(relayedTxData))
	relayedTx := vm.CreateTransaction(
		getNonce(args.testContextForRelayerAndUser, args.relayerAddress),
		big.NewInt(0),
		args.relayerAddress,
		args.userAddress,
		args.gasPrice,
		relayedTxGasLimit,
		relayedTxData,
	)
	// start executing relayed transaction
	retCode, err := args.testContextForRelayerAndUser.TxProcessor.ProcessTransaction(relayedTx)
	if retCode != vmcommon.Ok || err != nil {
		return scrs, retCode, err
	}

	log.Warn("relayer", "balance", getBalance(args.testContextForRelayerAndUser, args.relayerAddress).String())
	log.Warn("relayer", "tx", args.gasPrice*relayedTxGasLimit)

	// record the SCR and clean all intermediate results
	intermediateTxs := args.testContextForRelayerAndUser.GetIntermediateTransactions(tb)
	args.testContextForRelayerAndUser.CleanIntermediateTransactions(tb)
	require.Equal(tb, 1, len(intermediateTxs))
	scrRegister := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	scrs = append(scrs, scrRegister)

	log.Warn("scrRegister", "tx", args.gasPrice*scrRegister.GasLimit)

	// execute the scr on the shard that contains the dns contract
	retCode, err = args.testContextForDNSContract.ScProcessor.ProcessSmartContractResult(scrRegister)
	if retCode != vmcommon.Ok || err != nil {
		return scrs, retCode, err
	}

	// record the SCR and clean all intermediate results
	intermediateTxs = args.testContextForDNSContract.GetIntermediateTransactions(tb)
	args.testContextForDNSContract.CleanIntermediateTransactions(tb)
	require.Equal(tb, 1, len(intermediateTxs))
	scrSCProcess := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	scrs = append(scrs, scrSCProcess)

	log.Warn("scrSCProcess", "tx", args.gasPrice*scrSCProcess.GasLimit)

	// execute the scr on the initial shard that contains the user address (builtin function call)
	retCode, err = args.testContextForRelayerAndUser.ScProcessor.ProcessSmartContractResult(scrSCProcess)
	if retCode != vmcommon.Ok || err != nil {
		return scrs, retCode, err
	}

	// record the SCR and clean all intermediate results
	intermediateTxs = args.testContextForRelayerAndUser.GetIntermediateTransactions(tb)
	args.testContextForRelayerAndUser.CleanIntermediateTransactions(tb)
	require.Equal(tb, 1, len(intermediateTxs))
	scrFinishedBuiltinCall := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	scrs = append(scrs, scrFinishedBuiltinCall)

	// execute the finished scr on the shard that contains the dns contract
	retCode, err = args.testContextForDNSContract.ScProcessor.ProcessSmartContractResult(scrFinishedBuiltinCall)
	if retCode != vmcommon.Ok || err != nil {
		return scrs, retCode, err
	}

	// record the SCR and clean all intermediate results
	intermediateTxs = args.testContextForDNSContract.GetIntermediateTransactions(tb)
	args.testContextForDNSContract.CleanIntermediateTransactions(tb)
	require.Equal(tb, 1, len(intermediateTxs))
	scrSCFinished := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	scrs = append(scrs, scrSCFinished)

	// execute the scr on the initial shard that contains the user address (refund)
	retCode, err = args.testContextForRelayerAndUser.ScProcessor.ProcessSmartContractResult(scrSCFinished)
	if retCode != vmcommon.Ok || err != nil {
		return scrs, retCode, err
	}

	// record the SCR and clean all intermediate results
	intermediateTxs = args.testContextForRelayerAndUser.GetIntermediateTransactions(tb)
	args.testContextForRelayerAndUser.CleanIntermediateTransactions(tb)
	require.Equal(tb, 0, len(intermediateTxs))

	log.Warn("relayer", "balance", getBalance(args.testContextForRelayerAndUser, args.relayerAddress).String())

	// commit & cleanup
	_, err = args.testContextForRelayerAndUser.Accounts.Commit()
	require.Nil(tb, err)
	args.testContextForRelayerAndUser.CleanIntermediateTransactions(tb)

	_, err = args.testContextForDNSContract.Accounts.Commit()
	require.Nil(tb, err)
	args.testContextForDNSContract.CleanIntermediateTransactions(tb)

	return scrs, vmcommon.Ok, nil
}

func checkBalances(tb testing.TB, args argsProcessRegister, initialBalance *big.Int) {
	entireBalance := big.NewInt(0)
	entireBalance.Add(entireBalance, getBalance(args.testContextForRelayerAndUser, args.relayerAddress))
	entireBalance.Add(entireBalance, getBalance(args.testContextForRelayerAndUser, args.userAddress))
	entireBalance.Add(entireBalance, getBalance(args.testContextForDNSContract, args.scAddress))
	entireBalance.Add(entireBalance, args.testContextForRelayerAndUser.TxFeeHandler.GetAccumulatedFees())
	entireBalance.Add(entireBalance, args.testContextForDNSContract.TxFeeHandler.GetAccumulatedFees())

	assert.Equal(tb, initialBalance, entireBalance)
}
