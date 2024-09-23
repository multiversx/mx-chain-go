package Benchmark_ScenariosConverter_Mex

import (
	"testing"

	mc "github.com/multiversx/mx-chain-go/integrationTests/vm/wasm/wasmvm/scenariosConverter"
)

func TestScenariosConverter_MexState(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	mc.CheckConverter(t, "./swap_fixed_input.scen.json")
}
func Benchmark_ScenariosConverter_SwapFixedInput(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	mc.BenchmarkScenariosSpecificTx(b, "./swap_fixed_input.scen.json")
}
