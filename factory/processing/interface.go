package processing

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/track"
)

// GenesisMetaBlockChecker should handle genesis meta block checks after creation
type GenesisMetaBlockChecker interface {
	SetValidatorRootHashOnGenesisMetaBlock(genesisMetaBlock data.HeaderHandler, validatorStatsRootHash []byte) error
	IsInterfaceNil() bool
}

// TransactionCoordinatorCreator defines the transaction coordinator factory creator
type TransactionCoordinatorCreator interface {
	CreateTransactionCoordinator(argsTransactionCoordinator coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error)
	IsInterfaceNil() bool
}

// BlockProcessorCreator defines the block processor factory handler
type BlockProcessorCreator interface {
	CreateBlockProcessor(argumentsBaseProcessor block.ArgBaseProcessor) (process.DebuggerBlockProcessor, error)
	IsInterfaceNil() bool
}

// ValidatorStatisticsProcessorCreator is an interface for creating validator statistics processors
type ValidatorStatisticsProcessorCreator interface {
	CreateValidatorStatisticsProcessor(args peer.ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error)
	IsInterfaceNil() bool
}

// HeaderValidatorCreator is an interface for creating header validators
type HeaderValidatorCreator interface {
	CreateHeaderValidator(args block.ArgsHeaderValidator) (process.HeaderConstructionValidator, error)
	IsInterfaceNil() bool
}

// BlockTrackerCreator is an interface for creating block trackers
type BlockTrackerCreator interface {
	CreateBlockTracker(argShardTracker track.ArgShardTracker) (process.BlockTracker, error)
	IsInterfaceNil() bool
}

// ForkDetectorCreator is the interface needed by base fork detector to create fork detector
type ForkDetectorCreator interface {
	CreateForkDetector(args sync.ForkDetectorFactoryArgs) (process.ForkDetector, error)
	IsInterfaceNil() bool
}

// RequestHandlerCreator defines the resolver requester factory handler
type RequestHandlerCreator interface {
	CreateRequestHandler(resolverRequestArgs requestHandlers.RequestHandlerArgs) (process.RequestHandler, error)
	IsInterfaceNil() bool
}
