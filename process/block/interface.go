package block

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type blockProcessor interface {
	removeStartOfEpochBlockDataFromPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error
}

// OutportDriver defines the interface of the outport driver
type OutportDriver interface {
	DigestBlock(header data.HeaderHandler, body data.BodyHandler, txCoordinator process.TransactionCoordinator)
	IsInterfaceNil() bool
}
