package utils

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
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

var protoMarshalizer = &marshal.GogoProtoMarshalizer{}

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
	ownerNonce := senderAccount.GetNonce()
	owner := senderAccount.AddressBytes()
	scCode := []byte(arwen.GetSCCode(pathToContract))

	txData := bytes.Join([][]byte{scCode, []byte(arwen.VMTypeHex), []byte(arwen.DummyCodeMetadataHex)}, []byte("@"))
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
func CleanAccumulatedIntermediateTransactions(t *testing.T, testContext *vm.VMTestContext) {
	scForwarder := testContext.ScForwarder
	mockIntermediate, ok := scForwarder.(*mock.IntermediateTransactionHandlerMock)
	require.True(t, ok)

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
