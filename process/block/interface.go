package block

import (
	"github.com/multiversx/mx-chain-go/common"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-core-go/data"
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
	GetEpochStartHandler() data.EpochStartHandler
	GetLastFinalizedCrossChainHeaderHandler() data.EpochStartChainDataHandler
}

type crossNotarizer interface {
	getLastCrossNotarizedHeaders() []bootstrapStorage.BootstrapHeaderInfo
}

// BlockProcessorCreator defines the block processor factory handler
type BlockProcessorCreator interface {
	CreateBlockProcessor(argumentsBaseProcessor ArgBaseProcessor, argsMetaProcessorCreateFunc ExtraMetaBlockProcessorCreateFunc) (process.DebuggerBlockProcessor, error)
	IsInterfaceNil() bool
}

// HeaderValidatorCreator is an interface for creating header validators
type HeaderValidatorCreator interface {
	CreateHeaderValidator(args ArgsHeaderValidator) (process.HeaderConstructionValidator, error)
	IsInterfaceNil() bool
}

type runTypeComponentsHolder interface {
	AccountsCreator() state.AccountFactory
	OutGoingOperationsPoolHandler() sovereignBlock.OutGoingOperationsPool
	DataCodecHandler() sovereign.DataCodecHandler
	TopicsCheckerHandler() sovereign.TopicsCheckerHandler
	IsInterfaceNil() bool
}

// ExtraMetaBlockProcessorCreateFunc defines a func able to create extra args needed to create meta block processor args
type ExtraMetaBlockProcessorCreateFunc func(systemVM vmcommon.VMExecutionHandler) (*ExtraArgsMetaBlockProcessor, error)
