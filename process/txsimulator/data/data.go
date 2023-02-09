package data

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// SimulationResults is the data transfer object which will hold results for simulation a transaction's execution
type SimulationResults struct {
	transaction.SimulationResults
	VMOutput *vmcommon.VMOutput `json:"-"`
}
