//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/guardians"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/guardian"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareTestContextForFreezeAccounts(tb testing.TB) *vm.VMTestContext {
	unreachableEpoch := uint32(999999)
	db := integrationtests.CreateStorer(tb.TempDir())
	gasScheduleDir := "../../../cmd/node/config/gasSchedules"

	cfg := config.GasScheduleByEpochs{
		StartEpoch: 0,
		FileName:   getLatestGasScheduleVersion(tb, gasScheduleDir),
	}

	argsGasScheduleNotifier := forking.ArgsNewGasScheduleNotifier{
		GasScheduleConfig: config.GasScheduleConfig{
			GasScheduleByEpochs: []config.GasScheduleByEpochs{cfg},
		},
		ConfigDir:          gasScheduleDir,
		EpochNotifier:      forking.NewGenericEpochNotifier(),
		WasmVMChangeLocker: &sync.RWMutex{},
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
		testscommon.NewMultiShardsCoordinatorMock(2),
		db,
		gasScheduleNotifier,
	)
	require.Nil(tb, err)

	return testContext
}

func getLatestGasScheduleVersion(tb testing.TB, directoryToSearch string) string {
	fileInfoSlice, err := ioutil.ReadDir(directoryToSearch)
	require.Nil(tb, err)

	gasSchedulePrefix := "gasScheduleV"

	files := make([]string, 0)
	for _, fileInfo := range fileInfoSlice {
		if fileInfo.IsDir() {
			continue
		}
		if !strings.Contains(fileInfo.Name(), gasSchedulePrefix) {
			continue
		}

		files = append(files, fileInfo.Name())
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j]
	})

	gasSchedule := files[0]
	log.Info("using gas schedule", "file", gasSchedule)

	return gasSchedule
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

func isAccountGuarded(tb testing.TB, testContext *vm.VMTestContext, address []byte) bool {
	accnt, err := testContext.Accounts.GetExistingAccount(address)
	require.Nil(tb, err)

	userAccnt := accnt.(state.UserAccountHandler)
	return userAccnt.IsGuarded()
}

func setGuardian(testContext *vm.VMTestContext, userAddress []byte, guardianAddress []byte, uuid []byte) (vmcommon.ReturnCode, error) {
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

	return testContext.TxProcessor.ProcessTransaction(tx)
}

func setGuardianCoSigned(
	testContext *vm.VMTestContext,
	userAddress []byte,
	currentGuardianAddress []byte,
	newGuardianAddress []byte,
	uuid []byte,
) (vmcommon.ReturnCode, error) {
	gasPrice := uint64(10)
	gasLimit := uint64(250000 + 50000 + 1000)

	tx := vm.CreateTransaction(
		getNonce(testContext, userAddress),
		big.NewInt(0),
		userAddress,
		userAddress,
		gasPrice,
		gasLimit,
		[]byte("SetGuardian@"+hex.EncodeToString(newGuardianAddress)+"@"+hex.EncodeToString(uuid)))

	tx.GuardianAddr = currentGuardianAddress
	tx.Options = tx.Options | transaction.MaskGuardedTransaction
	tx.Version = core.InitialVersionOfTransaction + 1

	return testContext.TxProcessor.ProcessTransaction(tx)
}

func removeGuardiansCoSigned(
	testContext *vm.VMTestContext,
	userAddress []byte,
	currentGuardianAddress []byte,
) (vmcommon.ReturnCode, error) {
	gasPrice := uint64(10)
	gasLimit := uint64(250000 + 50000 + 1000)

	tx := vm.CreateTransaction(
		getNonce(testContext, userAddress),
		big.NewInt(0),
		userAddress,
		userAddress,
		gasPrice,
		gasLimit,
		[]byte("UnGuardAccount"))

	tx.GuardianAddr = currentGuardianAddress
	tx.Options = tx.Options | transaction.MaskGuardedTransaction
	tx.Version = core.InitialVersionOfTransaction + 1

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

func transferFundsCoSigned(
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
	activationEpoch uint32,
) {
	active, _, err := getGuardiansData(tb, testContext, address)
	require.Nil(tb, err)

	testGuardianData(tb, active, shouldBeNil, expectedAddress, expectedUUID, activationEpoch)
}

func testGuardianData(
	tb testing.TB,
	guardian *guardians.Guardian,
	shouldBeNil bool,
	expectedAddress []byte,
	expectedUUID []byte,
	activationEpoch uint32,
) {
	if shouldBeNil {
		require.Nil(tb, guardian)
		return
	}

	require.NotNil(tb, guardian)
	assert.Equal(tb, expectedAddress, guardian.Address)
	assert.Equal(tb, expectedUUID, guardian.ServiceUID)
	assert.Equal(tb, activationEpoch, guardian.ActivationEpoch)
}

func testPendingGuardian(
	tb testing.TB,
	testContext *vm.VMTestContext,
	address []byte,
	shouldBeNil bool,
	expectedAddress []byte,
	expectedUUID []byte,
	activationEpoch uint32,
) {
	_, pending, err := getGuardiansData(tb, testContext, address)
	require.Nil(tb, err)

	testGuardianData(tb, pending, shouldBeNil, expectedAddress, expectedUUID, activationEpoch)
}

func setNewEpochOnContext(testContext *vm.VMTestContext, epoch uint32) {
	hdr := &block.Header{
		Epoch: epoch,
	}
	testContext.EpochNotifier.CheckEpoch(hdr)
	log.Info("current epoch is now", "epoch", epoch)
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

	returnCode, err := setGuardianCoSigned(testContext, userAddress, guardianAddress, guardianAddress, uuid)
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

	returnCode, err := setGuardian(testContext, userAddress, guardianAddress, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	currentEpoch := uint32(0)

	testActiveGuardian(t, testContext, userAddress, true, nil, nil, 0)
	testPendingGuardian(t, testContext, userAddress, false, guardianAddress, uuid, currentEpoch+vm.EpochGuardianDelay)

	// can not activate guardian now
	returnCode, err = guardAccount(testContext, userAddress)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)

	currentEpoch = vm.EpochGuardianDelay
	setNewEpochOnContext(testContext, currentEpoch)

	testActiveGuardian(t, testContext, userAddress, false, guardianAddress, uuid, currentEpoch)
	testPendingGuardian(t, testContext, userAddress, true, nil, nil, 0)

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
	err = transferFundsCoSigned(testContext, userAddress, big.NewInt(transferValue), receiverAddress, guardianAddress)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "guarded transaction not expected")
	require.Equal(t, big.NewInt(transferValue), getBalance(testContext, receiverAddress))

	// userAddress can send funds while it just added a guardian
	returnCode, err := setGuardian(testContext, userAddress, guardianAddress, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	currentEpoch := uint32(0)

	testActiveGuardian(t, testContext, userAddress, true, nil, nil, 0)
	testPendingGuardian(t, testContext, userAddress, false, guardianAddress, uuid, currentEpoch+vm.EpochGuardianDelay)

	err = transferFunds(testContext, userAddress, big.NewInt(transferValue), receiverAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue*2), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while not protected with a guardian address
	err = transferFundsCoSigned(testContext, userAddress, big.NewInt(transferValue), receiverAddress, guardianAddress)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "guarded transaction not expected")
	require.Equal(t, big.NewInt(transferValue*2), getBalance(testContext, receiverAddress))

	// delay epoch pasts, the pending guardian is now active (but not activated), userAddress can send funds
	currentEpoch = vm.EpochGuardianDelay
	setNewEpochOnContext(testContext, currentEpoch)

	testActiveGuardian(t, testContext, userAddress, false, guardianAddress, uuid, currentEpoch)
	testPendingGuardian(t, testContext, userAddress, true, nil, nil, 0)

	err = transferFunds(testContext, userAddress, big.NewInt(transferValue), receiverAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue*3), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while protected without setting the guardian address
	returnCode, err = guardAccount(testContext, userAddress)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)

	testActiveGuardian(t, testContext, userAddress, false, guardianAddress, uuid, currentEpoch)
	testPendingGuardian(t, testContext, userAddress, true, nil, nil, 0)

	err = transferFunds(testContext, userAddress, big.NewInt(transferValue), receiverAddress)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "not allowed to bypass guardian")
	require.Equal(t, big.NewInt(transferValue*3), getBalance(testContext, receiverAddress))

	// userAddress can send funds while protected with the guardian address
	err = transferFundsCoSigned(testContext, userAddress, big.NewInt(transferValue), receiverAddress, guardianAddress)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue*4), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while protected with a wrong guardian address
	err = transferFundsCoSigned(testContext, userAddress, big.NewInt(transferValue), receiverAddress, wrongGuardianAddress)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "mismatch between transaction guardian and configured account guardian")
	require.Equal(t, big.NewInt(transferValue*4), getBalance(testContext, receiverAddress))

	// userAddress can not send funds while protected with an empty guardian address
	err = transferFundsCoSigned(testContext, userAddress, big.NewInt(transferValue), receiverAddress, nil)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "mismatch between transaction guardian and configured account guardian")
	require.Equal(t, big.NewInt(transferValue*4), getBalance(testContext, receiverAddress))
}

// Scenario 1 description:
// 1.  create & mint 4 addresses: alice, bob, charlie abd delta
// 2.  alice sets bob as guardian (test if pending)
// 3.  alice can not set bob as guardian again (test if pending & same activation epoch)
//   3.1 alice can not set bob as guardian again even if one epoch past
// 4.  alice activates the guardian (test if active)
// 5.  alice sets charlie as pending guardian (test if pending & different activation epoch)
//   5.1. alice wants to set delta as pending guardian (transaction is not executable, will not be included in a miniblock)
// 6.  alice sets charlie as guardian immediately through a cosigned transaction (test active & pending guardians)
// 7.  alice immediately sets bob as guardian through a cosigned transaction (test active & pending guardians)
// 8.  alice adds charlie as a pending guardian (test if pending & different activation epoch)
//     wait until charlie becomes active, no more pending guardians
// 9.  alice adds bob as a pending guardian and calls set charlie immediately cosigned and should remove the pending guardian
// 10. alice un-guards the account immediately by using a cosigned transaction
// 11. alice guards the account immediately by calling the GuardAccount function
// 13. alice sends a guarded transaction, while account is guarded -> should work
// 14. alice un-guards the accounts immediately using a cosigned transaction and then sends a guarded transaction -> should error
//   14.1 alice sends unguarded transaction -> should work
func TestGuardAccount_Scenario1(t *testing.T) {
	testContext := prepareTestContextForFreezeAccounts(t)
	defer testContext.Close()

	uuid := []byte("uuid")
	// transferValue := int64(2000000)
	initialMint := big.NewInt(1000000000000000000)

	alice := []byte("alice-12345678901234567890123456")
	bob := []byte("bob-1234567890123456789012345678")
	charlie := []byte("charlie-123456789012345678901234")
	delta := []byte("delta-12345678901234567890123456")

	// step 1 -  mint addresses
	allAddresses := [][]byte{alice, bob, charlie}
	for _, address := range allAddresses {
		mintAddress(t, testContext, address, initialMint)
	}
	for _, address := range allAddresses {
		testNoGuardianIsSet(t, testContext, address)
	}
	currentEpoch := uint32(0)

	// step 2 - alice sets bob as guardian
	step2Epoch := currentEpoch
	returnCode, err := setGuardian(testContext, alice, bob, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.False(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, true, nil, nil, 0)
	testPendingGuardian(t, testContext, alice, false, bob, uuid, step2Epoch+vm.EpochGuardianDelay)

	// step 3 - alice wants to set bob as guardian again - should fail
	returnCode, err = setGuardian(testContext, alice, bob, uuid)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.False(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, true, nil, nil, 0)
	testPendingGuardian(t, testContext, alice, false, bob, uuid, step2Epoch+vm.EpochGuardianDelay)

	// step 3.1 - one epoch pass, try to make bob again as guardian
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = setGuardian(testContext, alice, bob, uuid)
	assert.Equal(t, process.ErrFailedTransaction, err)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.False(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, true, nil, nil, currentEpoch)
	testPendingGuardian(t, testContext, alice, false, bob, uuid, step2Epoch+vm.EpochGuardianDelay) // initial epoch + EpochGuardianDelay

	// step 4 - alice activates the guardian
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = guardAccount(testContext, alice)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.True(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, bob, uuid, step2Epoch+vm.EpochGuardianDelay) // initial epoch + EpochGuardianDelay
	testPendingGuardian(t, testContext, alice, true, nil, nil, 0)

	// step 5 - alice sets charlie as pending guardian
	step5Epoch := currentEpoch
	returnCode, err = setGuardian(testContext, alice, charlie, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.True(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, bob, uuid, step2Epoch+vm.EpochGuardianDelay)
	testPendingGuardian(t, testContext, alice, false, charlie, uuid, step5Epoch+vm.EpochGuardianDelay)

	// step 5.1 - alice tries to set delta as pending guardian, overwriting charlie
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = setGuardian(testContext, alice, delta, uuid)
	assert.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	assert.Equal(t, vmcommon.UserError, returnCode)
	assert.True(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, bob, uuid, step2Epoch+vm.EpochGuardianDelay)
	testPendingGuardian(t, testContext, alice, false, charlie, uuid, step5Epoch+vm.EpochGuardianDelay)

	// step 6 - alice sets charlie as guardian immediately through a cosigned transaction
	step6Epoch := currentEpoch
	returnCode, err = setGuardianCoSigned(testContext, alice, bob, charlie, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.True(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, charlie, uuid, step6Epoch) // instant set, no delay add
	testPendingGuardian(t, testContext, alice, true, nil, nil, 0)

	// step 7 - alice immediately sets bob as guardian through a cosigned transaction
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	step7Epoch := currentEpoch
	returnCode, err = setGuardianCoSigned(testContext, alice, charlie, bob, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.True(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, bob, uuid, step7Epoch) // instant set, no delay add
	testPendingGuardian(t, testContext, alice, true, nil, nil, 0)

	// step 8 - alice adds charlie as a pending guardian (test if pending & different activation epoch)
	step8Epoch := currentEpoch
	returnCode, err = setGuardian(testContext, alice, charlie, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	testActiveGuardian(t, testContext, alice, false, bob, uuid, step7Epoch)
	testPendingGuardian(t, testContext, alice, false, charlie, uuid, step8Epoch+vm.EpochGuardianDelay)
	currentEpoch += vm.EpochGuardianDelay
	setNewEpochOnContext(testContext, currentEpoch)
	assert.True(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, charlie, uuid, step8Epoch+vm.EpochGuardianDelay)
	testPendingGuardian(t, testContext, alice, true, nil, nil, 0)

	// step 9 - alice adds bob as a pending guardian and calls set charlie immediately cosigned and should remove the pending guardian
	step9Epoch := currentEpoch
	returnCode, err = setGuardian(testContext, alice, bob, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	testActiveGuardian(t, testContext, alice, false, charlie, uuid, step8Epoch+vm.EpochGuardianDelay)
	testPendingGuardian(t, testContext, alice, false, bob, uuid, step9Epoch+vm.EpochGuardianDelay)
	// guard account by charlie should remove bob pending guardian
	returnCode, err = setGuardianCoSigned(testContext, alice, charlie, charlie, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.True(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, charlie, uuid, step8Epoch+vm.EpochGuardianDelay)
	testPendingGuardian(t, testContext, alice, true, nil, nil, 0)

	// step 10 - alice un-guards the account immediately by using a cosigned transaction
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = removeGuardiansCoSigned(testContext, alice, charlie)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.False(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, charlie, uuid, step8Epoch+vm.EpochGuardianDelay)
	testPendingGuardian(t, testContext, alice, true, nil, nil, 0)

	// step 11 - alice guards the account immediately by calling the GuardAccount function
	returnCode, err = guardAccount(testContext, alice)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.True(t, isAccountGuarded(t, testContext, alice))
	testActiveGuardian(t, testContext, alice, false, charlie, uuid, step8Epoch+vm.EpochGuardianDelay)
	testPendingGuardian(t, testContext, alice, true, nil, nil, 0)

	// 13. alice sends a guarded transaction, while account is guarded -> should work
	err = transferFundsCoSigned(testContext, alice, big.NewInt(100), delta, charlie)
	assert.Nil(t, err)

	// 14. alice un-guards the accounts immediately using a cosigned transaction and then sends a guarded transaction -> should error
	returnCode, err = removeGuardiansCoSigned(testContext, alice, charlie)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	assert.False(t, isAccountGuarded(t, testContext, alice))
	err = transferFundsCoSigned(testContext, alice, big.NewInt(100), delta, charlie)
	assert.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	// 14.1 alice sends unguarded transaction -> should work
	err = transferFunds(testContext, alice, big.NewInt(100), delta)
	assert.Nil(t, err)
}
