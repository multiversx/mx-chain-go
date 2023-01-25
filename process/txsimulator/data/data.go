package data

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// SimulationResults is the data transfer object which will hold results for simulation a transaction's execution
type SimulationResults struct {
	Status     transaction.TxStatus                           `json:"status,omitempty"`
	FailReason string                                         `json:"failReason,omitempty"`
	ScResults  map[string]*transaction.ApiSmartContractResult `json:"scResults,omitempty"`
	Receipts   map[string]*transaction.ApiReceipt             `json:"receipts,omitempty"`
	Hash       string                                         `json:"hash,omitempty"`
	VMOutput   *vmcommon.VMOutput                             `json:"-"`
}
