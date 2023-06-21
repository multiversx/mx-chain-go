package factory

import (
	"errors"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap"
	"github.com/multiversx/mx-chain-go/factory"
	processDisabled "github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
	"github.com/multiversx/mx-chain-go/process/track"
	"time"
)

// SovereignBlockChainHookFactory - factory for sovereign run
type SovereignBlockChainHookFactory struct {
}

func (bhf *SovereignBlockChainHookFactory) CreateBlockChainHook(args hooks.ArgBlockChainHook) (process.BlockChainHookHandler, error) {
	bh, _ := hooks.NewBlockChainHookImpl(args)
	return hooks.NewSovereignBlockChainHook(bh)
}

type SovereignBlockProcessorFactory struct {
}

func (s SovereignBlockProcessorFactory) CreateBlockProcessor(argumentsBaseProcessor factory.ArgBaseProcessor) (process.DebuggerBlockProcessor, error) {
	argShardProcessor := block.ArgShardProcessor{
		ArgBaseProcessor: argumentsBaseProcessor,
	}

	shardProcessor, err := block.NewShardProcessor(argShardProcessor)
	if err != nil {
		return nil, errors.New("could not create shard block processor: " + err.Error())
	}

	scbp, err := block.NewSovereignChainBlockProcessor(
		shardProcessor,
		argumentsBaseProcessor.ValidatorStatisticsProcessor,
	)

	return scbp, nil
}

type SovereignTransactionCoordinatorFactory struct {
}

func (tcf *SovereignTransactionCoordinatorFactory) CreateTransactionCoordinator(argsTransactionCoordinator coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	transactionCoordinator, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	if err != nil {
		return nil, err
	}

	return coordinator.NewSovereignChainTransactionCoordinator(transactionCoordinator)
}

type SovereignResolverRequestHandler struct {
}

func (rrh *SovereignResolverRequestHandler) CreateResolverRequestHandler(resolverRequestArgs factory.ResolverRequestArgs) (process.RequestHandler, error) {
	requestHandler, err := requestHandlers.NewResolverRequestHandler(
		resolverRequestArgs.RequestersFinder,
		resolverRequestArgs.RequestedItemsHandler,
		resolverRequestArgs.WhiteListHandler,
		common.MaxTxsToRequest,
		resolverRequestArgs.ShardId,
		time.Second,
	)
	if err != nil {
		return nil, err
	}

	return requestHandlers.NewSovereignResolverRequestHandler(requestHandler)
}

type SovereignScheduledTxsExecutionFactory struct {
}

func (stxef *SovereignScheduledTxsExecutionFactory) CreateScheduledTxsExecutionHandler(_ factory.ScheduledTxsExecutionFactoryArgs) (process.ScheduledTxsExecutionHandler, error) {
	return &processDisabled.ScheduledTxsExecutionHandler{}, nil
}

type SovereignValidatorStatisticsFactory struct {
}

func (vsf *SovereignValidatorStatisticsFactory) CreateValidatorStatisticsProcessor(args peer.ArgValidatorStatisticsProcessor) (process.ValidatorStatisticsProcessor, error) {
	validatorStatisticsProcessor, err := peer.NewValidatorStatisticsProcessor(args)
	if err != nil {
		return nil, err
	}

	return peer.NewSovereignChainValidatorStatisticsProcessor(validatorStatisticsProcessor)
}

type SovereignHeaderValidatorFactory struct {
}

func (hvf *SovereignHeaderValidatorFactory) CreateHeaderValidator(args factory.ArgsHeaderValidator) (process.HeaderConstructionValidator, error) {
	headerValidator, err := block.NewHeaderValidator(args)
	if err != nil {
		return nil, err
	}

	return block.NewSovereignChainHeaderValidator(headerValidator)
}

type SovereignBlockTrackerFactory struct {
}

func (btf *SovereignBlockTrackerFactory) CreateShardBlockTracker(argBaseTracker track.ArgBaseTracker) (process.BlockTracker, error) {
	arguments := track.ArgShardTracker{
		ArgBaseTracker: argBaseTracker,
	}
	blockTracker, err := track.NewShardBlockTrack(arguments)
	if err != nil {
		return nil, err
	}

	return track.NewSovereignChainShardBlockTrack(blockTracker)
}

type SovereignShardForkDetectorFactory struct {
}

func (sfd *SovereignShardForkDetectorFactory) CreateShardForkDetector(args factory.ShardForkDetectorFactoryArgs) (process.ForkDetector, error) {
	forkDetector, err := sync.NewShardForkDetector(args.RoundHandler, args.HeaderBlackList, args.BlockTracker, args.GenesisTime)
	if err != nil {
		return nil, err
	}

	return sync.NewSovereignChainShardForkDetector(forkDetector)
}

type SovereignPreProcessorsFactory struct {
}

func (ppcf *SovereignPreProcessorsFactory) CreatePreProcessor(args process.PreProcessorsFactoryArgs) (process.PreProcessor, error) {
	// TODO: refactor this: 	ppcf.requestHandler.RequestUnsignedTransactions
	return preprocess.NewSmartContractResultPreprocessor(
		args.UnsignedTransactions,
		args.Store,
		args.Hasher,
		args.Marshaller,
		args.ScResultProcessor,
		args.ShardCoordinator,
		args.Accounts,
		args.RequestHandler.RequestUnsignedTransactions,
		args.GasHandler,
		args.EconomicsFee,
		args.PubkeyConverter,
		args.BlockSizeComputation,
		args.BalanceComputation,
		args.EnableEpochsHandler,
		args.ProcessedMiniBlocksTracker,
	)
}

type SovereignEpochStartBootstrapperFactory struct {
}

func (bcf *SovereignEpochStartBootstrapperFactory) CreateEpochStartBootstrapper(epochStartBootstrapArgs bootstrap.ArgsEpochStartBootstrap) (factory.EpochStartBootstrapper, error) {
	epochStartBootstrapper, err := bootstrap.NewEpochStartBootstrap(epochStartBootstrapArgs)
	if err != nil {
		return nil, err
	}

	return bootstrap.NewSovereignChainEpochStartBootstrap(epochStartBootstrapper)
}

type SovereignSCRProcessorFactory struct {
}

func (s *SovereignSCRProcessorFactory) CreateSCRProcessor(args process.ArgsNewSmartContractProcessor) (process.SCRProcessorHandler, error) {
	scrProc, _ := smartContract.NewSmartContractProcessor(args)
	return smartContract.NewSovereignSCRProcessor(scrProc)
}

type SovereignShardBootstrapFactory struct {
}

func (sbf *SovereignShardBootstrapFactory) CreateShardBootstrapFactory(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
	bootstrapper, err := sync.NewShardBootstrap(argsBaseBootstrapper)
	if err != nil {
		return nil, err
	}

	return sync.NewSovereignChainShardBootstrap(bootstrapper)
}

type SovereignShardStorageBootstrapperFactory struct {
}

func (ssbf *SovereignShardStorageBootstrapperFactory) CreateShardStorageBootstrapper(args storageBootstrap.ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error) {
	shardStorageBootstrapper, err := storageBootstrap.NewShardStorageBootstrapper(args)
	if err != nil {
		return nil, err
	}

	return storageBootstrap.NewSovereignChainShardStorageBootstrapper(shardStorageBootstrapper)
}
