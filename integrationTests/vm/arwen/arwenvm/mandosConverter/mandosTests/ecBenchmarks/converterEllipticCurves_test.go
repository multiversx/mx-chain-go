package Benchmark_TestEllipticCurveScalarMultP224

import (
	"fmt"
	"testing"

	mc "github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm/mandosConverter"
)

func Benchmark_MandosConverter_EllipticCurves(b *testing.B) {
	testContext, transactions, err := mc.SetStateFromMandosTest("./elliptic_curves.scen.json")
	if err != nil {
		fmt.Println("Setting state went wrong: ", err)
		return
	}
	b.ResetTimer()
	err = mc.RunSingleTransactionBenchmark(b, testContext, transactions[0])
	if err != nil {
		fmt.Println("Proccess transaction went wrong: ", err)
	}
}
