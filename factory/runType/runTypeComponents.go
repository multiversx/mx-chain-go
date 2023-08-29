package runType

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
)

// RunTypeComponentsFactoryArgs holds the arguments needed for creating a state components factory
type RunTypeComponentsFactoryArgs struct {
	BlockChainHookHandlerCreator        factory.BlockChainHookHandlerCreator
	EpochStartBootstrapperCreator       factory.EpochStartBootstrapperCreator
	BootstrapperFromStorageCreator      factory.BootstrapperFromStorageCreator
	BlockProcessorCreator               factory.BlockProcessorCreator
	ForkDetectorCreator                 factory.ForkDetectorCreator
	BlockTrackerCreator                 factory.BlockTrackerCreator
	RequestHandlerCreator               factory.RequestHandlerCreator
	HeaderValidatorCreator              factory.HeaderValidatorCreator
	ScheduledTxsExecutionCreator        factory.ScheduledTxsExecutionCreator
	TransactionCoordinatorCreator       factory.TransactionCoordinatorCreator
	ValidatorStatisticsProcessorCreator factory.ValidatorStatisticsProcessorCreator
	AdditionalStorageServiceCreator     factory.AdditionalStorageServiceCreator
}

type runTypeComponentsFactory struct {
	blockChainHookHandlerCreator        factory.BlockChainHookHandlerCreator
	epochStartBootstrapperCreator       factory.EpochStartBootstrapperCreator
	bootstrapperFromStorageCreator      factory.BootstrapperFromStorageCreator
	blockProcessorCreator               factory.BlockProcessorCreator
	forkDetectorCreator                 factory.ForkDetectorCreator
	blockTrackerCreator                 factory.BlockTrackerCreator
	requestHandlerCreator               factory.RequestHandlerCreator
	headerValidatorCreator              factory.HeaderValidatorCreator
	scheduledTxsExecutionCreator        factory.ScheduledTxsExecutionCreator
	transactionCoordinatorCreator       factory.TransactionCoordinatorCreator
	validatorStatisticsProcessorCreator factory.ValidatorStatisticsProcessorCreator
	additionalStorageServiceCreator     factory.AdditionalStorageServiceCreator
}

// runTypeComponents struct holds the components needed for a run type
type runTypeComponents struct {
	blockChainHookHandlerCreator        factory.BlockChainHookHandlerCreator
	epochStartBootstrapperCreator       factory.EpochStartBootstrapperCreator
	bootstrapperFromStorageCreator      factory.BootstrapperFromStorageCreator
	blockProcessorCreator               factory.BlockProcessorCreator
	forkDetectorCreator                 factory.ForkDetectorCreator
	blockTrackerCreator                 factory.BlockTrackerCreator
	requestHandlerCreator               factory.RequestHandlerCreator
	headerValidatorCreator              factory.HeaderValidatorCreator
	scheduledTxsExecutionCreator        factory.ScheduledTxsExecutionCreator
	transactionCoordinatorCreator       factory.TransactionCoordinatorCreator
	validatorStatisticsProcessorCreator factory.ValidatorStatisticsProcessorCreator
	additionalStorageServiceCreator     factory.AdditionalStorageServiceCreator
}

// NewRunTypeComponentsFactory will return a new instance of runTypeComponentsFactory
func NewRunTypeComponentsFactory(args RunTypeComponentsFactoryArgs) (*runTypeComponentsFactory, error) {
	if check.IfNil(args.BlockChainHookHandlerCreator) {
		return nil, errors.ErrNilBlockChainHookHandlerCreator
	}
	if check.IfNil(args.EpochStartBootstrapperCreator) {
		return nil, errors.ErrNilEpochStartBootstrapperCreator
	}
	if check.IfNil(args.BootstrapperFromStorageCreator) {
		return nil, errors.ErrNilBootstrapperFromStorageCreator
	}
	if check.IfNil(args.BlockProcessorCreator) {
		return nil, errors.ErrNilBlockProcessorCreator
	}
	if check.IfNil(args.ForkDetectorCreator) {
		return nil, errors.ErrNilForkDetectorCreator
	}
	if check.IfNil(args.BlockTrackerCreator) {
		return nil, errors.ErrNilBlockTrackerCreator
	}
	if check.IfNil(args.RequestHandlerCreator) {
		return nil, errors.ErrNilRequestHandlerCreator
	}
	if check.IfNil(args.HeaderValidatorCreator) {
		return nil, errors.ErrNilHeaderValidatorCreator
	}
	if check.IfNil(args.ScheduledTxsExecutionCreator) {
		return nil, errors.ErrNilScheduledTxsExecutionCreator
	}
	if check.IfNil(args.TransactionCoordinatorCreator) {
		return nil, errors.ErrNilTransactionCoordinatorCreator
	}
	if check.IfNil(args.ValidatorStatisticsProcessorCreator) {
		return nil, errors.ErrNilValidatorStatisticsProcessorCreator
	}
	if check.IfNil(args.AdditionalStorageServiceCreator) {
		return nil, errors.ErrNilAdditionalStorageServiceCreator
	}
	return &runTypeComponentsFactory{
		blockChainHookHandlerCreator:        args.BlockChainHookHandlerCreator,
		epochStartBootstrapperCreator:       args.EpochStartBootstrapperCreator,
		bootstrapperFromStorageCreator:      args.BootstrapperFromStorageCreator,
		blockProcessorCreator:               args.BlockProcessorCreator,
		forkDetectorCreator:                 args.ForkDetectorCreator,
		blockTrackerCreator:                 args.BlockTrackerCreator,
		requestHandlerCreator:               args.RequestHandlerCreator,
		headerValidatorCreator:              args.HeaderValidatorCreator,
		scheduledTxsExecutionCreator:        args.ScheduledTxsExecutionCreator,
		transactionCoordinatorCreator:       args.TransactionCoordinatorCreator,
		validatorStatisticsProcessorCreator: args.ValidatorStatisticsProcessorCreator,
		additionalStorageServiceCreator:     args.AdditionalStorageServiceCreator,
	}, nil
}

// Create creates the runType components
func (rcf *runTypeComponentsFactory) Create() (*runTypeComponents, error) {
	return &runTypeComponents{
		blockChainHookHandlerCreator:        rcf.blockChainHookHandlerCreator,
		epochStartBootstrapperCreator:       rcf.epochStartBootstrapperCreator,
		bootstrapperFromStorageCreator:      rcf.bootstrapperFromStorageCreator,
		blockProcessorCreator:               rcf.blockProcessorCreator,
		forkDetectorCreator:                 rcf.forkDetectorCreator,
		blockTrackerCreator:                 rcf.blockTrackerCreator,
		requestHandlerCreator:               rcf.requestHandlerCreator,
		headerValidatorCreator:              rcf.headerValidatorCreator,
		scheduledTxsExecutionCreator:        rcf.scheduledTxsExecutionCreator,
		transactionCoordinatorCreator:       rcf.transactionCoordinatorCreator,
		validatorStatisticsProcessorCreator: rcf.validatorStatisticsProcessorCreator,
		additionalStorageServiceCreator:     rcf.additionalStorageServiceCreator,
	}, nil
}

// Close closes all underlying components that need closing
func (rc *runTypeComponents) Close() error {
	return nil
}
