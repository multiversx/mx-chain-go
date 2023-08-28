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
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/guardian"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	testscommonIntegrationTests "github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const txWithOptionVersion = 2
const guardianSigVerificationGas = uint64(50000)
const guardAccountGas = uint64(250000)
const unGuardAccountGas = uint64(250000)
const setGuardianGas = uint64(250000)
const transferGas = uint64(1000)

var (
	alice        = []byte("alice-12345678901234567890123456")
	bob          = []byte("bob-1234567890123456789012345678")
	charlie      = []byte("charlie-123456789012345678901234")
	david        = []byte("david-12345678901234567890123456")
	allAddresses = map[string][]byte{
		"alice":   alice,
		"bob":     bob,
		"charlie": charlie,
		"david":   david,
	}
	uuid          = []byte("uuid")
	transferValue = big.NewInt(2000000)
	initialMint   = big.NewInt(1000000000000000000)
)

type guardianInfo struct {
	address []byte
	uuid    []byte
	epoch   uint32
}

type guardAccountStatus struct {
	isGuarded bool
	active    *guardianInfo
	pending   *guardianInfo
}

func createUnGuardedAccountStatus() guardAccountStatus {
	return guardAccountStatus{
		isGuarded: false,
		active:    nil,
		pending:   nil,
	}
}

func prepareTestContextForGuardedAccounts(tb testing.TB) *vm.VMTestContext {
	unreachableEpoch := uint32(999999)
	db := testscommonIntegrationTests.CreateStorer(tb.TempDir())
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

	testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGasAndRoundConfig(
		config.EnableEpochs{
			GovernanceEnableEpoch:                   unreachableEpoch,
			WaitingListFixEnableEpoch:               unreachableEpoch,
			SetSenderInEeiOutputTransferEnableEpoch: unreachableEpoch,
			RefactorPeersMiniBlocksEnableEpoch:      unreachableEpoch,
		},
		testscommon.NewMultiShardsCoordinatorMock(2),
		db,
		gasScheduleNotifier,
		integrationTests.GetDefaultRoundsConfig(),
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

func mintAddress(tb testing.TB, testContext *vm.VMTestContext, address []byte, value *big.Int, addressDescription string) {
	addressString := integrationTests.TestAddressPubkeyConverter.SilentEncode(address, log)
	log.Info("minting "+addressDescription+" address", "address", addressString, "value", value)

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

func getGuardiansData(tb testing.TB, testContext *vm.VMTestContext, address []byte) (*guardians.Guardian, *guardians.Guardian, bool) {
	accnt, err := testContext.Accounts.GetExistingAccount(address)
	require.Nil(tb, err)

	userAccnt := accnt.(state.UserAccountHandler)
	guardedAccount, err := guardian.NewGuardedAccount(
		testContext.Marshalizer,
		testContext.EpochNotifier,
		vm.EpochGuardianDelay)
	require.Nil(tb, err)

	active, pending, err := guardedAccount.GetConfiguredGuardians(userAccnt)
	require.Nil(tb, err)

	return active, pending, userAccnt.IsGuarded()
}

func setGuardian(testContext *vm.VMTestContext, userAddress []byte, guardianAddress []byte, uuid []byte) (vmcommon.ReturnCode, error) {
	gasLimit := setGuardianGas + transferGas

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
	gasLimit := setGuardianGas + guardianSigVerificationGas + transferGas

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
	tx.Version = txWithOptionVersion

	return testContext.TxProcessor.ProcessTransaction(tx)
}

func removeGuardiansCoSigned(
	testContext *vm.VMTestContext,
	userAddress []byte,
	currentGuardianAddress []byte,
) (vmcommon.ReturnCode, error) {
	gasLimit := unGuardAccountGas + guardianSigVerificationGas + transferGas

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
	tx.Version = txWithOptionVersion

	return testContext.TxProcessor.ProcessTransaction(tx)
}

func guardAccount(testContext *vm.VMTestContext, userAddress []byte) (vmcommon.ReturnCode, error) {
	gasLimit := guardAccountGas + transferGas

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
	tx := vm.CreateTransaction(
		getNonce(testContext, senderAddress),
		transferValue,
		senderAddress,
		receiverAddress,
		gasPrice,
		transferGas,
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
	gasLimit := guardianSigVerificationGas + transferGas

	tx := vm.CreateTransaction(
		getNonce(testContext, senderAddress),
		transferValue,
		senderAddress,
		receiverAddress,
		gasPrice,
		gasLimit,
		make([]byte, 0))
	tx.Version = txWithOptionVersion
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

func testGuardianStatus(
	tb testing.TB,
	testContext *vm.VMTestContext,
	address []byte,
	expectedStatus guardAccountStatus,
) {
	active, pending, isGuarded := getGuardiansData(tb, testContext, address)
	assert.Equal(tb, expectedStatus.isGuarded, isGuarded)

	testGuardianData(tb, active, expectedStatus.active)
	testGuardianData(tb, pending, expectedStatus.pending)
}

func testGuardianData(
	tb testing.TB,
	guardian *guardians.Guardian,
	info *guardianInfo,
) {
	if info == nil {
		require.Nil(tb, guardian)
		return
	}

	require.NotNil(tb, guardian)
	expectedAddress := integrationTests.TestAddressPubkeyConverter.SilentEncode(info.address, log)
	providedAddress := integrationTests.TestAddressPubkeyConverter.SilentEncode(guardian.Address, log)
	assert.Equal(tb, expectedAddress, providedAddress)
	assert.Equal(tb, info.uuid, guardian.ServiceUID)
	assert.Equal(tb, info.epoch, guardian.ActivationEpoch)
}

func setNewEpochOnContext(testContext *vm.VMTestContext, epoch uint32) {
	hdr := &block.Header{
		Epoch: epoch,
	}
	testContext.EpochNotifier.CheckEpoch(hdr)
	log.Info("current epoch is now", "epoch", epoch)
}

func TestGuardAccount_ShouldErrorIfInstantSetIsDoneOnANotProtectedAccount(t *testing.T) {
	testContext := prepareTestContextForGuardedAccounts(t)
	defer testContext.Close()

	// alice is the user, bob is the guardian
	mintAddress(t, testContext, alice, initialMint, "alice")

	expectedStatus := createUnGuardedAccountStatus()
	testGuardianStatus(t, testContext, alice, expectedStatus)

	returnCode, err := setGuardianCoSigned(testContext, alice, bob, bob, uuid)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Equal(t, vmcommon.UserError, returnCode)

	testGuardianStatus(t, testContext, alice, expectedStatus)
}

func TestGuardAccount_ShouldSetGuardianOnANotProtectedAccount(t *testing.T) {
	testContext := prepareTestContextForGuardedAccounts(t)
	defer testContext.Close()

	// alice is the user, bob is the guardian
	mintAddress(t, testContext, alice, initialMint, "alice")

	expectedStatus := createUnGuardedAccountStatus()
	testGuardianStatus(t, testContext, alice, expectedStatus)

	returnCode, err := setGuardian(testContext, alice, bob, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	currentEpoch := uint32(0)

	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active:    nil,
		pending: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   currentEpoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	allLogs := testContext.TxsLogsProcessor.GetAllCurrentLogs()
	require.NotNil(t, allLogs)

	event := allLogs[0].LogHandler.GetLogEvents()[0]
	require.Equal(t, &transaction.Event{
		Address:        alice,
		Identifier:     []byte(core.BuiltInFunctionSetGuardian),
		Topics:         [][]byte{bob, uuid},
		Data:           nil,
		AdditionalData: nil,
	}, event)
	testContext.TxsLogsProcessor.Clean()

	// can not activate guardian now
	returnCode, err = guardAccount(testContext, alice)
	require.Equal(t, process.ErrFailedTransaction, err)
	require.Equal(t, vmcommon.UserError, returnCode)

	currentEpoch = vm.EpochGuardianDelay
	setNewEpochOnContext(testContext, currentEpoch)

	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   currentEpoch,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	allLogs = testContext.TxsLogsProcessor.GetAllCurrentLogs()
	require.NotNil(t, allLogs)

	event = allLogs[0].LogHandler.GetLogEvents()[0]
	require.Equal(t, &transaction.Event{
		Address:        alice,
		Identifier:     []byte(core.SignalErrorOperation),
		Topics:         [][]byte{alice, []byte("account has no active guardian")},
		Data:           []byte("@6163636f756e7420686173206e6f2061637469766520677561726469616e"),
		AdditionalData: [][]byte{[]byte("@6163636f756e7420686173206e6f2061637469766520677561726469616e")},
	}, event)
	testContext.TxsLogsProcessor.Clean()

	// can activate guardian now
	returnCode, err = guardAccount(testContext, alice)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)

	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   currentEpoch,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	allLogs = testContext.TxsLogsProcessor.GetAllCurrentLogs()
	require.NotNil(t, allLogs)

	event = allLogs[0].LogHandler.GetLogEvents()[0]
	require.Equal(t, &transaction.Event{
		Address:        alice,
		Identifier:     []byte(core.BuiltInFunctionGuardAccount),
		Data:           nil,
		AdditionalData: nil,
	}, event)
	testContext.TxsLogsProcessor.Clean()
}

func TestGuardAccount_SendingFundsWhileProtectedAndNotProtected(t *testing.T) {
	testContext := prepareTestContextForGuardedAccounts(t)
	defer testContext.Close()

	// alice is the user, bob is the guardian, charlie is the receiver, david is the wrong guardian
	mintAddress(t, testContext, alice, initialMint, "alice")

	expectedStatus := createUnGuardedAccountStatus()
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// userAddress can send funds while not protected
	err := transferFunds(testContext, alice, transferValue, charlie)
	require.Nil(t, err)
	require.Equal(t, transferValue, getBalance(testContext, charlie))

	// userAddress can not send funds while not protected with a guardian address
	err = transferFundsCoSigned(testContext, alice, transferValue, charlie, bob)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "guarded transaction not expected")
	require.Equal(t, transferValue, getBalance(testContext, charlie))

	// userAddress can send funds while it just added a guardian
	returnCode, err := setGuardian(testContext, alice, bob, uuid)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnCode)
	currentEpoch := uint32(0)

	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active:    nil,
		pending: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   currentEpoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	err = transferFunds(testContext, alice, transferValue, charlie)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue.Int64()*2), getBalance(testContext, charlie))

	// userAddress can not send funds while not protected with a guardian address
	err = transferFundsCoSigned(testContext, alice, transferValue, charlie, bob)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "guarded transaction not expected")
	require.Equal(t, big.NewInt(transferValue.Int64()*2), getBalance(testContext, charlie))

	// delay epoch pasts, the pending guardian is now active (but not activated), userAddress can send funds
	currentEpoch = vm.EpochGuardianDelay
	setNewEpochOnContext(testContext, currentEpoch)

	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   currentEpoch,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	err = transferFunds(testContext, alice, transferValue, charlie)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue.Int64()*3), getBalance(testContext, charlie))

	// userAddress can not send funds while protected without setting the guardian address
	returnCode, err = guardAccount(testContext, alice)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)

	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   currentEpoch,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	err = transferFunds(testContext, alice, transferValue, charlie)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "not allowed to bypass guardian")
	require.Equal(t, big.NewInt(transferValue.Int64()*3), getBalance(testContext, charlie))

	// userAddress can send funds while protected with the guardian address
	err = transferFundsCoSigned(testContext, alice, transferValue, charlie, bob)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(transferValue.Int64()*4), getBalance(testContext, charlie))

	// userAddress can not send funds while protected with a wrong guardian address (david)
	err = transferFundsCoSigned(testContext, alice, transferValue, charlie, david)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "mismatch between transaction guardian and configured account guardian")
	require.Equal(t, big.NewInt(transferValue.Int64()*4), getBalance(testContext, charlie))

	// userAddress can not send funds while protected with an empty guardian address
	err = transferFundsCoSigned(testContext, alice, transferValue, charlie, nil)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Contains(t, err.Error(), "mismatch between transaction guardian and configured account guardian")
	require.Equal(t, big.NewInt(transferValue.Int64()*4), getBalance(testContext, charlie))
}

// Scenario 1 description:
//  1. create & mint 4 addresses: alice, bob, charlie and david
//  2. alice sets bob as guardian (test if pending)
//  3. alice can not set bob as guardian again (test if pending & same activation epoch)
//     3.1 alice can not set bob as guardian again even if one epoch past
//  4. alice activates the guardian (test if active)
//  5. alice sets charlie as pending guardian (test if pending & different activation epoch)
//     5.1. alice wants to set david as pending guardian (transaction is not executable, will not be included in a miniblock)
//  6. alice sets charlie as guardian immediately through a cosigned transaction (test active & pending guardians)
//  7. alice immediately sets bob as guardian through a cosigned transaction (test active & pending guardians)
//  8. alice adds charlie as a pending guardian (test if pending & different activation epoch)
//     wait until charlie becomes active, no more pending guardians
//  9. alice adds bob as a pending guardian and calls set charlie immediately cosigned and should remove the pending guardian
//  10. alice un-guards the account immediately by using a cosigned transaction
//  11. alice guards the account immediately by calling the GuardAccount function
//  12. alice sends a malformed set guardian transaction, should not consume gas
//  13. alice sends a guarded transaction, while account is guarded -> should work
//  14. alice un-guards the accounts immediately using a cosigned transaction and then sends a guarded transaction -> should error
//     14.1 alice sends unguarded transaction -> should work
func TestGuardAccount_Scenario1(t *testing.T) {
	testContext := prepareTestContextForGuardedAccounts(t)
	defer testContext.Close()

	// step 1 -  mint addresses
	for addressDescription, address := range allAddresses {
		mintAddress(t, testContext, address, initialMint, addressDescription)
	}
	expectedStatus := createUnGuardedAccountStatus()
	for _, address := range allAddresses {
		testGuardianStatus(t, testContext, address, expectedStatus)
	}
	currentEpoch := uint32(0)

	// step 2 - alice sets bob as guardian
	step2Epoch := currentEpoch
	returnCode, err := setGuardian(testContext, alice, bob, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active:    nil,
		pending: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step2Epoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 3 - alice wants to set bob as guardian again - should fail
	returnCode, err = setGuardian(testContext, alice, bob, uuid)
	require.Equal(t, process.ErrFailedTransaction, err)
	require.Equal(t, vmcommon.UserError, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active:    nil,
		pending: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step2Epoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 3.1 - one epoch pass, try to make bob again as guardian
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = setGuardian(testContext, alice, bob, uuid)
	require.Equal(t, process.ErrFailedTransaction, err)
	require.Equal(t, vmcommon.UserError, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active:    nil,
		pending: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step2Epoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 4 - alice activates the guardian
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = guardAccount(testContext, alice)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step2Epoch + vm.EpochGuardianDelay,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 5 - alice sets charlie as pending guardian
	step5Epoch := currentEpoch
	returnCode, err = setGuardian(testContext, alice, charlie, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step2Epoch + vm.EpochGuardianDelay,
		},
		pending: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step5Epoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 5.1 - alice tries to set david as pending guardian, overwriting charlie
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = setGuardian(testContext, alice, david, uuid)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	require.Equal(t, vmcommon.UserError, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step2Epoch + vm.EpochGuardianDelay,
		},
		pending: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step5Epoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 6 - alice sets charlie as guardian immediately through a cosigned transaction
	step6Epoch := currentEpoch
	returnCode, err = setGuardianCoSigned(testContext, alice, bob, charlie, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{ // instant set, no delay added
			address: charlie,
			uuid:    uuid,
			epoch:   step6Epoch,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 7 - alice immediately sets bob as guardian through a cosigned transaction
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	step7Epoch := currentEpoch
	returnCode, err = setGuardianCoSigned(testContext, alice, charlie, bob, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{ // instant set, no delay added
			address: bob,
			uuid:    uuid,
			epoch:   step7Epoch,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 8 - alice adds charlie as a pending guardian (test if pending & different activation epoch)
	step8Epoch := currentEpoch
	returnCode, err = setGuardian(testContext, alice, charlie, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step7Epoch,
		},
		pending: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step8Epoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)
	currentEpoch += vm.EpochGuardianDelay
	setNewEpochOnContext(testContext, currentEpoch)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step8Epoch + vm.EpochGuardianDelay,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 9 - alice adds bob as a pending guardian and calls set charlie immediately cosigned and should remove the pending guardian
	step9Epoch := currentEpoch
	returnCode, err = setGuardian(testContext, alice, bob, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step8Epoch + vm.EpochGuardianDelay,
		},
		pending: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step9Epoch + vm.EpochGuardianDelay,
		},
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)
	// guard account by charlie should remove bob pending guardian
	returnCode, err = setGuardianCoSigned(testContext, alice, charlie, charlie, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step8Epoch + vm.EpochGuardianDelay,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 10 - alice un-guards the account immediately by using a cosigned transaction
	currentEpoch++
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = removeGuardiansCoSigned(testContext, alice, charlie)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step8Epoch + vm.EpochGuardianDelay,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// step 11 - alice guards the account immediately by calling the GuardAccount function
	returnCode, err = guardAccount(testContext, alice)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step8Epoch + vm.EpochGuardianDelay,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	// 12. alice sends a malformed set guardian transaction, should not consume gas
	// 12.1. transaction with value
	nonceAlice := getNonce(testContext, alice)
	setGuardianData := []byte("SetGuardian@" + hex.EncodeToString(bob) + "@" + hex.EncodeToString(uuid))
	tx := &transaction.Transaction{
		Nonce:    nonceAlice,
		Value:    big.NewInt(10),
		RcvAddr:  alice,
		SndAddr:  alice,
		GasPrice: gasPrice,
		GasLimit: setGuardianGas + transferGas,
		Data:     setGuardianData,
	}
	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, returnCode)
	require.NotNil(t, err)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	newNonceAlice := getNonce(testContext, alice)
	// tx not executed
	require.Equal(t, nonceAlice, newNonceAlice)

	// 12.2. too many parameters in data, last one not hex (tx with notarization not builtin func call)
	tx.Value = big.NewInt(0)
	tx.Data = []byte(string(setGuardianData) + "@extra")
	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, returnCode)
	require.NotNil(t, err)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	newNonceAlice = getNonce(testContext, alice)
	// tx not executed
	require.Equal(t, nonceAlice, newNonceAlice)

	// 12.3. too many parameters in data, failed builtin func call
	tx.Data = []byte(string(setGuardianData) + "@00")
	returnCode, err = testContext.TxProcessor.ProcessTransaction(tx)
	require.Equal(t, vmcommon.UserError, returnCode)
	require.NotNil(t, err)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	newNonceAlice = getNonce(testContext, alice)
	// tx not executed
	require.Equal(t, nonceAlice, newNonceAlice)

	// 13. alice sends a guarded transaction, while account is guarded -> should work
	err = transferFundsCoSigned(testContext, alice, transferValue, david, charlie)
	require.Nil(t, err)

	// 14. alice un-guards the accounts immediately using a cosigned transaction and then sends a guarded transaction -> should error
	returnCode, err = removeGuardiansCoSigned(testContext, alice, charlie)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: false,
		active: &guardianInfo{
			address: charlie,
			uuid:    uuid,
			epoch:   step8Epoch + vm.EpochGuardianDelay,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)
	err = transferFundsCoSigned(testContext, alice, transferValue, david, charlie)
	require.ErrorIs(t, err, process.ErrTransactionNotExecutable)
	// 14.1 alice sends unguarded transaction -> should work
	err = transferFunds(testContext, alice, transferValue, david)
	require.Nil(t, err)
}

//  1. create & mint 4 addresses: alice, bob, charlie and david
//  2. alice sets bob as guardian and the account becomes guarded
//  3. test that charlie can send a relayed transaction v1 on the behalf of alice to david
//     3.1 cosigned transaction should work
//     3.2 single signed transaction should not work
func TestGuardAccounts_RelayedTransactionV1(t *testing.T) {
	testContext := prepareTestContextForGuardedAccounts(t)
	defer testContext.Close()

	// step 1 -  mint addresses
	for addressDescription, address := range allAddresses {
		mintAddress(t, testContext, address, initialMint, addressDescription)
	}
	expectedStatus := createUnGuardedAccountStatus()
	for _, address := range allAddresses {
		testGuardianStatus(t, testContext, address, expectedStatus)
	}

	currentEpoch := uint32(0)

	// step 2 - alice sets bob as guardian
	step2Epoch := currentEpoch
	returnCode, err := setGuardian(testContext, alice, bob, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	currentEpoch += vm.EpochGuardianDelay
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = guardAccount(testContext, alice)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step2Epoch + vm.EpochGuardianDelay,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	aliceCurrentBalance := getBalance(testContext, alice)

	// step 3 - charlie sends a relayed transaction v1 on the behalf of alice
	// 3.1 cosigned transaction should work
	userTx := vm.CreateTransaction(
		getNonce(testContext, alice),
		transferValue,
		alice,
		david,
		gasPrice,
		transferGas+guardianSigVerificationGas,
		make([]byte, 0))

	userTx.GuardianAddr = bob
	userTx.Options = userTx.Options | transaction.MaskGuardedTransaction
	userTx.Version = txWithOptionVersion

	rtxData := integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit := 1 + transferGas + guardianSigVerificationGas + uint64(len(rtxData))
	rtx := vm.CreateTransaction(getNonce(testContext, charlie), big.NewInt(0), charlie, alice, gasPrice, rTxGasLimit, rtxData)
	returnCode, err = testContext.TxProcessor.ProcessTransaction(rtx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	// balance tests:
	//  alice: aliceCurrentBalance - transferValue (no fee for relayed transaction)
	//  bob: initialMint
	//  charlie: initialMint - rtxGasLimit * gasPrice
	//  david: initialMint + transferValue
	aliceExpectedBalance := big.NewInt(0).Sub(aliceCurrentBalance, transferValue)
	assert.Equal(t, aliceExpectedBalance, getBalance(testContext, alice))
	bobExpectedBalance := big.NewInt(0).Set(initialMint)
	assert.Equal(t, bobExpectedBalance, getBalance(testContext, bob))
	charlieExpectedBalance := big.NewInt(0).Sub(initialMint, big.NewInt(int64(rTxGasLimit*gasPrice)))
	assert.Equal(t, charlieExpectedBalance, getBalance(testContext, charlie))
	davidExpectedBalance := big.NewInt(0).Add(initialMint, transferValue)
	assert.Equal(t, davidExpectedBalance, getBalance(testContext, david))

	aliceCurrentBalance = getBalance(testContext, alice)
	charlieCurrentBalance := getBalance(testContext, charlie)
	davidCurrentBalance := getBalance(testContext, david)
	testContext.CleanIntermediateTransactions(t)

	// 3.1 single signed transaction should not work
	userTx = vm.CreateTransaction(
		getNonce(testContext, alice),
		transferValue,
		alice,
		david,
		gasPrice,
		transferGas+guardianSigVerificationGas,
		make([]byte, 0))

	userTx.Version = txWithOptionVersion

	rtxData = integrationTests.PrepareRelayedTxDataV1(userTx)
	rTxGasLimit = 1 + transferGas + guardianSigVerificationGas + uint64(len(rtxData))
	rtx = vm.CreateTransaction(getNonce(testContext, charlie), big.NewInt(0), charlie, alice, gasPrice, rTxGasLimit, rtxData)
	returnCode, err = testContext.TxProcessor.ProcessTransaction(rtx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, returnCode)
	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))
	scr := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	// expectedReturnMessage is hardcoded for backwards compatibility reasons
	expectedReturnMessage := "transaction is not executable and gas will not be consumed, not allowed to bypass guardian"
	require.Equal(t, expectedReturnMessage, string(scr.ReturnMessage))
	// balance tests:
	//  alice: aliceCurrentBalance (no fee for the failed relayed transaction)
	//  bob: initialMint
	//  charlie: charlieCurrentBalance - rtxGasLimit * gasPrice
	//  david: davidCurrentBalance
	assert.Equal(t, aliceCurrentBalance, getBalance(testContext, alice))
	bobExpectedBalance = big.NewInt(0).Set(initialMint)
	assert.Equal(t, bobExpectedBalance, getBalance(testContext, bob))
	charlieExpectedBalance = big.NewInt(0).Sub(charlieCurrentBalance, big.NewInt(int64(rTxGasLimit*gasPrice)))
	assert.Equal(t, charlieExpectedBalance, getBalance(testContext, charlie))
	assert.Equal(t, davidCurrentBalance, getBalance(testContext, david))
}

//  1. create & mint 4 addresses: alice, bob, charlie and david
//  2. alice sets bob as guardian and the account becomes guarded
//  3. test that charlie can not send a relayed transaction v2 on the behalf of alice to david
//     3.1 cosigned transaction should not work
//     3.2 single signed transaction should not work
func TestGuardAccounts_RelayedTransactionV2(t *testing.T) {
	testContext := prepareTestContextForGuardedAccounts(t)
	defer testContext.Close()

	// step 1 -  mint addresses
	for addressDescription, address := range allAddresses {
		mintAddress(t, testContext, address, initialMint, addressDescription)
	}
	expectedStatus := createUnGuardedAccountStatus()
	for _, address := range allAddresses {
		testGuardianStatus(t, testContext, address, expectedStatus)
	}

	currentEpoch := uint32(0)

	// step 2 - alice sets bob as guardian
	step2Epoch := currentEpoch
	returnCode, err := setGuardian(testContext, alice, bob, uuid)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	currentEpoch += vm.EpochGuardianDelay
	setNewEpochOnContext(testContext, currentEpoch)
	returnCode, err = guardAccount(testContext, alice)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)
	expectedStatus = guardAccountStatus{
		isGuarded: true,
		active: &guardianInfo{
			address: bob,
			uuid:    uuid,
			epoch:   step2Epoch + vm.EpochGuardianDelay,
		},
		pending: nil,
	}
	testGuardianStatus(t, testContext, alice, expectedStatus)

	aliceCurrentBalance := getBalance(testContext, alice)
	testContext.CleanIntermediateTransactions(t)

	// step 3 - charlie sends a relayed transaction v1 on the behalf of alice
	// 3.1 cosigned transaction should work
	userTx := vm.CreateTransaction(
		getNonce(testContext, alice),
		transferValue,
		alice,
		david,
		gasPrice,
		transferGas+guardianSigVerificationGas,
		make([]byte, 0))

	userTx.GuardianAddr = bob
	userTx.Options = userTx.Options | transaction.MaskGuardedTransaction
	userTx.Version = txWithOptionVersion

	rtxData := integrationTests.PrepareRelayedTxDataV2(userTx)
	rTxGasLimit := 1 + transferGas + guardianSigVerificationGas + uint64(len(rtxData))
	rtx := vm.CreateTransaction(getNonce(testContext, charlie), big.NewInt(0), charlie, alice, gasPrice, rTxGasLimit, rtxData)
	returnCode, err = testContext.TxProcessor.ProcessTransaction(rtx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, returnCode)
	intermediateTxs := testContext.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))
	scr := intermediateTxs[0].(*smartContractResult.SmartContractResult)
	// expectedReturnMessage is hardcoded for backwards compatibility reasons
	expectedReturnMessage := "transaction is not executable and gas will not be consumed, not allowed to bypass guardian"
	require.Equal(t, expectedReturnMessage, string(scr.ReturnMessage))
	// balance tests:
	//  alice: aliceCurrentBalance (no fee for relayed transaction V2)
	//  bob: initialMint
	//  charlie: initialMint - rtxGasLimit * gasPrice
	//  david: initialMint
	assert.Equal(t, aliceCurrentBalance, getBalance(testContext, alice))
	bobExpectedBalance := big.NewInt(0).Set(initialMint)
	assert.Equal(t, bobExpectedBalance, getBalance(testContext, bob))
	charlieExpectedBalance := big.NewInt(0).Sub(initialMint, big.NewInt(int64(rTxGasLimit*gasPrice)))
	assert.Equal(t, charlieExpectedBalance, getBalance(testContext, charlie))
	assert.Equal(t, initialMint, getBalance(testContext, david))

	charlieCurrentBalance := getBalance(testContext, charlie)
	testContext.CleanIntermediateTransactions(t)

	// 3.1 single signed transaction should not work
	userTx = vm.CreateTransaction(
		getNonce(testContext, alice),
		transferValue,
		alice,
		david,
		gasPrice,
		transferGas+guardianSigVerificationGas,
		make([]byte, 0))

	userTx.Version = txWithOptionVersion

	rtxData = integrationTests.PrepareRelayedTxDataV2(userTx)
	rTxGasLimit = 1 + transferGas + guardianSigVerificationGas + uint64(len(rtxData))
	rtx = vm.CreateTransaction(getNonce(testContext, charlie), big.NewInt(0), charlie, alice, gasPrice, rTxGasLimit, rtxData)
	returnCode, err = testContext.TxProcessor.ProcessTransaction(rtx)
	require.Nil(t, err)
	require.Equal(t, vmcommon.UserError, returnCode)
	intermediateTxs = testContext.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediateTxs))
	scr = intermediateTxs[0].(*smartContractResult.SmartContractResult)
	require.Equal(t, expectedReturnMessage, string(scr.ReturnMessage))
	// balance tests:
	//  alice: aliceCurrentBalance (no fee for the failed relayed transaction)
	//  bob: initialMint
	//  charlie: charlieCurrentBalance - rtxGasLimit * gasPrice
	//  david: davidCurrentBalance
	assert.Equal(t, aliceCurrentBalance, getBalance(testContext, alice))
	bobExpectedBalance = big.NewInt(0).Set(initialMint)
	assert.Equal(t, bobExpectedBalance, getBalance(testContext, bob))
	charlieExpectedBalance = big.NewInt(0).Sub(charlieCurrentBalance, big.NewInt(int64(rTxGasLimit*gasPrice)))
	assert.Equal(t, charlieExpectedBalance, getBalance(testContext, charlie))
	assert.Equal(t, initialMint, getBalance(testContext, david))
}
