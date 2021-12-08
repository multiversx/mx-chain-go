// +build !race

// TODO remove build condition above to allow -race -short, after Arwen fix

package arwenvm

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/stretchr/testify/require"
)

const numberOfArguments = 10001
const bigFloatMaxExponent = 65025
const bigFloatMinExponent = -65025

func Benchmark_VmDeployWithFibbonacciAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/fib_arwen/output/fib_arwen.wasm", 32, "_main", nil, b.N, nil)
}

func Benchmark_VmDeployWithBadContractAndExecute(b *testing.B) {
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
	runWASMVMBenchmark(b, "../testdata/misc/cpucalculate_arwen/output/cpucalculate.wasm", 8000, "cpuCalculate", nil, b.N, nil)
}

func Benchmark_VmDeployWithStringConcatAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/stringconcat_arwen/stringconcat_arwen.wasm", 10000, "_main", nil, b.N, nil)
}

func Benchmark_TestStore100(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/storage100/output/storage100.wasm", 0, "store100", nil, b.N, nil)
}

// MANAGED BUFFERS

func Benchmark_TestBufferGetArgument(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBufferGetArgument", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBufferFinish(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBufferFinish", bigFloatArguments, b.N, gasSchedule)
}

// BIG FLOAT

// 1) big floats can be encoded using .GobEncode() and decoded .GobDecode()
// and so this is the way they are casted to/from managed buffers
// 2) an encoded float looks like:
// [1, 10, 0, 0, 0, 53, 0, 0, 0, 25, 139, 30, 172, 233, 27, 5, 64, 0] which would be equal to =18234713.8211371
//
// to change exponent value, modify bytes [6:10]:           .  .  .  .
// to change mantissa value, modify bytes [10:]:                         .   .    .    .   .   .  .   .
// bigFloatArguments[1] = []byte{1, 10, 0, 0, 0, 53, 0, 0, 0, 25, 139, 30, 172, 233, 27, 5, 64, 0}

func Benchmark_TestBufferToBigFloat(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBufferToBigFloat", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBufferFromBigFloat(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBufferFromBigFloat", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatNewFromParts(b *testing.B) {
	exponentArgument := []byte{255, 255, 223, 246} // this has to be negative
	integralPartArgument := []byte{255, 255, 255, 255}
	fractionalPartArgument := []byte{255, 255, 255, 255}
	bigFloatArguments := make([][]byte, 4)
	bigFloatArguments[0] = []byte{0, 0, 27, 10}
	bigFloatArguments[1] = exponentArgument
	bigFloatArguments[2] = integralPartArgument
	bigFloatArguments[3] = fractionalPartArgument
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatNewFromParts", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatNewFromFrac(b *testing.B) {
	numeratorArgument := []byte{0, 76, 0, 23}
	denominatorArgument := []byte{247, 22, 152, 23}
	bigFloatArguments := make([][]byte, 3)
	bigFloatArguments[0] = []byte{0, 0, 27, 10}
	bigFloatArguments[1] = numeratorArgument
	bigFloatArguments[2] = denominatorArgument
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatNewFromFrac", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatNewFromSci(b *testing.B) {
	exponentArgument := []byte{255, 255, 255, 254} // this has to be negative
	significantArgument := []byte{0, 0, 0, 0}
	bigFloatArguments := make([][]byte, 3)
	bigFloatArguments[0] = []byte{0, 0, 27, 10}
	bigFloatArguments[1] = exponentArgument
	bigFloatArguments[2] = significantArgument
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatNewFromSci", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatAdd(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatAdd", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatSub(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatSub", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatSubGO(b *testing.B) {
	bigFloatArguments, _ := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	value1 := new(big.Float)
	value2 := new(big.Float)
	result := new(big.Float)
	b.ResetTimer()
	for i := 0; i < numberOfArguments-1; i++ {
		_ = value1.GobDecode(bigFloatArguments[i+1])
		_ = value2.GobDecode(bigFloatArguments[i+2])
		fmt.Println("VALUE1:", bigFloatArguments[i+1], "VALUE2:", bigFloatArguments[i+2])
		result.Sub(value1, value2)
		fmt.Println(result)
	}
}

func Benchmark_TestBigFloatMul(b *testing.B) {
	bigFloatArguments := getBigFloatMulArguments()
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatMul", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatDiv(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatDiv", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatDivGO(b *testing.B) {
	bigFloatArguments, _ := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	value1 := new(big.Float)
	value2 := new(big.Float)
	result := new(big.Float)
	b.ResetTimer()
	for i := 0; i < numberOfArguments-1; i++ {
		_ = value1.GobDecode(bigFloatArguments[i+1])
		_ = value2.GobDecode(bigFloatArguments[i+2])
		result.Quo(value1, value2)
	}
}

func Benchmark_TestBigFloatMulGO(b *testing.B) {
	bigFloatArguments, _ := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	value1 := new(big.Float)
	value2 := new(big.Float)
	result := new(big.Float)
	b.ResetTimer()
	for i := 0; i < numberOfArguments-1; i++ {
		_ = value1.GobDecode(bigFloatArguments[i+1])
		_ = value2.GobDecode(bigFloatArguments[i+2])
		result.Quo(value1, value2)
	}
}

func Benchmark_TestBigFloatTruncate(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatTruncate", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatAbs(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatAbs", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatNeg(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatNeg", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatCmp(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatCmp", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatSign(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatSign", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatClone(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatClone", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatSqrt(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatSqrt", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatPow(b *testing.B) {
	exponentArgument := []byte{0, 0, 0, 10}
	bigFloatArguments := make([][]byte, 3)
	bigFloatArguments[0] = []byte{0, 0, 27, 10}
	bigFloatArguments[1] = exponentArgument
	floatValue := big.NewFloat(2.64)
	encodedFloat, _ := floatValue.GobEncode()
	bigFloatArguments[2] = encodedFloat
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatPow", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatFloor(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatFloor", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatCeil(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatCeil", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatIsInt(b *testing.B) {
	bigFloatArguments, gasSchedule := getRandomBigFloatArgumentsAndGasSchedule()
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatIsInt", bigFloatArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatSetInt64(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatSetInt64", nil, b.N, nil)
}

func Benchmark_TestBigFloatSetBigInt(b *testing.B) {
	bigIntArguments, gasSchedule := getRandomBigIntArgumentsAndGasSchedule(1000)
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatSetBigInt", bigIntArguments, b.N, gasSchedule)
}

func Benchmark_TestBigIntGetUnsignedArgument(b *testing.B) {
	bigIntArguments, gasSchedule := getRandomBigIntArgumentsAndGasSchedule(100000)
	b.ResetTimer()
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigIntGetUnsignedArgument", bigIntArguments, b.N, gasSchedule)
}

func Benchmark_TestBigFloatGetConstPi(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatGetConstPi", nil, b.N, nil)
}

func Benchmark_TestBigFloatGetConstE(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/benchmark-big-floats/output/benchmark-big-floats.wasm", 0, "MeasureBigFloatGetConstE", nil, b.N, nil)
}

// BIG INT

func Benchmark_TestStorageBigIntNew(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntNewTest", nil, b.N, nil)
}

func Benchmark_TestBigIntGetUnSignedBytes(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntGetUnsignedBytesTest", nil, b.N, nil)
}

func Benchmark_TestBigIntAdd(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntAddTest", nil, b.N, nil)
}

func Benchmark_TestBigIntMul(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntMulTest", nil, b.N, nil)
}

func Benchmark_TestBigIntMul25(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntMul25Test", nil, b.N, nil)
}

func Benchmark_TestBigIntMul32(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntMul32Test", nil, b.N, nil)
}

func Benchmark_TestBigIntTDiv(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntTDivTest", nil, b.N, nil)
}

func Benchmark_TestBigIntTMod(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntTModTest", nil, b.N, nil)
}

func Benchmark_TestBigIntEDiv(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntEDivTest", nil, b.N, nil)
}

func Benchmark_TestBigIntEMod(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntEModTest", nil, b.N, nil)
}

func Benchmark_TestBigIntShr(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntShrTest", nil, b.N, nil)
}

func Benchmark_TestBigIntSetup(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/bigInt/output/cApiTest.wasm", 0, "bigIntInitSetup", nil, b.N, nil)
}

func Benchmark_TestCryptoSHA256(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "sha256Test", nil, b.N, nil)
}

func Benchmark_TestCryptoKeccak256(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "keccak256Test", nil, b.N, nil)
}

func Benchmark_TestCryptoRipMed160(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "ripemd160Test", nil, b.N, nil)
}

func Benchmark_TestCryptoBLS(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifyBLSTest", nil, b.N, nil)
}

func Benchmark_TestCryptoVerifyED25519(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifyEd25519Test", nil, b.N, nil)
}

func Benchmark_TestCryptoSecp256k1UnCompressed(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifySecp256k1UncompressedKeyTest", nil, b.N, nil)
}

func Benchmark_TestCryptoSecp256k1Compressed(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "verifySecp256k1CompressedKeyTest", nil, b.N, nil)
}

func Benchmark_TestEllipticCurveInitialVariablesAndCalls(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "initialVariablesAndCallsTest", nil, b.N, nil)
}

/// ELLIPTIC CURVES

func Benchmark_TestEllipticCurveAddP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224AddEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveAddP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256AddEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveAddP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384AddEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveAddP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521AddEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDoubleP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224DoubleEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDoubleP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256DoubleEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDoubleP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384DoubleEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDoubleP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521DoubleEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurveP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224IsOnCurveEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurveP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256IsOnCurveEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurveP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384IsOnCurveEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurveP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521IsOnCurveEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224MarshalEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256MarshalEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384MarshalEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521MarshalEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224UnmarshalEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256UnmarshalEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384UnmarshalEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521UnmarshalEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressedP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224MarshalCompressedEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressedP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256MarshalCompressedEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressedP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384MarshalCompressedEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressedP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521MarshalCompressedEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressedP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224UnmarshalCompressedEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressedP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256UnmarshalCompressedEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressedP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384UnmarshalCompressedEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressedP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521UnmarshalCompressedEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKeyP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224GenerateKeyEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKeyP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256GenerateKeyEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKeyP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384GenerateKeyEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKeyP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521GenerateKeyEcTest", getNumberOfRepsArgument(), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP224(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224ScalarMultEcTest", getNumberOfRepsAndScalarLengthArgs(10), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP256(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256ScalarMultEcTest", getNumberOfRepsAndScalarLengthArgs(10), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP384(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384ScalarMultEcTest", getNumberOfRepsAndScalarLengthArgs(10), b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP521(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521ScalarMultEcTest", getNumberOfRepsAndScalarLengthArgs(10), b.N, gasSchedule)
}

func Benchmark_TestCryptoDoNothing(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "doNothing", nil, b.N, nil)
}

func Benchmark_TestStorageRust(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	buff := make([]byte, 100)
	_, _ = rand.Read(buff)
	arguments := [][]byte{buff, big.NewInt(100).Bytes()}
	runWASMVMBenchmark(b, "../testdata/misc/str-repeat.wasm", 0, "repeat", arguments, b.N, gasSchedule)
}

func TestGasModel(t *testing.T) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")

	totalOp := uint64(0)
	for _, opCodeClass := range gasSchedule {
		for _, opCode := range opCodeClass {
			totalOp += opCode
		}
	}

	fmt.Println("gasSchedule: " + big.NewInt(int64(totalOp)).String())
	fmt.Println("ERC20 BIGINT")
	durations, err := DeployAndExecuteERC20WithBigInt(1, 100, gasSchedule, "../testdata/erc20-c-03/wrc20_arwen.wasm", "transferToken")
	require.Nil(t, err)
	displayBenchmarksResults(durations)

	runWASMVMBenchmark(t, "../testdata/misc/cpucalculate_arwen/output/cpucalculate.wasm", 8000, "cpuCalculate", nil, 2, gasSchedule)

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

func getBigFloatArgumentsAndGasSchedule() ([][]byte, map[string]map[string]uint64) {
	bigFloatArguments := make([][]byte, numberOfArguments)
	bigFloatArguments[0] = []byte{0, 0, 27, 10}
	buffer := []byte{1, 10, 0, 0, 0, 53, 0, 255, 255, 255, 139, 30, 172, 233, 27, 5, 64, 0}
	for i := 1; i < numberOfArguments; i++ {
		bigFloatArguments[i] = buffer
	}
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	return bigFloatArguments, gasSchedule
}

func getRandomBigFloatArgumentsAndGasSchedule() ([][]byte, map[string]map[string]uint64) {
	bigFloatArguments := make([][]byte, numberOfArguments+1)
	bigFloatArguments[0] = []byte{0, 0, 27, 10}
	expMantByteArray := make([]byte, 9)

	for i := 1; i < numberOfArguments+1; i++ {
		rand.Read(expMantByteArray)
		bigFloatArguments[i] = append([]byte{1, 10, 0, 0, 0, 53, 0, 0}, expMantByteArray...)
		for floatArgumentIsNotValid(bigFloatArguments[i]) {
			rand.Read(expMantByteArray)
			bigFloatArguments[i] = append([]byte{1, 10, 0, 0, 0, 53, 0, 0, 0}, expMantByteArray...)
		}

		// this is used if negative numbers are wanted
		if i%2 == 0 {
			bigFloatArguments[i][1] = byte(11)
		}

	}
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	return bigFloatArguments, gasSchedule
}

func floatArgumentIsNotValid(argument []byte) bool {
	float := new(big.Float)
	_ = float.GobDecode(argument)
	exp := float.MantExp(nil)
	if exp > bigFloatMaxExponent || exp < bigFloatMinExponent {
		return true
	}
	return false
}

func getBigFloatMulArguments() [][]byte {
	bigFloatArguments := make([][]byte, numberOfArguments+1)
	bigFloatArguments[0] = []byte{0, 0, 27, 10}

	for i := 1; i < numberOfArguments+1; i++ {

		// this is used if negative numbers are wanted
		if i%2 == 0 {
			expMantByteArray := make([]byte, 9)
			rand.Read(expMantByteArray)
			bigFloatArguments[i] = append([]byte{1, 10, 0, 0, 0, 53, 0, 0, 0}, expMantByteArray...)
			for floatArgumentIsNotValid(bigFloatArguments[i]) || floatMulExceedsMaxOrMinExponent(bigFloatArguments[i-1], bigFloatArguments[i]) {
				rand.Read(expMantByteArray)
				bigFloatArguments[i] = append([]byte{1, 10, 0, 0, 0, 53, 0, 0, 0}, expMantByteArray...)
			}
		} else {
			expMantByteArray := make([]byte, 10)
			rand.Read(expMantByteArray)
			bigFloatArguments[i] = append([]byte{1, 10, 0, 0, 0, 53, 0, 0}, expMantByteArray...)
			for floatArgumentIsNotValid(bigFloatArguments[i]) || floatMulExceedsMaxOrMinExponent(bigFloatArguments[i-1], bigFloatArguments[i]) {
				rand.Read(expMantByteArray)
				bigFloatArguments[i] = append([]byte{1, 10, 0, 0, 0, 53, 0, 0}, expMantByteArray...)
			}
		}

	}
	return bigFloatArguments
}

func floatMulExceedsMaxOrMinExponent(argument1 []byte, argument2 []byte) bool {
	float1, float2, resultMul := new(big.Float), new(big.Float), new(big.Float)
	_, _ = float1.GobDecode(argument1), float2.GobDecode(argument2)
	resultMul.Mul(float1, float2)
	resultExponent := resultMul.MantExp(nil)
	if resultExponent > bigFloatMaxExponent || resultExponent < bigFloatMinExponent {
		return true
	}
	return false
}

// func getRandomBuffers() [][]byte {
// 	bigFloatArguments := make([][]byte, numberOfArguments+1)
// 	bigFloatArguments[0] = []byte{0, 0, 27, 10}
// 	randomBuffer := make([]byte, 10000)
// 	for i := 1; i < numberOfArguments+1; i++ {
// 		rand.Read(randomBuffer)
// 		copy(bigFloatArguments[i], randomBuffer)
// 	}
// 	return bigFloatArguments
// }

func getRandomBigIntArgumentsAndGasSchedule(byteLength int32) ([][]byte, map[string]map[string]uint64) {
	bigIntArguments := make([][]byte, numberOfArguments+1)
	randomBigIntBytes := make([]byte, byteLength)
	for i := 1; i < numberOfArguments+1; i++ {
		rand.Read(randomBigIntBytes)
		copy(bigIntArguments[i], randomBigIntBytes)
	}

	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV5.toml")
	return bigIntArguments, gasSchedule
}
