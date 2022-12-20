package testadder

import (
	"testing"

	mc "github.com/ElrondNetwork/elrond-go/integrationTests/vm/wasm/wasmvm/mandosConverter"
)

func TestMandosConverter_AdderWithExternalSteps(t *testing.T) {
	mc.CheckConverter(t, "./adder_with_external_steps.scen.json")
}

func Benchmark_MandosConverter_AdderWithExternalSteps(b *testing.B) {
	mc.BenchmarkMandosSpecificTx(b, "./adder_with_external_steps.scen.json")
}
