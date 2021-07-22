// +build !race

// TODO remove build condition above to allow -race -short, after Arwen fix

package arwenvm

import (
	"crypto/rand"
	"fmt"

	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/stretchr/testify/require"
)

func Benchmark_VmDeployWithFibbonacciAndExecute(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/misc/fib_arwen/output/fib_arwen.wasm", 32, "_main", nil, b.N, nil)
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

func Benchmark_TestEllipticCurveAdd(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "addEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveAddP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224AddEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveAddP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256AddEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveAddP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384AddEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveAddP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521AddEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDouble(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "doubleEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDoubleP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224DoubleEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDoubleP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256DoubleEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDoubleP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384DoubleEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveDoubleP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521DoubleEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurve(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "isOnCurveEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurveP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224IsOnCurveEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurveP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256IsOnCurveEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurveP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384IsOnCurveEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveIsOnCurveP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521IsOnCurveEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshal(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "marshalEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224MarshalEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256MarshalEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384MarshalEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521MarshalEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshal(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "unmarshalEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224UnmarshalEcTest", nil, b.N, gasSchedule)
}
func Benchmark_TestEllipticCurveUnmarshalP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256UnmarshalEcTest", nil, b.N, gasSchedule)
}
func Benchmark_TestEllipticCurveUnmarshalP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384UnmarshalEcTest", nil, b.N, gasSchedule)
}
func Benchmark_TestEllipticCurveUnmarshalP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521UnmarshalEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressed(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "marshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressedP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224MarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressedP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256MarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressedP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384MarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveMarshalCompressedP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521MarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressed(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "unmarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressedP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224UnmarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressedP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256UnmarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressedP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384UnmarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveUnmarshalCompressedP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521UnmarshalCompressedEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKey(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "generateKeyEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKeyP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224GenerateKeyEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKeyP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256GenerateKeyEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKeyP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384GenerateKeyEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveGenerateKeyP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521GenerateKeyEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMult(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "scalarMultEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP224(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p224ScalarMultEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP256(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p256ScalarMultEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP384(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p384ScalarMultEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestEllipticCurveScalarMultP521(b *testing.B) {
	gasSchedule, _ := core.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	runWASMVMBenchmark(b, "../testdata/c-api-tests/ecBenchmark/output/ecBenchmark.wasm", 0, "p521ScalarMultEcTest", nil, b.N, gasSchedule)
}

func Benchmark_TestCryptoDoNothing(b *testing.B) {
	runWASMVMBenchmark(b, "../testdata/c-api-tests/crypto/output/cryptoTest.wasm", 0, "doNothing", nil, b.N, nil)
}

func Benchmark_TestStorageRust(b *testing.B) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")
	buff := make([]byte, 100)
	_, _ = rand.Read(buff)
	arguments := [][]byte{buff, big.NewInt(100).Bytes()}
	runWASMVMBenchmark(b, "../testdata/misc/str-repeat.wasm", 0, "repeat", arguments, b.N, gasSchedule)
}

func TestGasModel(t *testing.T) {
	gasSchedule, _ := common.LoadGasScheduleConfig("../../../../cmd/node/config/gasSchedules/gasScheduleV2.toml")

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
	result, err := RunTest(fileSC, testingValue, function, arguments, numRun, gasSchedule)
	require.Nil(tb, err)

	log.Info("test completed",
		"function", result.FunctionName,
		"consumed gas", result.GasUsed,
		"time took", result.ExecutionTimeSpan,
		"numRun", numRun,
	)
}
