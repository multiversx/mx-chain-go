package wasmvm

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/stretchr/testify/require"
)

func Benchmark_VmDeployWithFibbonacciAndExecute(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/misc/fib_wasm/output/fib_wasm.wasm", 32, "_main", nil, b.N, nil)
}

func Benchmark_searchingForPanic(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}
	for i := 0; i < 10; i++ {
		runWASMVMBenchmark(b, "../testdata/misc/fib_wasm/output/fib_wasm.wasm", 100, "_main", nil, b.N, nil)
	}
}

func Test_searchingForPanic(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	for i := 0; i < 10; i++ {
		runWASMVMBenchmark(t, "../testdata/misc/fib_wasm/output/fib_wasm.wasm", 100, "_main", nil, 1, nil)
	}
}

func Benchmark_VmDeployWithBadContractAndExecute(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV4.toml")

	result, err := RunTest("../testdata/misc/bad.wasm", 0, "bigLoop", nil, b.N, gasSchedule, 1500000000)
	require.Nil(b, err)

	log.Info("test completed",
		"function", result.FunctionName,
		"consumed gas", result.GasUsed,
		"time took", result.ExecutionTimeSpan,
		"numRun", b.N,
	)
}

func Benchmark_VmDeployWithBadContractAndExecute2(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV4.toml")

	arg, _ := hex.DecodeString("012c")
	result, err := RunTest("../testdata/misc/bad.wasm", 0, "stressBigInts", [][]byte{arg}, b.N, gasSchedule, 1500000000)
	require.Nil(b, err)

	log.Info("test completed",
		"function", result.FunctionName,
		"consumed gas", result.GasUsed,
		"time took", result.ExecutionTimeSpan,
		"numRun", b.N,
	)
}

func Benchmark_VmDeployWithCPUCalculateAndExecute(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/misc/cpucalculate_wasm/output/cpucalculate.wasm", 8000, "cpuCalculate", nil, b.N, nil)
}

func Benchmark_VmDeployWithStringConcatAndExecute(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/misc/stringconcat_wasm/stringconcat_wasm.wasm", 10000, "_main", nil, b.N, nil)
}

func Benchmark_TestStore100(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/storage100/output/storage100.wasm", 0, "store100", nil, b.N, nil)
}

func Benchmark_TestStorageBigIntNew(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntNewTest", nil, b.N, nil)
}

func Benchmark_TestBigIntGetUnSignedBytes(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntGetUnsignedBytesTest", nil, b.N, nil)
}

func Benchmark_TestBigIntAdd(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntAddTest", nil, b.N, nil)
}

func Benchmark_TestBigIntMul(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntMulTest", nil, b.N, nil)
}

func Benchmark_TestBigIntMul25(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntMul25Test", nil, b.N, nil)
}

func Benchmark_TestBigIntMul32(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntMul32Test", nil, b.N, nil)
}

func Benchmark_TestBigIntTDiv(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntTDivTest", nil, b.N, nil)
}

func Benchmark_TestBigIntTMod(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntTModTest", nil, b.N, nil)
}

func Benchmark_TestBigIntEDiv(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntEDivTest", nil, b.N, nil)
}

func Benchmark_TestBigIntEMod(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntEModTest", nil, b.N, nil)
}

func Benchmark_TestBigIntShr(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntShrTest", nil, b.N, nil)
}

func Benchmark_TestBigIntSetup(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntInitSetup", nil, b.N, nil)
}

func Benchmark_TestCryptoSHA256(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "sha256Test", nil, b.N, nil)
}

func Benchmark_TestCryptoKeccak256(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "keccak256Test", nil, b.N, nil)
}

func Benchmark_TestCryptoRipMed160(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "ripemd160Test", nil, b.N, nil)
}

func Benchmark_TestCryptoBLS(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifyBLSTest", nil, b.N, nil)
}

func Benchmark_TestCryptoVerifyED25519(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifyEd25519Test", nil, b.N, nil)
}

func Benchmark_TestCryptoSecp256k1UnCompressed(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifySecp256k1UncompressedKeyTest", nil, b.N, nil)
}

func Benchmark_TestCryptoSecp256k1Compressed(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifySecp256k1CompressedKeyTest", nil, b.N, nil)
}

func Benchmark_TestEllipticCurveInitialVariablesAndCalls(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "initialVariablesAndCallsTest", nil, b.N, nil)
}

// elliptic curves

func Benchmark_TestEllipticCurve(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	testEllipticCurve(b, "p224Add")
	testEllipticCurve(b, "p256Add")
	testEllipticCurve(b, "p384Add")
	testEllipticCurve(b, "p521Add")
	testEllipticCurve(b, "p224Double")
	testEllipticCurve(b, "p256Double")
	testEllipticCurve(b, "p384Double")
	testEllipticCurve(b, "p521Double")
	testEllipticCurve(b, "p224IsOnCurve")
	testEllipticCurve(b, "p256IsOnCurve")
	testEllipticCurve(b, "p384IsOnCurve")
	testEllipticCurve(b, "p521IsOnCurve")
	testEllipticCurve(b, "p224Marshal")
	testEllipticCurve(b, "p256Marshal")
	testEllipticCurve(b, "p384Marshal")
	testEllipticCurve(b, "p521Marshal")
	testEllipticCurve(b, "p224Unmarshal")
	testEllipticCurve(b, "p256Unmarshal")
	testEllipticCurve(b, "p384Unmarshal")
	testEllipticCurve(b, "p521Unmarshal")
	testEllipticCurve(b, "p224MarshalCompressed")
	testEllipticCurve(b, "p256MarshalCompressed")
	testEllipticCurve(b, "p384MarshalCompressed")
	testEllipticCurve(b, "p521MarshalCompressed")
	testEllipticCurve(b, "p224UnmarshalCompressed")
	testEllipticCurve(b, "p256UnmarshalCompressed")
	testEllipticCurve(b, "p384UnmarshalCompressed")
	testEllipticCurve(b, "p521UnmarshalCompressed")
	testEllipticCurve(b, "p224GenerateKey")
	testEllipticCurve(b, "p256GenerateKey")
	testEllipticCurve(b, "p384GenerateKey")
	testEllipticCurve(b, "p521GenerateKey")
}

func Benchmark_TestEllipticCurveScalarMultP224(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224ScalarMultEcTest", getNumberOfRepsAndScalarLengthArgs(10), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP256(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256ScalarMultEcTest", getNumberOfRepsAndScalarLengthArgs(10), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP384(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384ScalarMultEcTest", getNumberOfRepsAndScalarLengthArgs(10), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP521(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521ScalarMultEcTest", getNumberOfRepsAndScalarLengthArgs(10), b.N, gasSchedule)
}

func testEllipticCurve(b *testing.B, function string) {
	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, function+"EcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestCryptoDoNothing(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "doNothing", nil, b.N, nil)
}

func Benchmark_TestStorageRust(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)
	buff := make([]byte, 100)
	_, _ = rand.Read(buff)
	arguments := [][]byte{buff, big.NewInt(100).Bytes()}
	runWASMVMBenchmark(b, "../testdata/misc/str-repeat.wasm", 0, "repeat", arguments, b.N, gasSchedule)
}

func TestGasModel(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	gasSchedule, _ := common.LoadGasScheduleConfig(integrationTests.GasSchedulePath)

	totalOp := uint64(0)
	for _, opCodeClass := range gasSchedule {
		for _, opCode := range opCodeClass {
			totalOp += opCode
		}
	}

	fmt.Println("gasSchedule: " + big.NewInt(int64(totalOp)).String())
	fmt.Println("ERC20 BIGINT")
	durations, err := DeployAndExecuteERC20WithBigInt(1, 100, gasSchedule, "../testdata/erc20-c-03/wrc20_wasm.wasm", "transferToken")
	require.Nil(t, err)
	displayBenchmarksResults(durations)

	runWASMVMBenchmark(t, "../testdata/misc/cpucalculate_wasm/output/cpucalculate.wasm", 8000, "cpuCalculate", nil, 2, gasSchedule)

	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntNewTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntGetUnsignedBytesTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntAddTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntMulTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntMul25Test", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntTModTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntTDivTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntEDivTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntEModTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntShrTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntInitSetup", nil, 2, gasSchedule)

	runWASMVMBenchmark(t, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "sha256Test", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "keccak256Test", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "ripemd160Test", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifyBLSTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifyEd25519Test", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifySecp256k1UncompressedKeyTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifySecp256k1CompressedKeyTest", nil, 2, gasSchedule)
	runWASMVMBenchmark(t, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "doNothing", nil, 2, gasSchedule)
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
	result, err := RunTest(fileSC, testingValue, function, arguments, numRun, gasSchedule, 0)
	require.Nil(tb, err)

	log.Info("test completed",
		"function", result.FunctionName,
		"consumed gas", result.GasUsed,
		"time took", result.ExecutionTimeSpan,
		"numRun", numRun,
	)
}

func getNumberOfRepsArgument() [][]byte {
	return [][]byte{{0x27, 0x10}} // = 10000
}

func getNumberOfRepsAndScalarLengthArgs(scalarLength int) [][]byte {
	return [][]byte{{0x27, 0x10}, {byte(scalarLength)}}
}
