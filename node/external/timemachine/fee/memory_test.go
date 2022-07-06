package fee

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestFeeComputer_MemoryFootprint(t *testing.T) {
	numEpochs := 10000
	maxFootprintNumBytes := 20_000_000

	journal := &memoryFootprintJournal{}
	journal.before = getMemStats()

	computer, _ := NewFeeComputer(ArgsNewFeeComputer{
		BuiltInFunctionsCostHandler: &testscommon.BuiltInCostHandlerStub{},
		EconomicsConfig:             testscommon.GetEconomicsConfig(),
	})

	tx := &transaction.Transaction{
		GasLimit: 50000,
		GasPrice: 1000000000,
	}

	for i := 0; i < numEpochs; i++ {
		apiTx := &transaction.ApiTransactionResult{
			Epoch: uint32(i),
			Tx:    tx,
		}

		_ = computer.ComputeTransactionFee(apiTx)
	}

	journal.after = getMemStats()

	// This line protects the fee computer from being garbage-collected (for the purpose of the test).
	_ = computer.ComputeTransactionFee(&transaction.ApiTransactionResult{Epoch: uint32(0), Tx: tx})

	journal.display()
	require.Len(t, computer.economicsInstances, numEpochs)
	require.Less(t, journal.footprint(), uint64(maxFootprintNumBytes))
}

func getMemStats() runtime.MemStats {
	runtime.GC()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats
}

type memoryFootprintJournal struct {
	before runtime.MemStats
	after  runtime.MemStats
}

func (journal *memoryFootprintJournal) footprint() uint64 {
	return uint64(core.MaxInt(0, int(journal.after.HeapInuse)-int(journal.before.HeapInuse)))
}

func (journal *memoryFootprintJournal) display() {
	// See: https://golang.org/pkg/runtime/#MemStats

	fmt.Printf("before:\tHeapAlloc = %v MiB\tHeapInUse = %v MiB\n",
		bToMb(journal.before.HeapAlloc), bToMb(journal.before.HeapInuse))

	fmt.Printf("after:\tHeapAlloc = %v MiB\tHeapInUse = %v MiB\n",
		bToMb(journal.after.HeapAlloc), bToMb(journal.after.HeapInuse))

	fmt.Println("Footprint:", bToMb(journal.footprint()))
}

func bToMb(b uint64) int {
	return int(b / 1024 / 1024)
}
