package arwenVM

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func Benchmark_VmDeployWithFibbonacciAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/fib_arwen/output/fib_arwen.wasm", 32, "_main", nil, b.N, nil)
}

func Benchmark_VmDeployWithCPUCalculateAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/cpucalculate_arwen/cpucalculate_arwen.wasm", 8000, "_main", nil, b.N, nil)
}

func Benchmark_VmDeployWithStringConcatAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/stringconcat_arwen/stringconcat_arwen.wasm", 10000, "_main", nil, b.N, nil)
}

func Benchmark_TestStore100(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/storage100/output/storage100.wasm", 0, "store100", nil, b.N, nil)
}

func Benchmark_TestStorageBigIntNew(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/cApiTest.wasm", 0, "bigIntNewTest", nil, b.N, nil)
}

func Benchmark_TestBigIntGetUnSignedBytes(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/cApiTest.wasm", 0, "bigIntGetUnsignedBytesTest", nil, b.N, nil)
}

func Benchmark_TestBigIntAdd(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/cApiTest.wasm", 0, "bigIntAddTest", nil, b.N, nil)
}

func Benchmark_TestBigIntMul(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/cApiTest.wasm", 0, "bigIntMulTest", nil, b.N, nil)
}

func Benchmark_TestBigIntShr(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/cApiTest.wasm", 0, "bigIntShrTest", nil, b.N, nil)
}

func Benchmark_TestStorageRust(b *testing.B) {
	buff := make([]byte, 1000)
	_, _ = rand.Read(buff)
	arguments := [][]byte{buff, big.NewInt(1000).Bytes()}
	runWASMVMBenchmark(b, "../testdata/misc/str-repeat.wasm", 0, "repeat", arguments, b.N, nil)
}

func TestGasModel(t *testing.T) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../gasSchedule.toml")

	totalOp := uint64(0)
	for _, opCodeClass := range gasSchedule {
		for _, opCode := range opCodeClass {
			totalOp += opCode
		}
	}
	fmt.Println("gasSchedule: " + big.NewInt(int64(totalOp)).String())
	fmt.Println("ERC20 BIGINT")
	deployAndExecuteERC20WithBigInt(t, 1, 100, gasSchedule, "../testdata/erc20-c-03/wrc20_arwen.wasm", "transferToken", false, false)
}

func runWASMVMBenchmark(
	tb testing.TB,
	fileSC string,
	testingValue uint64,
	function string,
	arguments [][]byte,
	numRun int,
	gasSchedule map[string]map[string]uint64,
) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(0xfffffffffffffff)
	ownerBalance.Mul(ownerBalance, big.NewInt(0xffffffff))
	gasPrice := uint64(1)
	gasLimit := uint64(0xffffffffffffffff)

	scCode := arwen.GetSCCode(fileSC)

	tx := &transaction.Transaction{
		Nonce:     ownerNonce,
		Value:     big.NewInt(0),
		RcvAddr:   vm.CreateEmptyAddress(),
		SndAddr:   ownerAddressBytes,
		GasPrice:  gasPrice,
		GasLimit:  gasLimit,
		Data:      []byte(arwen.CreateDeployTxData(scCode)),
		Signature: nil,
	}

	testContext := vm.CreateTxProcessorArwenVMWithGasSchedule(ownerNonce, ownerAddressBytes, ownerBalance, gasSchedule, false, false)
	defer testContext.Close()

	scAddress, _ := testContext.BlockchainHook.NewAddress(ownerAddressBytes, ownerNonce, factory.ArwenVirtualMachine)

	returnCode, err := testContext.TxProcessor.ProcessTransaction(tx)
	require.Nil(tb, err)
	require.Equal(tb, returnCode, vmcommon.Ok)

	_, err = testContext.Accounts.Commit()
	require.Nil(tb, err)

	alice := []byte("12345678901234567890123456789111")
	aliceNonce := uint64(0)
	_, _ = vm.CreateAccount(testContext.Accounts, alice, aliceNonce, big.NewInt(10000000000))

	txData := function
	for _, arg := range arguments {
		txData += "@" + hex.EncodeToString(arg)
	}
	tx = &transaction.Transaction{
		Nonce:     aliceNonce,
		Value:     new(big.Int).Set(big.NewInt(0).SetUint64(testingValue)),
		RcvAddr:   scAddress,
		SndAddr:   alice,
		GasPrice:  0,
		GasLimit:  gasLimit,
		Data:      []byte(txData),
		Signature: nil,
	}

	startTime := time.Now()
	for i := 0; i < numRun; i++ {
		tx.Nonce = aliceNonce

		_, _ = testContext.TxProcessor.ProcessTransaction(tx)

		aliceNonce++
	}
	log.Info("elapsed time to process", "time", time.Since(startTime), "numRun", numRun)
}
