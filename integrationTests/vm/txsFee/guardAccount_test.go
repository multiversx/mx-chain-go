//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Arwen fix

package txsFee

import (
	"encoding/hex"
	"math/big"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/guardians"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/guardian"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/integrationtests"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const instantSetGuardian = true
const delayedSetGuardian = false

func prepareTestContextForFreezeAccounts(tb testing.TB) *vm.VMTestContext {
	unreachableEpoch := uint32(999999)
	db := integrationtests.CreateStorer(tb.TempDir())
	gasScheduleDir := "../../../cmd/node/config/gasSchedules"

	cfg := config.GasScheduleByEpochs{
		StartEpoch: 0,
		FileName:   "gasScheduleV6.toml",
	}

	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig: config.GasScheduleConfig{
			GasScheduleByEpochs: []config.GasScheduleByEpochs{cfg},
		},
		ConfigDir:         gasScheduleDir,
		EpochNotifier:     forking.NewGenericEpochNotifier(),
		ArwenChangeLocker: &sync.RWMutex{},
	}
	gasScheduleNotifier, err := forking.NewGasScheduleNotifier(argsGasScheduleNotifier)
	require.Nil(tb, err)

	testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(
		config.EnableEpochs{
			GovernanceEnableEpoch:                   unreachableEpoch,
			WaitingListFixEnableEpoch:               unreachableEpoch,
			SetSenderInEeiOutputTransferEnableEpoch: unreachableEpoch,
			RefactorPeersMiniBlocksEnableEpoch:      unreachableEpoch,
			GuardAccountFeatureEnableEpoch:          0,
		},
		mock.NewMultiShardsCoordinatorMock(2),
		db,
		gasScheduleNotifier,
	)
	require.Nil(tb, err)

	return testContext
}

func mintAddress(tb testing.TB, testContext *vm.VMTestContext, address []byte, value *big.Int) {
	accnt, err := testContext.Accounts.LoadAccount(address)
	require.Nil(tb, err)

	userAccnt := accnt.(vmcommon.UserAccountHandler)
	err = userAccnt.AddToBalance(value)
	require.Nil(tb, err)

	err = testContext.Accounts.SaveAccount(accnt)
	require.Nil(tb, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)
}

func getNonce(testContext *vm.VMTestContext, address []byte) uint64 {
	accnt, _ := testContext.Accounts.LoadAccount(address)

	return accnt.GetNonce()
}

func getGuardiansData(tb testing.TB, testContext *vm.VMTestContext, address []byte) (active *guardians.Guardian, pending *guardians.Guardian, err error) {
	accnt, err := testContext.Accounts.GetExistingAccount(address)
	require.Nil(tb, err)

	userAccnt := accnt.(state.UserAccountHandler)
	guardedAccount, err := guardian.NewGuardedAccount(
		testContext.Marshalizer,
		testContext.EpochNotifier,
		vm.EpochGuardianDelay)
	require.Nil(tb, err)

	return guardedAccount.GetConfiguredGuardians(userAccnt)
}

func setGuardian(testContext *vm.VMTestContext, userAddress []byte, guardianAddress []byte, uuid []byte, isInstantSet bool) (vmcommon.ReturnCode, error) {
	gasPrice := uint64(10)
	gasLimit := uint64(250000 + 1000)

	tx := vm.CreateTransaction(
		getNonce(testContext, userAddress),
		big.NewInt(0),
		userAddress,
		userAddress,
		gasPrice,
		gasLimit,
		[]byte("SetGuardian@"+hex.EncodeToString(guardianAddress)+"@"+hex.EncodeToString(uuid)))

	if isInstantSet {
		tx.GuardianAddr = guardianAddress
	}

	return testContext.TxProcessor.ProcessTransaction(tx)
}

func guardAccount(testContext *vm.VMTestContext, userAddress []byte) (vmcommon.ReturnCode, error) {
	gasPrice := uint64(10)
	gasLimit := uint64(400000 + 1000)

	tx := vm.CreateTransaction(
		getNonce(testContext, userAddress),
		big.NewInt(0),
		userAddress,
		userAddress,
		gasPrice,
		gasLimit,
		[]byte("GuardAccount"),
	)
	return testContext.TxProcessor.ProcessTransaction(tx)
}

func transferFunds(
	testContext *vm.VMTestContext,
	senderAddress []byte,
	transferValue *big.Int,
	receiverAddress []byte,
) error {
	gasPrice := uint64(10)
	gasLimit := uint64(1000)

	tx := vm.CreateTransaction(
		getNonce(testContext, senderAddress),
		transferValue,
		senderAddress,
		receiverAddress,
		gasPrice,
		gasLimit,
		make([]byte, 0))

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	return err
}

func transferFundsWithGuardian(
	testContext *vm.VMTestContext,
	senderAddress []byte,
	transferValue *big.Int,
	receiverAddress []byte,
	guardianAddress []byte,
) error {
	gasPrice := uint64(10)
	gasLimit := uint64(51000)

	tx := vm.CreateTransaction(
		getNonce(testContext, senderAddress),
		transferValue,
		senderAddress,
		receiverAddress,
		gasPrice,
		gasLimit,
		make([]byte, 0))
	tx.Version = core.InitialVersionOfTransaction + 1
	tx.Options = tx.Options | transaction.MaskGuardedTransaction
	tx.GuardianAddr = guardianAddress

	_, err := testContext.TxProcessor.ProcessTransaction(tx)
	return err
}

func getBalance(testContext *vm.VMTestContext, address []byte) *big.Int {
	accnt, _ := testContext.Accounts.LoadAccount(address)
	userAccnt := accnt.(vmcommon.UserAccountHandler)

	return userAccnt.GetBalance()
}

func testNoGuardianIsSet(tb testing.TB, testContext *vm.VMTestContext, address []byte) {
	active, pending, err := getGuardiansData(tb, testContext, address)
	require.Nil(tb, err)
	assert.Nil(tb, active)
	assert.Nil(tb, pending)
}

func testActiveGuardian(
	tb testing.TB,
	testContext *vm.VMTestContext,
	address []byte,
	shouldBeNil bool,
	expectedAddress []byte,
	expectedUUID []byte,
) {
	active, _, err := getGuardiansData(tb, testContext, address)
	require.Nil(tb, err)

	testGuardianData(tb, active, shouldBeNil, expectedAddress, expectedUUID)
}

func testGuardianData(
	tb testing.TB,
	guardian *guardians.Guardian,
	shouldBeNil bool,
	expectedAddress []byte,
	expectedUUID []byte,
) {
	if shouldBeNil {
		require.Nil(tb, guardian)
		return
	}

	require.NotNil(tb, guardian)
	assert.Equal(tb, expectedAddress, guardian.Address)
	assert.Equal(tb, expectedUUID, guardian.ServiceUID)
}

func testPendingGuardian(
	tb testing.TB,
	testContext *vm.VMTestContext,
	address []byte,
	shouldBeNil bool,
	expectedAddress []byte,
	expectedUUID []byte,
) {
	_, pending, err := getGuardiansData(tb, testContext, address)
	require.Nil(tb, err)

	testGuardianData(tb, pending, shouldBeNil, expectedAddress, expectedUUID)
}

func TestGuardAccount_ShouldErrorIfInstantSetIsDoneOnANotProtectedAccount(t *testing.T) {
	testContext := prepareTestContextForFreezeAccounts(t)
	defer testContext.Close()

	uuid := []byte("uuid")
	initialMint := big.NewInt(1000000000000000000)
	userAddress := []byte("user-123456789012345678901234567")
	guardianAddress := []byte("guardian-12345678901234567890123")
	mintAddress(t, testContext, userAddress, initialMint)

	testNoGuardianIsSet(t, testContext, userAddress)

	returnCode, err := setGuardian(testContext, userAddress, guardianAddress, uuid, instantSetGuardian)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)

	testNoGuardianIsSet(t, testContext, userAddress)
}

func TestGuardAccount_ShouldSetGuardianOnANotProtectedAccount(t *testing.T) {
	testContext := prepareTestContextForFreezeAccounts(t)
	defer testContext.Close()

	uuid := []byte("uuid")
	initialMint := big.NewInt(1000000000000000000)
	userAddress := []byte("user-123456789012345678901234567")
	guardianAddress := []byte("guardian-12345678901234567890123")
	mintAddress(t, testContext, userAddress, initialMint)

	testNoGuardianIsSet(t, testContext, userAddress)

	returnCode, err := setGuardian(testContext, userAddress, guardianAddress, uuid, delayedSetGuardian)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)

	testActiveGuardian(t, testContext, userAddress, true, nil, nil)
	testPendingGuardian(t, testContext, userAddress, false, guardianAddress, uuid)

	// can not activate guardian now
	returnCode, err = guardAccount(testContext, userAddress)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)

	hdr := &block.Header{
		Epoch: vm.EpochGuardianDelay,
	}
	testContext.EpochNotifier.CheckEpoch(hdr)

	testActiveGuardian(t, testContext, userAddress, false, guardianAddress, uuid)
	testPendingGuardian(t, testContext, userAddress, true, nil, nil)

	// can activate guardian now
	returnCode, err = guardAccount(testContext, userAddress)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
}

func TestGuardAccount_SendingFundsWhileProtectedAndNotProtected(t *testing.T) {
	testContext := prepareTestContextForFreezeAccounts(t)
	defer testContext.Close()

	uuid := []byte("uuid")
	transferValue := int64(2000000)
	initialMint := big.NewInt(1000000000000000000)
	userAddress := []byte("user-123456789012345678901234567")
	receiverAddress := []byte("recv-123456789012345678901234567")
	guardianAddress := []byte("guardian-12345678901234567890123")
	wrongGuardianAddress := []byte("wrong-guardian-12345678901234523")
	mintAddress(t, testContext, userAddress, initialMint)

	testNoGuardianIsSet(t, testContext, userAddress)

	// userAddress can send funds while not protected
	err := transferFunds(testContext, userAddress, big.NewInt(transferValue), receiverAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while not protected with a guardian address
	err = transferFundsWithGuardian(testContext, userAddress, big.NewInt(transferValue), receiverAddress, guardianAddress)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "guarded transaction not expected")
	require.Equal(t, big.NewInt(transferValue), getBalance(testContext, receiverAddress))

	// userAddress can send funds while it just added a guardian
	returnCode, err := setGuardian(testContext, userAddress, guardianAddress, uuid, delayedSetGuardian)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)

	testActiveGuardian(t, testContext, userAddress, true, nil, nil)
	testPendingGuardian(t, testContext, userAddress, false, guardianAddress, uuid)

	err = transferFunds(testContext, userAddress, big.NewInt(transferValue), receiverAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue*2), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while not protected with a guardian address
	err = transferFundsWithGuardian(testContext, userAddress, big.NewInt(transferValue), receiverAddress, guardianAddress)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "guarded transaction not expected")
	require.Equal(t, big.NewInt(transferValue*2), getBalance(testContext, receiverAddress))

	// delay epoch pasts, the pending guardian is now active (but not activated), userAddress can send funds
	hdr := &block.Header{
		Epoch: vm.EpochGuardianDelay,
	}
	testContext.EpochNotifier.CheckEpoch(hdr)

	testActiveGuardian(t, testContext, userAddress, false, guardianAddress, uuid)
	testPendingGuardian(t, testContext, userAddress, true, nil, nil)

	err = transferFunds(testContext, userAddress, big.NewInt(transferValue), receiverAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue*3), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while protected without setting the guardian address
	returnCode, err = guardAccount(testContext, userAddress)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)

	testActiveGuardian(t, testContext, userAddress, false, guardianAddress, uuid)
	testPendingGuardian(t, testContext, userAddress, true, nil, nil)

	err = transferFunds(testContext, userAddress, big.NewInt(transferValue), receiverAddress)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "not allowed to bypass guardian")
	require.Equal(t, big.NewInt(transferValue*3), getBalance(testContext, receiverAddress))

	// userAddress can send funds while protected with the guardian address
	err = transferFundsWithGuardian(testContext, userAddress, big.NewInt(transferValue), receiverAddress, guardianAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue*4), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while protected with a wrong guardian address
	err = transferFundsWithGuardian(testContext, userAddress, big.NewInt(transferValue), receiverAddress, wrongGuardianAddress)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "mismatch between transaction guardian and configured account guardian")
	require.Equal(t, big.NewInt(transferValue*4), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while protected with an empty guardian address
	err = transferFundsWithGuardian(testContext, userAddress, big.NewInt(transferValue), receiverAddress, nil)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "mismatch between transaction guardian and configured account guardian")
	require.Equal(t, big.NewInt(transferValue*4), getBalance(testContext, receiverAddress))
}
