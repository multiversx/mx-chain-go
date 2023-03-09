package testadder

import (
	"testing"

	mc "github.com/multiversx/mx-chain-go/integrationTests/vm/wasm/wasmvm/scenariosConverter"
)

func TestScenariosConverter_AdderWithExternalSteps(t *testing.T) {
	mc.CheckConverter(t, "./adder_with_external_steps.scen.json")
}

func Benchmark_ScenariosConverter_AdderWithExternalSteps(b *testing.B) {
	mc.BenchmarkScenariosSpecificTx(b, "./adder_with_external_steps.scen.json")
}
