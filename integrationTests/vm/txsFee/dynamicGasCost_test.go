//go:build !race

// TODO remove build condition above to allow -race -short, after Wasm VM fix

package txsFee

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/txsFee/utils"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/sharding"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestDynamicGasCostForDataTrieStorageLoad(t *testing.T) {
	enableEpochs := config.EnableEpochs{
		DynamicGasCostForDataTrieStorageLoadEnableEpoch: 0,
	}
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, 1)
	gasScheduleNotifier := vm.CreateMockGasScheduleNotifier()

	testContext, err := vm.CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(enableEpochs, shardCoordinator, integrationTests.CreateMemUnit(), gasScheduleNotifier)
	require.Nil(t, err)
	defer testContext.Close()

	scAddress, _ := utils.DoDeployNoChecks(t, testContext, "../wasm/testdata/trieStoreAndLoad/output/storage.wasm")
	acc := getAccount(t, testContext, scAddress)
	require.Nil(t, acc.DataTrie())

	sndNonce := uint64(0)
	sndAddr := []byte("12345678901234567890123456789112")
	senderBalance := big.NewInt(100000000)
	_, _ = vm.CreateAccount(testContext.Accounts, sndAddr, sndNonce, senderBalance)

	keys := insertInDataTrie(t, testContext, scAddress, 15)

	dataTrieInstance := getAccountDataTrie(t, testContext, scAddress)
	trieKeysDepth := getTrieDepthForKeys(t, dataTrieInstance, keys)

	initCost := 35
	apiCallsCost := 3
	loadValCost := 32
	wasmOpsCost := 14

	contractCode := wasm.GetSCCode("../wasm/testdata/trieStoreAndLoad/output/storage.wasm")
	latestGasSchedule := gasScheduleNotifier.LatestGasSchedule()
	aotPrepare := latestGasSchedule[common.BaseOperationCost]["AoTPreparePerByte"] * uint64(len(contractCode)) / 2

	gasCost := int64(initCost) + int64(apiCallsCost+loadValCost+wasmOpsCost) + int64(aotPrepare)

	for i, key := range keys {
		trieLoadCost := getExpectedConsumedGasForTrieLoad(testContext, int64(trieKeysDepth[i]))

		expectedGasCost := trieLoadCost + gasCost

		fmt.Println("trie level", trieKeysDepth[i])
		gasLimit := uint64(trieLoadCost) + 10000
		testGasConsumedForDataTrieLoad(t, testContext, sndNonce, key, sndAddr, scAddress, gasPrice, gasLimit, expectedGasCost)
		sndNonce++
	}
}

func insertInDataTrie(
	t *testing.T,
	testContext *vm.VMTestContext,
	accAddr []byte,
	maxTrieLevel int,
) [][]byte {
	acc := getAccount(t, testContext, accAddr)
	keys := integrationTests.GenerateTrieKeysForMaxLevel(maxTrieLevel, 2)
	for _, key := range keys {
		err := acc.SaveKeyValue(key, key)
		require.Nil(t, err)
	}

	err := testContext.Accounts.SaveAccount(acc)
	require.Nil(t, err)

	return keys
}

func testGasConsumedForDataTrieLoad(
	t *testing.T,
	testContext *vm.VMTestContext,
	nonce uint64,
	key []byte,
	sndAddr []byte,
	scAddress []byte,
	gasPrice uint64,
	gasLimit uint64,
	expectedGasConsumed int64,
) {
	testContext.CleanIntermediateTransactions(t)

	txData := []byte("trieLoad@" + hex.EncodeToString(key) + "@aa@bb@00")
	scr := &smartContractResult.SmartContractResult{
		Nonce:    nonce,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  sndAddr,
		Data:     txData,
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		CallType: 1,
	}
	returnCode, errProcess := testContext.ScProcessor.ProcessSmartContractResult(scr)
	require.Nil(t, errProcess)
	require.Equal(t, vmcommon.Ok, returnCode)

	intermediate := testContext.GetIntermediateTransactions(t)
	require.Equal(t, 1, len(intermediate))

	gasConsumed := gasLimit - intermediate[0].GetGasLimit()
	fmt.Println("gas consumed", gasConsumed)
	fmt.Println("expected gas consumed", expectedGasConsumed)
	require.Equal(t, uint64(expectedGasConsumed), gasConsumed)

}

func getTrieDepthForKeys(t *testing.T, tr common.Trie, keys [][]byte) []uint32 {
	trieLevels := make([]uint32, len(keys))
	for i, key := range keys {
		tld, err := tr.Get(integrationTests.TestHasher.Compute(string(key)))
		require.Nil(t, err)
		trieLevels[i] = tld.Depth()
	}

	return trieLevels
}

type funcParams struct {
	Quadratic int64
	Linear    int64
	Constant  int64
}

func getExpectedConsumedGasForTrieLoad(testContext *vm.VMTestContext, trieLevel int64) int64 {
	gasSchedule := testContext.GasSchedule.LatestGasSchedule()
	f := getFunc(gasSchedule)

	return trieLevel*trieLevel*f.Quadratic + trieLevel*f.Linear + f.Constant
}

func getFunc(gasSchedule map[string]map[string]uint64) funcParams {
	dynamicStorageLoad := gasSchedule["DynamicStorageLoad"]
	return funcParams{
		Quadratic: getSignedCoefficient(dynamicStorageLoad["QuadraticCoefficient"], dynamicStorageLoad["SignOfQuadratic"]),
		Linear:    getSignedCoefficient(dynamicStorageLoad["LinearCoefficient"], dynamicStorageLoad["SignOfLinear"]),
		Constant:  getSignedCoefficient(dynamicStorageLoad["ConstantCoefficient"], dynamicStorageLoad["SignOfConstant"]),
	}
}

func getSignedCoefficient(coefficient uint64, sign uint64) int64 {
	isNegativeNumber := uint64(1)
	if sign == isNegativeNumber {
		return int64(coefficient) * -1
	}

	return int64(coefficient)
}
