package outport

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TransactionLogProcessor defines the interface of the logs processor
type TransactionLogProcessor interface {
	GetLog(txHash []byte) (data.LogHandler, error)
	SaveLog(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error
	IsInterfaceNil() bool
}

// Driver defines the interface of the outport driver
type Driver interface {
	DigestCommittedBlock(header data.HeaderHandler)
	// TODO: add DigestInvalidTransaction()
	IsInterfaceNil() bool
}

// TransactionCoordinator defines the interface of the transactions coordinator
type TransactionCoordinator interface {
	GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler
	IsInterfaceNil() bool
}
