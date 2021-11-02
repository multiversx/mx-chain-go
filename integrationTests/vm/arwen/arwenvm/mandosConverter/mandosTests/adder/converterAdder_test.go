package testadder

import (
	"testing"

	mc "github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm/mandosConverter"
)

func TestMandosConverter_AdderWithExternalSteps(t *testing.T) {
	mc.CheckConverter(t, "./adder_with_external_steps.scen.json")
}

func Benchmark_MandosConverter_SimpleAdderWithDeploy(b *testing.B) {
	mc.BenchmarkMandosSpecificTx(b, "./adder.scen.json")
}

func Benchmark_MandosConverter_AdderWithExternalSteps(b *testing.B) {
	mc.BenchmarkMandosSpecificTx(b, "./adder_with_external_steps.scen.json")
}
