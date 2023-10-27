package block

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
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
	SaveReceipts(holder common.ReceiptsHolder, header data.HeaderHandler, headerHash []byte) error
	IsInterfaceNil() bool
}

type validatorStatsRootHashGetter interface {
	GetValidatorStatsRootHash() []byte
}

type extendedShardHeaderHashesGetter interface {
	GetExtendedShardHeaderHashes() [][]byte
}

type crossNotarizer interface {
	getLastCrossNotarizedHeaders() []bootstrapStorage.BootstrapHeaderInfo
}

type OutGoingOperationsPool interface {
	Add(hash []byte, data []byte)
	Get(hash []byte) []byte
	Delete(hash []byte)
	GetUnconfirmedOperations() [][]byte
	IsInterfaceNil() bool
}
