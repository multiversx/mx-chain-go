package outport

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// Driver defines the interface of the outport driver
type Driver interface {
	DigestBlock(header data.HeaderHandler, body data.BodyHandler, txCoordinator process.TransactionCoordinator)
	IsInterfaceNil() bool
}
