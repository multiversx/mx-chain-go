package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	log = logger.GetOrCreate("integrationTests/vm/txFee/utils")
)

// DoDeploy -
func DoDeploy(
	t *testing.T,
	testContext *vm.VMTestContext,
	pathToContract string,
) (scAddr []byte, owner []byte) {
	return doDeployInternal(t, testContext, pathToContract, 88100, 11900, 399)
}

// DoDeployOldCounter -
func DoDeployOldCounter(
	t *testing.T,
	testContext *vm.VMTestContext,
	pathToContract string,
) (scAddr []byte, owner []byte) {
	return doDeployInternal(t, testContext, pathToContract, 89030, 10970, 368)
}

func doDeployInternal(
	t *testing.T,
	testContext *vm.VMTestContext,
	pathToContract string,
	expectedBalance, accFees, devFees int64,
) (scAddr []byte, owner []byte) {
	owner = []byte("12345678901234567890123456789011")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(2000)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)

	scCode := wasm.GetSCCode(pathToContract)
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(wasm.CreateDeployTxData(scCode)))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	vm.TestAccount(t, testContext.Accounts, owner, senderNonce+1, big.NewInt(expectedBalance))

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(accFees), accumulatedFees)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.WasmVirtualMachine)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(devFees), developerFees)

	return scAddr, owner
}

func generateRandomArray(length int) []byte {
	result := make([]byte, length)
	_, _ = rand.Read(result)

	return result
}

func generateAddressForContextShardID(testContext *vm.VMTestContext) []byte {
	shardID := testContext.ShardCoordinator.SelfId()
	for {
		address := generateRandomArray(32)
		if testContext.ShardCoordinator.ComputeId(address) == shardID {
			return address
		}
	}
}

// DoDeployWithCustomParams -
func DoDeployWithCustomParams(
	tb testing.TB,
	testContext *vm.VMTestContext,
	pathToContract string,
	senderBalance *big.Int,
	gasLimit uint64,
	contractHexParams []string,
) (scAddr []byte, owner []byte) {
	owner = generateAddressForContextShardID(testContext)
	account, err := testContext.Accounts.LoadAccount(owner)
	require.Nil(tb, err)
	senderNonce := account.GetNonce()
	gasPrice := uint64(10)

	_, err = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)
	require.Nil(tb, err)

	scCode := wasm.GetSCCode(pathToContract)
	txData := wasm.CreateDeployTxData(scCode)
	if len(contractHexParams) > 0 {
		txData = strings.Join(append([]string{txData}, contractHexParams...), "@")
	}
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(txData))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(tb, vmcommon.Ok, retCode)
	require.Nil(tb, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.WasmVirtualMachine)

	return scAddr, owner
}

// DoDeployNoChecks -
func DoDeployNoChecks(t *testing.T, testContext *vm.VMTestContext, pathToContract string) (scAddr []byte, owner []byte) {
	owner = []byte("12345678901234567890123456789011")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(2000)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)

	scCode := wasm.GetSCCode(pathToContract)
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(wasm.CreateDeployTxData(scCode)))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.WasmVirtualMachine)

	return scAddr, owner
}

// DoColdDeploy will deploy the SC code but won't call the constructor
func DoColdDeploy(
	tb testing.TB,
	testContext *vm.VMTestContext,
	pathToContract string,
	senderBalance *big.Int,
	codeMetadata string,
) (scAddr []byte, owner []byte) {
	owner = []byte("12345678901234567890123456789011")
	senderNonce := uint64(0)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, senderNonce, senderBalance)
	scCode := wasm.GetSCCode(pathToContract)
	scCodeBytes, err := hex.DecodeString(scCode)
	require.Nil(tb, err)

	codeMetadataBytes, err := hex.DecodeString(codeMetadata)
	require.Nil(tb, err)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, senderNonce, factory.WasmVirtualMachine)
	account, err := testContext.Accounts.LoadAccount(scAddr)
	require.Nil(tb, err)

	userAccount := account.(state.UserAccountHandler)
	userAccount.SetOwnerAddress(owner)
	userAccount.SetCodeMetadata(codeMetadataBytes)
	userAccount.SetCode(scCodeBytes)

	err = testContext.Accounts.SaveAccount(account)
	require.Nil(tb, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	return
}

// DoDeploySecond -
func DoDeploySecond(
	t *testing.T,
	testContext *vm.VMTestContext,
	pathToContract string,
	senderAccount vmcommon.AccountHandler,
	gasPrice uint64,
	gasLimit uint64,
	args [][]byte,
	value *big.Int,
) (scAddr []byte) {
	return DoDeployWithMetadata(t, testContext, pathToContract, senderAccount, gasPrice, gasLimit, []byte(wasm.DummyCodeMetadataHex), args, value)
}

// DoDeployWithMetadata -
func DoDeployWithMetadata(
	t *testing.T,
	testContext *vm.VMTestContext,
	pathToContract string,
	senderAccount vmcommon.AccountHandler,
	gasPrice uint64,
	gasLimit uint64,
	metadata []byte,
	args [][]byte,
	value *big.Int,
) (scAddr []byte) {
	ownerNonce := senderAccount.GetNonce()
	owner := senderAccount.AddressBytes()
	scCode := []byte(wasm.GetSCCode(pathToContract))

	txData := bytes.Join([][]byte{scCode, []byte(wasm.VMTypeHex), metadata}, []byte("@"))
	if args != nil {
		txData = []byte(string(txData) + "@" + string(bytes.Join(args, []byte("@"))))
	}

	tx := vm.CreateTransaction(ownerNonce, value, owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, txData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	acc, _ := testContext.Accounts.LoadAccount(owner)
	require.Equal(t, ownerNonce+1, acc.GetNonce())

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, ownerNonce, factory.WasmVirtualMachine)

	return scAddr
}

// DoDeployDNS -
func DoDeployDNS(t *testing.T, testContext *vm.VMTestContext, pathToContract string) (scAddr []byte, owner []byte) {
	owner = []byte("12345678901234567890123456789011")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(10000000)
	gasPrice := uint64(10)
	gasLimit := uint64(400000)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)

	initParameter := hex.EncodeToString(big.NewInt(1000).Bytes())
	scCode := []byte(wasm.GetSCCode(pathToContract))
	txData := bytes.Join([][]byte{scCode, []byte(wasm.VMTypeHex), []byte(initParameter), []byte("00")}, []byte("@"))
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, txData)

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}

	testContext.TxFeeHandler.CreateBlockStarted(gasAndFees)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.WasmVirtualMachine)
	fmt.Println(hex.EncodeToString(scAddr))
	return scAddr, owner
}

// CheckOwnerAddr -
func CheckOwnerAddr(t *testing.T, testContext *vm.VMTestContext, scAddr []byte, owner []byte) {
	acc, err := testContext.Accounts.GetExistingAccount(scAddr)
	require.Nil(t, err)

	userAcc, ok := acc.(state.UserAccountHandler)
	require.True(t, ok)

	currentOwner := userAcc.GetOwnerAddress()
	require.Equal(t, owner, currentOwner)
}

// TestAccount -
func TestAccount(
	t *testing.T,
	accnts state.AccountsAdapter,
	senderAddressBytes []byte,
	expectedNonce uint64,
	expectedBalance *big.Int,
) *big.Int {

	senderRecovAccount, err := accnts.GetExistingAccount(senderAddressBytes)
	if err != nil {
		assert.Nil(t, err)
		return big.NewInt(0)
	}

	senderRecovShardAccount := senderRecovAccount.(state.UserAccountHandler)

	require.Equal(t, expectedNonce, senderRecovShardAccount.GetNonce())
	require.Equal(t, expectedBalance, senderRecovShardAccount.GetBalance())
	return senderRecovShardAccount.GetBalance()
}

// CreateSmartContractCall -
func CreateSmartContractCall(
	nonce uint64,
	sndAddr []byte,
	rcvAddr []byte,
	gasPrice uint64,
	gasLimit uint64,
	endpointName string,
	arguments ...[]byte) *transaction.Transaction {

	txData := txDataBuilder.NewBuilder()
	txData.Func(endpointName)

	for _, arg := range arguments {
		txData.Bytes(arg)
	}

	return &transaction.Transaction{
		Nonce:    nonce,
		SndAddr:  sndAddr,
		RcvAddr:  rcvAddr,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		Data:     txData.ToBytes(),
		Value:    big.NewInt(0),
	}
}

// ProcessSCRResult -
func ProcessSCRResult(
	tb testing.TB,
	testContext *vm.VMTestContext,
	tx data.TransactionHandler,
	expectedCode vmcommon.ReturnCode,
	expectedErr error,
) {
	scProcessor := testContext.ScProcessor
	require.NotNil(nil, scProcessor)

	scr, ok := tx.(*smartContractResult.SmartContractResult)
	require.True(tb, ok)

	retCode, err := scProcessor.ProcessSmartContractResult(scr)
	require.Equal(tb, expectedCode, retCode)
	require.Equal(tb, expectedErr, err)
}

// CleanAccumulatedIntermediateTransactions -
func CleanAccumulatedIntermediateTransactions(tb testing.TB, testContext *vm.VMTestContext) {
	scForwarder := testContext.ScForwarder
	mockIntermediate, ok := scForwarder.(*mock.IntermediateTransactionHandlerMock)
	require.True(tb, ok)

	mockIntermediate.Clean()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

// randStringBytes -
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		idxBig, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		b[i] = letterBytes[idxBig.Int64()]
	}
	return string(b)
}

// GenerateUserNameForDefaultDNSContract -
func GenerateUserNameForDefaultDNSContract() []byte {
	return GenerateUserNameForDNSContract([]byte{49})
}

// GenerateUserNameForDNSContract -
func GenerateUserNameForDNSContract(contractAddress []byte) []byte {
	testHasher := keccak.NewKeccak()
	contractLastByte := contractAddress[len(contractAddress)-1]

	for {
		userName := randStringBytes(10)
		userName += ".elrond"
		userNameHash := testHasher.Compute(userName)

		if userNameHash[len(userNameHash)-1] == contractLastByte {
			return []byte(userName)
		}
	}
}

// OverwriteAccountStorageWithHexFileContent applies pairs of <key,value> from provided file to the state of the provided address
// Before applying the data it does a cleanup on the old state
// the data from the file must be in the following format:
//
// hex(key1),hex(value1)
// hex(key2),hex(value2)
// ...
//
// Example:
// 61750100,0000
// 61750101,0001
func OverwriteAccountStorageWithHexFileContent(tb testing.TB, testContext *vm.VMTestContext, address []byte, pathToData string) {
	allData, err := os.ReadFile(filepath.Clean(pathToData))
	require.Nil(tb, err)

	account, err := testContext.Accounts.GetExistingAccount(address)
	require.Nil(tb, err)

	userAccount := account.(state.UserAccountHandler)
	userAccount.SetRootHash(nil)
	err = testContext.Accounts.SaveAccount(account)
	require.Nil(tb, err)

	lines := strings.Split(string(allData), "\n")
	numProcessed := 0
	for _, line := range lines {
		split := strings.Split(line, ",")
		if len(split) != 2 {
			continue
		}

		key, errDecode := hex.DecodeString(strings.TrimSpace(split[0]))
		require.Nil(tb, errDecode)

		value, errDecode := hex.DecodeString(strings.TrimSpace(split[1]))
		require.Nil(tb, errDecode)

		err = userAccount.SaveKeyValue(key, value)
		require.Nil(tb, err)
		numProcessed++
	}
	log.Info("ApplyData", "total file lines", len(lines), "processed lines", numProcessed)

	err = testContext.Accounts.SaveAccount(account)
	require.Nil(tb, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)
}
