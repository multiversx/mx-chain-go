package runType

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
)

// SovereignRunTypeComponentsFactoryArgs holds the arguments needed for creating a state components factory
type SovereignRunTypeComponentsFactoryArgs struct {
	RunTypeComponentsFactory *runTypeComponentsFactory
}

type sovereignRunTypeComponentsFactory struct {
	runTypeComponentsFactory *runTypeComponentsFactory
}

// NewSovereignRunTypeComponentsFactory will return a new instance of runTypeComponentsFactory
func NewSovereignRunTypeComponentsFactory(args SovereignRunTypeComponentsFactoryArgs) (*sovereignRunTypeComponentsFactory, error) {
	if check.IfNil(args.RunTypeComponentsFactory) {
		return nil, errors.ErrNilRunTypeComponentsFactory
	}

	return &sovereignRunTypeComponentsFactory{
		runTypeComponentsFactory: args.RunTypeComponentsFactory,
	}, nil
}

// Create creates the runType components
func (rcf *sovereignRunTypeComponentsFactory) Create() (*runTypeComponents, error) {
	rtc, err := rcf.runTypeComponentsFactory.Create()
	if err != nil {
		return nil, err
	}

	blockChainHookHandlerFactory, err := hooks.NewSovereignBlockChainHookFactory(rtc.blockChainHookHandlerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockChainHookFactory failed: %w", err)
	}

	epochStartBootstrapperFactory, err := bootstrap.NewSovereignEpochStartBootstrapperFactory(rtc.epochStartBootstrapperCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignEpochStartBootstrapperFactory failed: %w", err)
	}

	bootstrapperFromStorageFactory, err := storageBootstrap.NewSovereignShardStorageBootstrapperFactory(rtc.bootstrapperFromStorageCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignShardStorageBootstrapperFactory failed: %w", err)
	}

	blockProcessorFactory, err := block.NewSovereignBlockProcessorFactory(rtc.blockProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockProcessorFactory failed: %w", err)
	}

	forkDetectorFactory, err := sync.NewSovereignForkDetectorFactory(rtc.forkDetectorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignForkDetectorFactory failed: %w", err)
	}

	blockTrackerFactory, err := track.NewSovereignBlockTrackerFactory(rtc.blockTrackerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignBlockTrackerFactory failed: %w", err)
	}

	requestHandlerFactory, err := requestHandlers.NewSovereignResolverRequestHandlerFactory(rtc.requestHandlerCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignResolverRequestHandlerFactory failed: %w", err)
	}

	headerValidatorFactory, err := block.NewSovereignHeaderValidatorFactory(rtc.headerValidatorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignHeaderValidatorFactory failed: %w", err)
	}

	scheduledTxsExecutionFactory, err := preprocess.NewSovereignScheduledTxsExecutionFactory()
	if err != nil {
		return nil, fmt.Errorf("createManagedRunTypeComponents - NewSovereignScheduledTxsExecutionFactory failed: %w", err)
	}

	transactionCoordinatorFactory, err := coordinator.NewSovereignTransactionCoordinatorFactory(rtc.transactionCoordinatorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignTransactionCoordinatorFactory failed: %w", err)
	}

	validatorStatisticsProcessorFactory, err := peer.NewSovereignValidatorStatisticsProcessorFactory(rtc.validatorStatisticsProcessorCreator)
	if err != nil {
		return nil, fmt.Errorf("sovereignRunTypeComponentsFactory - NewSovereignValidatorStatisticsProcessorFactory failed: %w", err)
	}

	return &runTypeComponents{
		blockChainHookHandlerCreator:        blockChainHookHandlerFactory,
		epochStartBootstrapperCreator:       epochStartBootstrapperFactory,
		bootstrapperFromStorageCreator:      bootstrapperFromStorageFactory,
		blockProcessorCreator:               blockProcessorFactory,
		forkDetectorCreator:                 forkDetectorFactory,
		blockTrackerCreator:                 blockTrackerFactory,
		requestHandlerCreator:               requestHandlerFactory,
		headerValidatorCreator:              headerValidatorFactory,
		scheduledTxsExecutionCreator:        scheduledTxsExecutionFactory,
		transactionCoordinatorCreator:       transactionCoordinatorFactory,
		validatorStatisticsProcessorCreator: validatorStatisticsProcessorFactory,
	}, nil
}
