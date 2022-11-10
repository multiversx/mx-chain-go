package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	protoMarshalizer = &marshal.GogoProtoMarshalizer{}
	log              = logger.GetOrCreate("integrationTests/vm/txFee/utils")
)

// DoDeploy -
func DoDeploy(t *testing.T, testContext *vm.VMTestContext, pathToContract string) (scAddr []byte, owner []byte) {
	owner = []byte("12345678901234567890123456789011")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(2000)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)

	scCode := arwen.GetSCCode(pathToContract)
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(arwen.CreateDeployTxData(scCode)))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	expectedBalance := big.NewInt(89030)
	vm.TestAccount(t, testContext.Accounts, owner, senderNonce+1, expectedBalance)

	// check accumulated fees
	accumulatedFees := testContext.TxFeeHandler.GetAccumulatedFees()
	require.Equal(t, big.NewInt(10970), accumulatedFees)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.ArwenVirtualMachine)

	developerFees := testContext.TxFeeHandler.GetDeveloperFees()
	require.Equal(t, big.NewInt(368), developerFees)

	return scAddr, owner
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
	owner = []byte("12345678901234567890123456789011")
	senderNonce := uint64(0)
	gasPrice := uint64(10)

	_, _ = vm.CreateAccount(testContext.Accounts, owner, 0, senderBalance)

	scCode := arwen.GetSCCode(pathToContract)
	txData := arwen.CreateDeployTxData(scCode)
	if len(contractHexParams) > 0 {
		txData = strings.Join(append([]string{txData}, contractHexParams...), "@")
	}
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(txData))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(tb, vmcommon.Ok, retCode)
	require.Nil(tb, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.ArwenVirtualMachine)

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

	scCode := arwen.GetSCCode(pathToContract)
	tx := vm.CreateTransaction(senderNonce, big.NewInt(0), owner, vm.CreateEmptyAddress(), gasPrice, gasLimit, []byte(arwen.CreateDeployTxData(scCode)))

	retCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.Ok, retCode)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.ArwenVirtualMachine)

	return scAddr, owner
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
	return DoDeployWithMetadata(t, testContext, pathToContract, senderAccount, gasPrice, gasLimit, []byte(arwen.DummyCodeMetadataHex), args, value)
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
	scCode := []byte(arwen.GetSCCode(pathToContract))

	txData := bytes.Join([][]byte{scCode, []byte(arwen.VMTypeHex), metadata}, []byte("@"))
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

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, ownerNonce, factory.ArwenVirtualMachine)

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
	scCode := []byte(arwen.GetSCCode(pathToContract))
	txData := bytes.Join([][]byte{scCode, []byte(arwen.VMTypeHex), []byte(initParameter), []byte("00")}, []byte("@"))
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

	scAddr, _ = testContext.BlockchainHook.NewAddress(owner, 0, factory.ArwenVirtualMachine)
	fmt.Println(hex.EncodeToString(scAddr))
	return scAddr, owner
}

// PrepareRelayerTxData -
func PrepareRelayerTxData(innerTx *transaction.Transaction) []byte {
	userTxBytes, _ := protoMarshalizer.Marshal(innerTx)
	return []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxBytes))
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
	t *testing.T,
	testContext *vm.VMTestContext,
	tx data.TransactionHandler,
	expectedCode vmcommon.ReturnCode,
	expectedErr error,
) {
	scProcessor := testContext.ScProcessor
	require.NotNil(nil, scProcessor)

	scr, ok := tx.(*smartContractResult.SmartContractResult)
	require.True(t, ok)

	retCode, err := scProcessor.ProcessSmartContractResult(scr)
	require.Equal(t, expectedCode, retCode)
	require.Equal(t, expectedErr, err)
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
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GenerateUserNameForMyDNSContract -
func GenerateUserNameForMyDNSContract() []byte {
	testHasher := keccak.NewKeccak()
	contractLastByte := byte(49)

	for {
		userName := randStringBytes(10)
		userName += ".elrond"
		userNameHash := testHasher.Compute(userName)

		if userNameHash[len(userNameHash)-1] == contractLastByte {
			return []byte(userName)
		}
	}
}

// ApplyDataOverwritingExistingData applies pairs of <key,value> from provided file to the state of the provided address
// Before applying the data it does a cleanup on the old state
// the data from the file must be in the following format:
//hex(key1),hex(value1)
//hex(key2),hex(value2)
//...
// Example:
//61750100,0000
//61750101,0001
func ApplyDataOverwritingExistingData(tb testing.TB, testContext *vm.VMTestContext, address []byte, pathToData string) {
	allData, err := ioutil.ReadFile(filepath.Clean(pathToData))
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
