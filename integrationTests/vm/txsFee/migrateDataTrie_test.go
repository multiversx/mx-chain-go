//go:build !race
// +build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/sharding"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

type statsCollector interface {
	GetTrieStats(address string, rootHash []byte) (common.TrieStatisticsHandler, error)
}

type dataTrie interface {
	UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion) error
}

func TestMigrateDataTrieBuiltInFunc(t *testing.T) {
	t.Parallel()

	enableEpochs := config.EnableEpochs{
		AutoBalanceDataTriesEnableEpoch: 0,
	}
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, 1)
	gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()
	trieLoadPerNode := uint64(20000)
	trieStorePerNode := uint64(50000)
	gasScheduleNotifier.GasSchedule[core.BuiltInCostString]["TrieLoadPerNode"] = trieLoadPerNode
	gasScheduleNotifier.GasSchedule[core.BuiltInCostString]["TrieStorePerNode"] = trieStorePerNode
	sndAddr := []byte("12345678901234567890123456789111")

	t.Run("deterministic trie", func(t *testing.T) {
		t.Parallel()

		testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(enableEpochs, shardCoordinator, integrationTests.CreateMemUnit(), gasScheduleNotifier)
		require.Nil(t, err)
		defer testContext.Close()

		senderBalance := big.NewInt(100000000)
		sndNonce := uint64(0)
		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, sndNonce, senderBalance)

		numDataTrieLeaves := 10
		keyGenerator := func(i int) []byte {
			return []byte(strconv.Itoa(i))
		}
		rootHash, _ := generateDataTrie(t, testContext, sndAddr, numDataTrieLeaves, keyGenerator)

		dtr := getAccountDataTrie(t, testContext, sndAddr)
		stats, ok := dtr.(statsCollector)
		require.True(t, ok)

		dts, err := stats.GetTrieStats("", rootHash)
		require.Nil(t, err)
		require.Equal(t, uint64(numDataTrieLeaves-1), dts.GetLeavesMigrationStats()[core.NotSpecified])
		require.Equal(t, uint64(1), dts.GetLeavesMigrationStats()[core.AutoBalanceEnabled])

		// migrate first 2 leaves, return when loading the third leaf (not enough gas for the migration)
		// 5 loads + 2 stores = 200k gas
		gasLimit := uint64(220000)
		migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.Ok)
		testGasConsumed(t, testContext, gasLimit, 200000)

		// migrate 2 leaves, 4 loads + 2 stores = 180k gas
		gasLimit = uint64(200000)
		migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.Ok)
		testGasConsumed(t, testContext, gasLimit, 180000)

		// do not start the migration process, not enough gas for at least one migration
		gasLimit = uint64(50000)
		migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.UserError)
		testGasConsumed(t, testContext, gasLimit, 50000)

		// return after loading a branch node, not enough gas for the migration
		gasLimit = uint64(70000)
		migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.Ok)
		testGasConsumed(t, testContext, gasLimit, 60000)

		// migrate 2 leaves, 5 loads + 2 stores = 200k gas
		gasLimit = uint64(200000)
		migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.Ok)
		testGasConsumed(t, testContext, gasLimit, 200000)

		// migrate 2 leaves, 3 loads + 2 stores = 160k gas
		gasLimit = uint64(200000)
		migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.Ok)
		testGasConsumed(t, testContext, gasLimit, 160000)

		// migrate 1 leaf, 2 loads + 1 store = 90k gas
		gasLimit = uint64(200000)
		migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.Ok)
		testGasConsumed(t, testContext, gasLimit, 90000)

		// no leaf left to migrate, 1 load = 20k gas
		gasLimit = uint64(200000)
		migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.Ok)
		testGasConsumed(t, testContext, gasLimit, 20000)

		err = dtr.Commit()
		require.Nil(t, err)

		rootHash, err = dtr.RootHash()
		require.Nil(t, err)

		dts, err = stats.GetTrieStats("", rootHash)
		require.Nil(t, err)
		require.Equal(t, uint64(0), dts.GetLeavesMigrationStats()[core.NotSpecified])
		require.Equal(t, uint64(numDataTrieLeaves), dts.GetLeavesMigrationStats()[core.AutoBalanceEnabled])
	})

	t.Run("random trie - all leaves are migrated in multiple transactions", func(t *testing.T) {
		t.Parallel()

		testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(enableEpochs, shardCoordinator, integrationTests.CreateMemUnit(), gasScheduleNotifier)
		require.Nil(t, err)
		defer testContext.Close()

		sndNonce := uint64(0)
		senderBalance := big.NewInt(math.MaxInt64)
		_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, sndNonce, senderBalance)

		numLeaves := 10000
		keyGenerator := func(i int) []byte {
			return integrationTests.GenerateRandomSlice(32)
		}
		rootHash, allKeys := generateDataTrie(t, testContext, sndAddr, numLeaves, keyGenerator)
		nonMigratedKeys := allKeys[1:]

		acc := getAccount(t, testContext, sndAddr)
		dtr, ok := acc.DataTrie().(common.Trie)
		require.True(t, ok)
		stats, ok := dtr.(statsCollector)
		require.True(t, ok)

		dts, err := stats.GetTrieStats("", rootHash)
		require.Nil(t, err)

		numNonMigratedLeaves := dts.GetLeavesMigrationStats()[core.NotSpecified]
		numMigratedLeaves := dts.GetLeavesMigrationStats()[core.AutoBalanceEnabled]
		require.Equal(t, uint64(numLeaves-1), numNonMigratedLeaves)
		require.Equal(t, uint64(1), numMigratedLeaves)

		gasLimit := uint64(100_000_000)

		numMigrateDataTrieCalls := 0
		maxExpectedNumCalls := 15
		for numNonMigratedLeaves > 0 {
			migrateDataTrie(t, testContext, sndAddr, gasPrice, gasLimit, vmcommon.Ok)
			numMigrateDataTrieCalls++

			err = dtr.Commit()
			require.Nil(t, err)

			rootHash, err = dtr.RootHash()
			require.Nil(t, err)

			dts, err = stats.GetTrieStats("", rootHash)
			require.Nil(t, err)

			require.True(t, dts.GetLeavesMigrationStats()[core.NotSpecified] < numNonMigratedLeaves)
			require.True(t, dts.GetLeavesMigrationStats()[core.AutoBalanceEnabled] > numMigratedLeaves)

			numNonMigratedLeaves = dts.GetLeavesMigrationStats()[core.NotSpecified]
			numMigratedLeaves = dts.GetLeavesMigrationStats()[core.AutoBalanceEnabled]
		}

		require.True(t, numMigrateDataTrieCalls < maxExpectedNumCalls)

		require.Equal(t, uint64(0), dts.GetLeavesMigrationStats()[core.NotSpecified])
		require.Equal(t, uint64(numLeaves), dts.GetLeavesMigrationStats()[core.AutoBalanceEnabled])

		err = testContext.Accounts.SaveAccount(acc)
		require.Nil(t, err)

		acc = getAccount(t, testContext, sndAddr)

		for _, key := range nonMigratedKeys {
			val, _, errRetrieve := acc.RetrieveValue(key)
			require.Nil(t, errRetrieve)
			require.Equal(t, key, val)
		}
	})
}

func generateDataTrie(
	t *testing.T,
	testContext *vm.VMTestContext,
	accAddr []byte,
	numLeaves int,
	keyGenerator func(i int) []byte,
) ([]byte, [][]byte) {
	acc := getAccount(t, testContext, accAddr)
	keys := make([][]byte, numLeaves)

	firstKey := initDataTrie(t, testContext, acc)
	keys[0] = firstKey

	dataTr := getAccountDataTrie(t, testContext, accAddr)
	tr, ok := dataTr.(dataTrie)
	require.True(t, ok)

	for i := 1; i < numLeaves; i++ {
		key := keyGenerator(i)
		err := tr.UpdateWithVersion(key, key, core.NotSpecified)
		require.Nil(t, err)

		keys[i] = key
	}

	rootHash := saveAccount(t, testContext, dataTr, acc)

	return rootHash, keys
}

func initDataTrie(
	t *testing.T,
	testContext *vm.VMTestContext,
	acc common.UserAccountHandler,
) []byte {
	key := []byte("initDataTrieKey")
	err := acc.SaveKeyValue(key, key)
	require.Nil(t, err)
	err = testContext.Accounts.SaveAccount(acc)
	require.Nil(t, err)

	return key
}

func saveAccount(
	t *testing.T,
	testContext *vm.VMTestContext,
	dataTr common.Trie,
	acc common.UserAccountHandler,
) []byte {
	rootHash, _ := dataTr.RootHash()
	acc.SetRootHash(rootHash)

	err := testContext.Accounts.SaveAccount(acc)
	require.Nil(t, err)

	_, err = testContext.Accounts.Commit()
	require.Nil(t, err)

	return rootHash
}

func migrateDataTrie(
	t *testing.T,
	testContext *vm.VMTestContext,
	sndAddr []byte,
	gasPrice uint64,
	gasLimit uint64,
	expectedReturnCode vmcommon.ReturnCode,
) {
	testContext.CleanIntermediateTransactions(t)

	gasLocked := "00" //use all available gas
	txData := core.BuiltInFunctionMigrateDataTrie + "@" + gasLocked

	scr := &smartContractResult.SmartContractResult{
		Value:    big.NewInt(0),
		RcvAddr:  sndAddr,
		SndAddr:  sndAddr,
		Data:     []byte(txData),
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		CallType: 1,
	}
	returnCode, errProcess := testContext.ScProcessor.ProcessSmartContractResult(scr)
	if expectedReturnCode == vmcommon.Ok {
		require.Nil(t, errProcess)
	} else {
		require.NotNil(t, errProcess)
	}

	require.Equal(t, expectedReturnCode, returnCode)
}

func testGasConsumed(
	t *testing.T,
	testContext *vm.VMTestContext,
	gasLimit uint64,
	expectedGasConsumed uint64,
) {
	intermediate := testContext.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediate))

	gasConsumed := gasLimit - intermediate[0].GetGasLimit()
	fmt.Println("gas consumed", gasConsumed)
	fmt.Println("expected gas consumed", expectedGasConsumed)
	require.Equal(t, expectedGasConsumed, gasConsumed)
}
