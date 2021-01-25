package factory

import (
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/cmd/assessment/benchmarks"
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
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "fibonacci",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "fibonacci.wasm"),
		TestingValue: 32,
		Function:     "_main",
		Arguments:    nil,
		NumRuns:      10,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCPUCalculateBenchmark(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "cpu calculate",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cpucalculate.wasm"),
		TestingValue: 8000,
		Function:     "cpuCalculate",
		Arguments:    nil,
		NumRuns:      2000,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createStoreBenchmark(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "storage100",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "storage100.wasm"),
		TestingValue: 0,
		Function:     "store100",
		Arguments:    nil,
		NumRuns:      2000,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiBigInt(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API big int",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cApiTest.wasm"),
		TestingValue: 0,
		Function:     "bigIntNewTest",
		Arguments:    nil,
		NumRuns:      700,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiBigIntMul32(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API big int mul 32",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cApiTest.wasm"),
		TestingValue: 0,
		Function:     "bigIntMul32Test",
		Arguments:    nil,
		NumRuns:      2,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiSha256(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API sha256",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "sha256Test",
		Arguments:    nil,
		NumRuns:      30,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiShaKeccak256(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API keccak256",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "keccak256Test",
		Arguments:    nil,
		NumRuns:      30,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiRipemd160(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API ripemd160",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "ripemd160Test",
		Arguments:    nil,
		NumRuns:      30,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiVerifyBLS(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API verify BLS",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "verifyBLSTest",
		Arguments:    nil,
		NumRuns:      5,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiVerifyED25519(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API verify ED25519",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "verifyEd25519Test",
		Arguments:    nil,
		NumRuns:      50,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiVerifySecp256k1Uncompressed(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API verify secp256k1 uncompressed",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "verifySecp256k1UncompressedKeyTest",
		Arguments:    nil,
		NumRuns:      400,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createCApiVerifySecp256k1Compressed(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgArwenBenchmark{
		Name:         "C API verify secp256k1 compressed",
		GasFilename:  filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:   filepath.Join(testDataDirectory, "cryptoTest.wasm"),
		TestingValue: 0,
		Function:     "verifySecp256k1CompressedKeyTest",
		Arguments:    nil,
		NumRuns:      200,
	}

	return benchmarks.NewArwenBenchmark(arg)
}

func createDelegation(testDataDirectory string) benchmarks.BenchmarkRunner {
	arg := benchmarks.ArgDelegationBenchmark{
		Name:               "Delegation in Arwen",
		GasFilename:        filepath.Join(testDataDirectory, "testGasSchedule.toml"),
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
		GasFilename:        filepath.Join(testDataDirectory, "testGasSchedule.toml"),
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
		GasFilename:        filepath.Join(testDataDirectory, "testGasSchedule.toml"),
		ScFilename:         filepath.Join(testDataDirectory, "erc20_rust.wasm"),
		Function:           "transfer",
		NumRuns:            1,
		NumTransfersPerRun: 300,
	}

	return benchmarks.NewErc20Benchmark(arg)
}
