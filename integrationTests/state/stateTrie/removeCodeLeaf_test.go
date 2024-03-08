package stateTrie

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	testStorage "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountsDB_MigrateCodeLeaf(t *testing.T) {
	t.Parallel()

	t.Run("account with migrated code leaf activation, should save code data to storage directly", func(t *testing.T) {
		t.Parallel()

		enableEpochs := config.EnableEpochs{
			MigrateCodeLeafEnableEpoch: 0,
		}

		shardID := uint32(1)
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, shardID)
		gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()

		storer := integrationTests.CreateMemUnit()
		trieStorage, _ := integrationTests.CreateTrieStorageManager(storer)

		testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(enableEpochs, shardCoordinator, storer, gasScheduleNotifier)
		require.Nil(t, err)
		defer testContext.Close()

		adr := integrationTests.CreateRandomAddress()
		adr[31] = 1 // set address for shard 1

		adb := testContext.Accounts

		stateMock, err := adb.LoadAccount(adr)
		require.Nil(t, err)

		err = adb.SaveAccount(stateMock)
		require.Nil(t, err)

		snapshotMod := adb.JournalLen()

		stateMock, err = adb.LoadAccount(adr)
		require.Nil(t, err)

		codeData := []byte("codeData")
		codeDataHash := integrationTests.TestHasher.Compute(string(codeData))
		stateMock.(state.UserAccountHandler).SetCodeHash(codeDataHash)
		stateMock.(state.UserAccountHandler).SetCode(codeData)

		err = adb.SaveAccount(stateMock)
		require.Nil(t, err)

		testMigrateCodeLeaf(t, testContext, stateMock.AddressBytes())

		stateMock, err = adb.LoadAccount(adr)
		require.Nil(t, err)

		accVersion := stateMock.(state.UserAccountHandler).GetVersion()
		require.Equal(t, uint8(core.WithoutCodeLeaf), accVersion)

		// check code data saved to storage
		val, err := trieStorage.Get(codeDataHash)
		require.Nil(t, err)
		require.Equal(t, codeData, val)

		err = adb.RevertToSnapshot(snapshotMod)
		require.Nil(t, err)

		stateMock, err = adb.LoadAccount(adr)
		require.Nil(t, err)

		accVersion = stateMock.(state.UserAccountHandler).GetVersion()
		require.Equal(t, uint8(core.NotSpecified), accVersion)

		// revert will not remove code from storage
		val, err = trieStorage.Get(codeDataHash)
		require.Nil(t, err)
		require.Equal(t, codeData, val)
	})

	t.Run("2 accounts with same code, migrate one account and reverted it", func(t *testing.T) {
		t.Parallel()

		enableEpochs := config.EnableEpochs{
			MigrateCodeLeafEnableEpoch: 0,
		}

		shardID := uint32(1)
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, shardID)
		gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()

		adr1 := integrationTests.CreateRandomAddress()
		adr1[31] = 1 // set address for shard 1

		adr2 := integrationTests.CreateRandomAddress()

		storer := integrationTests.CreateMemUnit()
		trieStorage, _ := integrationTests.CreateTrieStorageManager(storer)

		testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(enableEpochs, shardCoordinator, storer, gasScheduleNotifier)
		require.Nil(t, err)
		defer testContext.Close()

		adb := testContext.Accounts

		codeData := []byte("codeData")
		codeDataHash := integrationTests.TestHasher.Compute(string(codeData))

		stateMock1, err := adb.LoadAccount(adr1)
		require.Nil(t, err)
		stateMock1.(state.UserAccountHandler).SetCodeHash(codeDataHash)
		stateMock1.(state.UserAccountHandler).SetCode(codeData)

		err = adb.SaveAccount(stateMock1)
		require.Nil(t, err)

		stateMock2, err := adb.LoadAccount(adr2)
		require.Nil(t, err)
		stateMock2.(state.UserAccountHandler).SetCodeHash(codeDataHash)
		stateMock2.(state.UserAccountHandler).SetCode(codeData)

		err = adb.SaveAccount(stateMock2)
		require.Nil(t, err)

		snapshotMod := adb.JournalLen()

		testMigrateCodeLeaf(t, testContext, stateMock1.AddressBytes())

		stateMock1, err = adb.LoadAccount(adr1)
		require.Nil(t, err)

		accVersion := stateMock1.(state.UserAccountHandler).GetVersion()
		require.Equal(t, uint8(core.WithoutCodeLeaf), accVersion)

		accVersion2 := stateMock2.(state.UserAccountHandler).GetVersion()
		require.Equal(t, uint8(core.NotSpecified), accVersion2)

		// check code data saved to storage
		val, err := trieStorage.Get(codeDataHash)
		require.Nil(t, err)
		require.Equal(t, codeData, val)

		err = adb.RevertToSnapshot(snapshotMod)
		require.Nil(t, err)

		stateMock1, err = adb.LoadAccount(adr1)
		require.Nil(t, err)

		stateMock2, err = adb.LoadAccount(adr2)
		require.Nil(t, err)

		// revert will set old version back
		accVersion = stateMock1.(state.UserAccountHandler).GetVersion()
		require.Equal(t, uint8(core.NotSpecified), accVersion)

		// revert will not remove code from storage
		val, err = trieStorage.Get(codeDataHash)
		require.Nil(t, err)
		require.Equal(t, codeData, val)
	})
}

func TestAccountsDB_MigrateCodeLeaf_UpdatesAfterMigration(t *testing.T) {
	t.Parallel()

	t.Run("account with migrated code leaf, update code after migration", func(t *testing.T) {
		t.Parallel()

		enableEpochs := config.EnableEpochs{
			MigrateCodeLeafEnableEpoch: 0,
		}

		shardID := uint32(1)
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, shardID)
		gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()

		storer := integrationTests.CreateMemUnit()
		trieStorage, _ := integrationTests.CreateTrieStorageManager(storer)

		testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(enableEpochs, shardCoordinator, storer, gasScheduleNotifier)
		require.Nil(t, err)
		defer testContext.Close()

		adr := integrationTests.CreateRandomAddress()
		adr[31] = 1 // set address for shard 1

		adb := testContext.Accounts

		stateMock, err := adb.LoadAccount(adr)
		require.Nil(t, err)

		err = adb.SaveAccount(stateMock)
		require.Nil(t, err)

		stateMock, err = adb.LoadAccount(adr)
		require.Nil(t, err)

		codeData := []byte("codeData")
		codeDataHash := integrationTests.TestHasher.Compute(string(codeData))
		stateMock.(state.UserAccountHandler).SetCodeHash(codeDataHash)
		stateMock.(state.UserAccountHandler).SetCode(codeData)

		err = adb.SaveAccount(stateMock)
		require.Nil(t, err)

		testMigrateCodeLeaf(t, testContext, stateMock.AddressBytes())

		stateMock, err = adb.LoadAccount(adr)
		require.Nil(t, err)

		accVersion := stateMock.(state.UserAccountHandler).GetVersion()
		require.Equal(t, uint8(core.WithoutCodeLeaf), accVersion)

		val, err := trieStorage.Get(codeDataHash)
		require.Nil(t, err)
		require.Equal(t, codeData, val)

		snapshotMod := adb.JournalLen()

		newCodeData := []byte("newCodeData")
		newCodeDataHash := integrationTests.TestHasher.Compute(string(newCodeData))
		stateMock.(state.UserAccountHandler).SetCodeHash(newCodeDataHash)
		stateMock.(state.UserAccountHandler).SetCode(newCodeData)

		err = adb.SaveAccount(stateMock)
		require.Nil(t, err)

		val, err = trieStorage.Get(newCodeDataHash)
		require.Nil(t, err)
		require.Equal(t, newCodeData, val)

		stateMock, err = adb.LoadAccount(adr)
		require.Nil(t, err)
		require.Equal(t, newCodeDataHash, stateMock.(state.UserAccountHandler).GetCodeHash())

		// old code data not deleted from storage
		val, err = trieStorage.Get(codeDataHash)
		require.Nil(t, err)
		require.Equal(t, codeData, val)

		// revert should reference old code from storage
		err = adb.RevertToSnapshot(snapshotMod)
		require.Nil(t, err)

		stateMock, err = adb.LoadAccount(adr)
		require.Nil(t, err)
		require.Equal(t, codeDataHash, stateMock.(state.UserAccountHandler).GetCodeHash())

		// remove account will not delete code from storage
		err = adb.RemoveAccount(adr)
		require.Nil(t, err)

		val, err = trieStorage.Get(codeDataHash)
		require.Nil(t, err)
		require.Equal(t, codeData, val)

		val, err = trieStorage.Get(newCodeDataHash)
		require.Nil(t, err)
		require.Equal(t, newCodeData, val)
	})
}

func testMigrateCodeLeaf(
	t *testing.T,
	testContext *vm.VMTestContext,
	rcvAddr []byte,
) {
	testContext.CleanIntermediateTransactions(t)

	txData := []byte("MigrateCodeLeaf@aa@bb@00")
	sndAddr := []byte("12345678901234567890123456789112")
	scr := &smartContractResult.SmartContractResult{
		Value:    big.NewInt(0),
		RcvAddr:  rcvAddr,
		SndAddr:  sndAddr,
		Data:     txData,
		GasLimit: 10000,
		GasPrice: 10,
		CallType: 1,
	}
	returnCode, errProcess := testContext.ScProcessor.ProcessSmartContractResult(scr)
	require.Nil(t, errProcess)
	require.Equal(t, vmcommon.Ok, returnCode)

	intermediate := testContext.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediate))
}

func TestSnaphostState_WithMigratedCodeLeaf_ShouldWork(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)

	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return flag == common.MigrateCodeLeafFlag
		},
	}

	epochsNotifier := &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			header := &block.Header{Epoch: epoch}
			handler.EpochStartAction(header)
		},
	}

	mainStorer, persistersMap, err := stateMock.CreateTestingTriePruningStorer(testscommon.NewMultiShardsCoordinatorMock(1), epochsNotifier)
	args := testStorage.GetStorageManagerArgs()
	args.MainStorer = mainStorer
	args.Marshalizer = integrationTests.TestMarshalizer
	args.Hasher = integrationTests.TestHasher

	tsm, _ := trie.NewTrieStorageManager(args)
	adb, _ := integrationTests.CreateAccountsDBWithEnableEpochsHandler(0, tsm, enableEpochsHandler)
	_ = adb.SetSyncer(&mock.AccountsDBSyncerStub{})

	rootHash, err := addDataTriesForAccountsStartingWithIndex(0, 1, 1, adb)
	assert.Nil(t, err)

	codeData := []byte("codeData1")
	codeDataHash := integrationTests.TestHasher.Compute(string(codeData))

	address := integrationTests.CreateAccount(adb, 1, big.NewInt(100))
	acc, _ := adb.LoadAccount(address)
	userAcc, ok := acc.(state.UserAccountHandler)
	assert.True(t, ok)
	userAcc.SetCode(codeData)
	userAcc.SetCodeHash(codeDataHash)

	err = adb.SaveAccount(userAcc)
	require.Nil(t, err)

	userAcc.SetVersion(uint8(core.WithoutCodeLeaf))

	valSize := 1 << 21
	_ = addValuesToDataTrie(t, adb, userAcc, 3, valSize)

	rootHash, err = adb.Commit()
	require.Nil(t, err)

	lastestEpoch, err := tsm.GetLatestStorageEpoch()
	require.Nil(t, err)
	assert.Equal(t, uint32(1), lastestEpoch)

	persisterEpoch0 := persistersMap.GetPersister("Epoch_0/Shard_0/id")
	retCode, err := persisterEpoch0.Get(codeDataHash)
	require.Equal(t, codeData, retCode)
	require.Nil(t, err)

	// should not find code in epoch 1
	persisterEpoch1 := persistersMap.GetPersister("Epoch_1/Shard_0/id")
	retCode, err = persisterEpoch1.Get(codeDataHash)
	require.Nil(t, retCode)
	require.Error(t, err)

	adb.SnapshotState(rootHash, epoch)
	for adb.IsSnapshotInProgress() {
		time.Sleep(10 * time.Millisecond)
	}

	persisterEpoch0 = persistersMap.GetPersister("Epoch_0/Shard_0/id")
	retCode, err = persisterEpoch0.Get(codeDataHash)
	require.Equal(t, codeData, retCode)
	require.Nil(t, err)

	// should find code in epoch 1
	persisterEpoch1 = persistersMap.GetPersister("Epoch_1/Shard_0/id")
	retCode, err = persisterEpoch1.Get(codeDataHash)
	require.Equal(t, codeData, retCode)
	require.Nil(t, err)
}

func addValuesToDataTrie(t *testing.T, adb state.AccountsAdapter, acc state.UserAccountHandler, numVals int, valSize int) []byte {
	for i := 0; i < numVals; i++ {
		keyRandBytes := integrationTests.CreateRandomBytes(32)
		valRandBytes := integrationTests.CreateRandomBytes(valSize)
		_ = acc.SaveKeyValue(keyRandBytes, valRandBytes)
	}

	err := adb.SaveAccount(acc)
	assert.Nil(t, err)

	_, err = adb.Commit()
	assert.Nil(t, err)

	return acc.GetRootHash()
}
