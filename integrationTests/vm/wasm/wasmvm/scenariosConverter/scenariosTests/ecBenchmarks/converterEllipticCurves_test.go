package Benchmark_TestEllipticCurveScalarMultP224

import (
	"testing"

	mc "github.com/multiversx/mx-chain-go/integrationTests/vm/wasm/wasmvm/scenariosConverter"
)

func TestScenariosConverter_EllipticCurves(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	mc.CheckConverter(t, "./elliptic_curves.scen.json")
}

func Benchmark_ScenariosConverter_EllipticCurves(b *testing.B) {
	if testing.Short() {
		b.Skip("this is not a short benchmark")
	}

	mc.BenchmarkScenariosSpecificTx(b, "./elliptic_curves.scen.json")
}
