package block

import (
	"github.com/multiversx/mx-chain-core-go/data"
	sovereignCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
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

type sovereignChainHeader interface {
	GetExtendedShardHeaderHashes() [][]byte
	GetOutGoingMiniBlockHeaderHandler() data.OutGoingMiniBlockHeaderHandler
}

type crossNotarizer interface {
	getLastCrossNotarizedHeaders() []bootstrapStorage.BootstrapHeaderInfo
}

// OutGoingOperationsPool defines the behavior of a timed cache for outgoing operations
type OutGoingOperationsPool interface {
	Add(data *sovereignCore.BridgeOutGoingData)
	Get(hash []byte) *sovereignCore.BridgeOutGoingData
	Delete(hash []byte)
	GetUnconfirmedOperations() []*sovereignCore.BridgeOutGoingData
	ConfirmOperation(hashOfHashes []byte, hash []byte) error
	IsInterfaceNil() bool
}
