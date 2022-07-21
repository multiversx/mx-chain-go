package block

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type blockProcessor interface {
	removeStartOfEpochBlockDataFromPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error
}

type gasConsumedProvider interface {
	TotalGasProvided() uint64
	TotalGasProvidedWithScheduled() uint64
	TotalGasRefunded() uint64
	TotalGasPenalized() uint64
	IsInterfaceNil() bool
}

type peerAccountsDBHandler interface {
	MarkSnapshotDone()
}

type receiptsRepository interface {
	SaveReceipts(holder *process.ReceiptsHolder, header data.HeaderHandler, headerHash []byte) error
	IsInterfaceNil() bool
}
