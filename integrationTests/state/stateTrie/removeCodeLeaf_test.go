package stateTrie

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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
