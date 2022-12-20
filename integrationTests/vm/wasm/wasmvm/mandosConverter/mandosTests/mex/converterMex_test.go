package Benchmark_MandosConverter_Mex

import (
	"testing"

	mc "github.com/ElrondNetwork/elrond-go/integrationTests/vm/wasm/wasmvm/mandosConverter"
)

func TestMandosConverter_MexState(t *testing.T) {
	mc.CheckConverter(t, "./swap_fixed_input.scen.json")
}
func Benchmark_MandosConverter_SwapFixedInput(b *testing.B) {
	mc.BenchmarkMandosSpecificTx(b, "./swap_fixed_input.scen.json")
}
