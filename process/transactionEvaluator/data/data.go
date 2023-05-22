package data

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// SimulationResultsWithVMOutput is the data transfer object which will hold results for simulation a transaction's execution
type SimulationResultsWithVMOutput struct {
	transaction.SimulationResults
	VMOutput *vmcommon.VMOutput `json:"-"`
}
