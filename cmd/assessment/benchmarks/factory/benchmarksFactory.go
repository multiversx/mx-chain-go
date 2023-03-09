package factory

import (
	"path/filepath"

	"github.com/multiversx/mx-chain-go/cmd/assessment/benchmarks"
)

// CreateBenchmarksList creates the list of benchmarks
func CreateBenchmarksList(testDataDirectory string) []benchmarks.BenchmarkRunner {
	list := make([]benchmarks.BenchmarkRunner, 0)

	list = append(list, createFibBenchmark(testDataDirectory))
	list = append(list, createCPUCalculateBenchmark(testDataDirectory))
	list = append(list, createStoreBenchmark(testDataDirectory))
	list = append(list, createCApiBigInt(testDataDirectory))
	list = append(list, createCApiBigIntMul32(testDataDirectory))
	list = append(list, createCApiSha256(testDataDirectory))
	list = append(list, createCApiShaKeccak256(testDataDirectory))
	list = append(list, createCApiRipemd160(testDataDirectory))
	list = append(list, createCApiVerifyBLS(testDataDirectory))
	list = append(list, createCApiVerifyED25519(testDataDirectory))
	list = append(list, createCApiVerifySecp256k1Uncompressed(testDataDirectory))
	list = append(list, createCApiVerifySecp256k1Compressed(testDataDirectory))
	list = append(list, createDelegation(testDataDirectory))
	list = append(list, createErc20InC(testDataDirectory))
	list = append(list, createErc20InRust(testDataDirectory))

	return list
}

func createFibBenchmark(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "fibonacci",
		ScFilename:   filepath.Join(testDataDirectory, "fibonacci.wasm"),
		TestingValue: 32,
		Function:     "_main",
		Arguments:    nil,
		NumRuns:      10,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCPUCalculateBenchmark(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "cpu calculate",
		ScFilename:   filepath.Join(testDataDirectory, "cpucalculate.wasm"),
		TestingValue: 8000,
		Function:     "cpuCalculate",
		Arguments:    nil,
		NumRuns:      2000,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createStoreBenchmark(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "storage100",
		ScFilename:   filepath.Join(testDataDirectory, "storage100.wasm"),
		TestingValue: 0,
		Function:     "store100",
		Arguments:    nil,
		NumRuns:      2000,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiBigInt(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API big int",
		ScFilename:   filepath.Join(testDataDirectory, "cApiTest.wasm"),
		TestingValue: 0,
		Function:     "bigIntNewTest",
		Arguments:    nil,
		NumRuns:      700,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiBigIntMul32(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API big int mul 32",
		ScFilename:   filepath.Join(testDataDirectory, "cApiTest.wasm"),
		TestingValue: 0,
		Function:     "bigIntMul32Test",
		Arguments:    nil,
		NumRuns:      2,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiSha256(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API sha256",
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "sha256Test",
		Arguments:    nil,
		NumRuns:      30,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiShaKeccak256(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API keccak256",
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "keccak256Test",
		Arguments:    nil,
		NumRuns:      30,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiRipemd160(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API ripemd160",
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "ripemd160Test",
		Arguments:    nil,
		NumRuns:      30,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiVerifyBLS(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API verify BLS",
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "verifyBLSTest",
		Arguments:    nil,
		NumRuns:      5,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiVerifyED25519(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API verify ED25519",
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "verifyEd25519Test",
		Arguments:    nil,
		NumRuns:      50,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiVerifySecp256k1Uncompressed(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API verify secp256k1 uncompressed",
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "verifySecp256k1UncompressedKeyTest",
		Arguments:    nil,
		NumRuns:      400,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createCApiVerifySecp256k1Compressed(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgWasmBenchmark{
		Name:         "C API verify secp256k1 compressed",
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "verifySecp256k1CompressedKeyTest",
		Arguments:    nil,
		NumRuns:      200,
	}

	return benchmarks.NewWasmBenchmark(arg)
}

func createDelegation(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgDelegationBenchmark{
		Name:               "Delegation in Wasm VM",
		ScFilename:         filepath.Join(testDataDirectory, "delegation_v0_5_2_full.wasm"),
		NumRuns:            2,
		NumTxPerBatch:      100,
		NumBatches:         10,
		NumQueriesPerBatch: 0,
	}

	return benchmarks.NewDelegationBenchmark(arg)
}

func createErc20InC(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgErc20Benchmark{
		Name:               "ERC20 in C",
		ScFilename:         filepath.Join(testDataDirectory, "erc20_c.wasm"),
		Function:           "transferToken",
		NumRuns:            1,
		NumTransfersPerRun: 1000,
	}

	return benchmarks.NewErc20Benchmark(arg)
}

func createErc20InRust(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgErc20Benchmark{
		Name:               "ERC20 in Rust",
		ScFilename:         filepath.Join(testDataDirectory, "erc20_rust.wasm"),
		Function:           "transfer",
		NumRuns:            1,
		NumTransfersPerRun: 300,
	}

	return benchmarks.NewErc20Benchmark(arg)
}
