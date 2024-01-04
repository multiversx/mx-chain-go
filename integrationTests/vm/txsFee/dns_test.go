//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"unicode/utf8"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const returnOkData = "@6f6b"

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

	userName := utils.GenerateUserNameForDefaultDNSContract()
	txData := []byte("register@" + hex.EncodeToString(userName))
	// create username for sender
	tx := vm.CreateTransaction(0, big.NewInt(0), sndAddr, scAddress, gasPrice, gasLimit, txData)
	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	vm.TestAccount(t, testContext.Accounts, sndAddr, 1, big.NewInt(9721810))
	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(278190), accumulatedFees)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(27775), developerFees)

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

	vm.TestAccount(t, testContext.Accounts, rcvAddr, 1, big.NewInt(9721810))
	// check accumulated fees
	accumulatedFees = testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(556380), accumulatedFees)

	developerFees = testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(55550), developerFees)

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

// relayer address is in shard 2, creates a transaction on the behalf of the user from shard 2, that will call the DNS contract
// from shard 1.
func TestDeployDNSContract_TestGasWhenSaveUsernameFailsCrossShardBackwardsCompatibility(t *testing.T) {
	enableEpochs := config.EnableEpochs{
		ChangeUsernameEnableEpoch:                       1000, // flag disabled, backwards compatibility
		SCProcessorV2EnableEpoch:                        1000,
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: 1000,
	}

	vmConfig := vm.CreateVMConfigWithVersion("v1.4")
	testContextForDNSContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShardRoundVMConfig(
		1,
		enableEpochs,
		integrationTests.GetDefaultRoundsConfig(),
		vmConfig,
	)
	require.Nil(t, err)
	defer testContextForDNSContract.Close()

	testContextForRelayerAndUser, err := vm.CreatePreparedTxProcessorWithVMsMultiShardRoundVMConfig(
		2,
		enableEpochs,
		integrationTests.GetDefaultRoundsConfig(),
		vmConfig,
	)
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

	initialBalance := big.NewInt(10000000000)
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
	assert.Equal(t, 3, len(scrs))

	expectedTotalBalance := big.NewInt(0).Set(initialBalance)
	expectedTotalBalance.Sub(expectedTotalBalance, big.NewInt(10)) // due to a bug, some fees were burnt

	// check username
	acc, _ := testContextForRelayerAndUser.Accounts.GetExistingAccount(userAddress)
	account, _ := acc.(state.UserAccountHandler)
	require.Equal(t, firstUsername, account.GetUserName())
	checkBalances(t, args, expectedTotalBalance)

	secondUsername := utils.GenerateUserNameForDNSContract(scAddress)
	args.username = secondUsername

	_, retCode, err = processRegisterThroughRelayedTxs(t, args)
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, retCode)

	// check username hasn't changed
	acc, _ = testContextForRelayerAndUser.Accounts.GetExistingAccount(userAddress)
	account, _ = acc.(state.UserAccountHandler)
	require.Equal(t, firstUsername, account.GetUserName())
	checkBalances(t, args, expectedTotalBalance)
}

func TestDeployDNSContract_TestGasWhenSaveUsernameAfterDNSv2IsActivated(t *testing.T) {
	testContextForDNSContract, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(1, config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	})
	require.Nil(t, err)
	defer testContextForDNSContract.Close()

	testContextForRelayerAndUser, err := vm.CreatePreparedTxProcessorWithVMsMultiShard(2, config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: integrationTests.UnreachableEpoch,
	})
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

	initialBalance := big.NewInt(10000000000)
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
	assert.Equal(t, 3, len(scrs))

	// check username
	acc, _ := testContextForRelayerAndUser.Accounts.GetExistingAccount(userAddress)
	account, _ := acc.(state.UserAccountHandler)
	require.Equal(t, firstUsername, account.GetUserName())
	checkBalances(t, args, initialBalance)

	secondUsername := utils.GenerateUserNameForDNSContract(scAddress)
	args.username = secondUsername

	_, retCode, err = processRegisterThroughRelayedTxs(t, args)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, retCode)

	// check username has changed
	acc, _ = testContextForRelayerAndUser.Accounts.GetExistingAccount(userAddress)
	account, _ = acc.(state.UserAccountHandler)
	require.Equal(t, secondUsername, account.GetUserName())
	checkBalances(t, args, initialBalance)
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

	log.Info("user tx", "tx", txToString(userTx))

	// generate the relayed transaction
	relayedTxData := integrationTests.PrepareRelayedTxDataV1(userTx) // v1 will suffice
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

	log.Info("executing relayed tx", "tx", txToString(relayedTx))

	// start executing relayed transaction
	retCode, err := args.testContextForRelayerAndUser.TxProcessor.ProcessTransaction(relayedTx)
	if err != nil {
		return scrs, retCode, err
	}

	intermediateTxs := args.testContextForRelayerAndUser.GetIntermediateTransactions(tb)
	args.testContextForRelayerAndUser.CleanIntermediateTransactions(tb)
	if len(intermediateTxs) == 0 {
		return scrs, retCode, err // execution finished
	}
	testContexts := []*vm.VMTestContext{args.testContextForRelayerAndUser, args.testContextForDNSContract}

	globalReturnCode := vmcommon.Ok

	for {
		scr := intermediateTxs[0].(*smartContractResult.SmartContractResult)
		scrs = append(scrs, scr)

		context := chooseVMTestContexts(scr, testContexts)
		require.NotNil(tb, context)

		// execute the smart contract result
		log.Info("executing scr", "on shard", context.ShardCoordinator.SelfId(), "scr", scrToString(scr))

		retCode, err = context.ScProcessor.ProcessSmartContractResult(scr)
		if err != nil {
			return scrs, retCode, err
		}
		if retCode != vmcommon.Ok {
			globalReturnCode = retCode
		}
		if string(scr.Data) == returnOkData {
			return scrs, globalReturnCode, err // execution finished
		}

		intermediateTxs = context.GetIntermediateTransactions(tb)
		context.CleanIntermediateTransactions(tb)
		if len(intermediateTxs) == 0 {
			return scrs, globalReturnCode, err // execution finished
		}
	}
}

func chooseVMTestContexts(scr *smartContractResult.SmartContractResult, contexts []*vm.VMTestContext) *vm.VMTestContext {
	for _, context := range contexts {
		if context.ShardCoordinator.ComputeId(scr.RcvAddr) == context.ShardCoordinator.SelfId() {
			return context
		}
	}

	return nil
}

func scrToString(scr *smartContractResult.SmartContractResult) string {
	data := string(scr.Data)
	if !isASCII(data) {
		data = hex.EncodeToString(scr.Data)
	}

	hash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, scr)

	rcv, _ := integrationTests.TestAddressPubkeyConverter.Encode(scr.RcvAddr)
	snd, _ := integrationTests.TestAddressPubkeyConverter.Encode(scr.SndAddr)
	return fmt.Sprintf("hash: %s, nonce: %d, value: %s, rcvAddr: %s, sender: %s, gasLimit: %d, gasPrice: %d, data: %s",
		hex.EncodeToString(hash),
		scr.Nonce, scr.Value.String(),
		rcv,
		snd,
		scr.GasLimit, scr.GasPrice, data,
	)
}

func txToString(tx *transaction.Transaction) string {
	data := string(tx.Data)
	if !isASCII(data) {
		data = hex.EncodeToString(tx.Data)
	}

	hash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, tx)
	rcv, _ := integrationTests.TestAddressPubkeyConverter.Encode(tx.RcvAddr)
	snd, _ := integrationTests.TestAddressPubkeyConverter.Encode(tx.SndAddr)
	return fmt.Sprintf("hash: %s, nonce: %d, value: %s, rcvAddr: %s, sender: %s, gasLimit: %d, gasPrice: %d, data: %s",
		hex.EncodeToString(hash),
		tx.Nonce, tx.Value.String(),
		rcv,
		snd,
		tx.GasLimit, tx.GasPrice, data,
	)
}

func isASCII(data string) bool {
	for i := 0; i < len(data); i++ {
		if data[i] >= utf8.RuneSelf {
			return false
		}

		if data[i] >= logger.ASCIISpace {
			continue
		}

		if data[i] == logger.ASCIITab || data[i] == logger.ASCIILineFeed || data[i] == logger.ASCIINewLine {
			continue
		}

		return false
	}

	return true
}

func checkBalances(tb testing.TB, args argsProcessRegister, initialBalance *big.Int) {
	entireBalance := big.NewInt(0)
	relayerBalance := getBalance(args.testContextForRelayerAndUser, args.relayerAddress)
	userBalance := getBalance(args.testContextForRelayerAndUser, args.userAddress)
	scBalance := getBalance(args.testContextForDNSContract, args.scAddress)
	accumulatedFees := big.NewInt(0).Set(args.testContextForRelayerAndUser.TxFeeHandler.GetAccumulatedFees())
	accumulatedFees.Add(accumulatedFees, args.testContextForDNSContract.TxFeeHandler.GetAccumulatedFees())

	entireBalance.Add(entireBalance, relayerBalance)
	entireBalance.Add(entireBalance, userBalance)
	entireBalance.Add(entireBalance, scBalance)
	entireBalance.Add(entireBalance, accumulatedFees)

	log.Info("checkBalances",
		"relayerBalance", relayerBalance.String(),
		"userBalance", userBalance.String(),
		"scBalance", scBalance.String(),
		"accumulated fees", accumulatedFees.String(),
		"total", entireBalance.String(),
	)

	assert.Equal(tb, initialBalance, entireBalance)
}
