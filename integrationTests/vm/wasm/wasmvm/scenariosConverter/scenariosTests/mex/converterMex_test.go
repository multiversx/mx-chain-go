package Benchmark_ScenariosConverter_Mex

import (
	"testing"

	mc "github.com/multiversx/mx-chain-go/integrationTests/vm/wasm/wasmvm/scenariosConverter"
)

func TestScenariosConverter_MexState(t *testing.T) {
	mc.CheckConverter(t, "./swap_fixed_input.scen.json")
}
func Benchmark_ScenariosConverter_SwapFixedInput(b *testing.B) {
	mc.BenchmarkScenariosSpecificTx(b, "./swap_fixed_input.scen.json")
}
