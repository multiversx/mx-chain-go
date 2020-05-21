package factory

import (
	"errors"
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/partitioning"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/containers"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/factory/resolverscontainer"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/checking"
	processGenesis "github.com/ElrondNetwork/elrond-go/genesis/process"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/pendingMb"
	"github.com/ElrondNetwork/elrond-go/process/block/poolsCleaner"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory/interceptorscontainer"
	"github.com/ElrondNetwork/elrond-go/process/headerCheck"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/ElrondNetwork/elrond-go/process/transactionLog"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
)

// TODO: check underlying components if there are goroutines with infinite for loops

var log = logger.GetOrCreate("factory")

// timeSpanForBadHeaders is the expiry time for an added block header hash
var timeSpanForBadHeaders = time.Minute * 2

// processComponents struct holds the process components
type processComponents struct {
	InterceptorsContainer    process.InterceptorsContainer
	ResolversFinder          dataRetriever.ResolversFinder
	Rounder                  consensus.Rounder
	EpochStartTrigger        epochStart.TriggerHandler
	ForkDetector             process.ForkDetector
	BlockProcessor           process.BlockProcessor
	BlackListHandler         process.BlackListHandler
	BootStorer               process.BootStorer
	HeaderSigVerifier        process.InterceptedHeaderSigVerifier
	ValidatorsStatistics     process.ValidatorStatisticsProcessor
	ValidatorsProvider       process.ValidatorsProvider
	BlockTracker             process.BlockTracker
	PendingMiniBlocksHandler process.PendingMiniBlocksHandler
	RequestHandler           process.RequestHandler
	TxLogsProcessor          process.TransactionLogProcessorDatabase
}

// ProcessComponentsFactoryArgs holds the arguments needed to create a process components factory
type ProcessComponentsFactoryArgs struct {
	CoreComponents            *CoreComponentsFactoryArgs
	AccountsParser            genesis.AccountsParser
	SmartContractParser       genesis.InitialSmartContractParser
	EconomicsData             *economics.EconomicsData
	NodesConfig               *sharding.NodesSetup
	GasSchedule               map[string]map[string]uint64
	Rounder                   consensus.Rounder
	ShardCoordinator          sharding.Coordinator
	NodesCoordinator          sharding.NodesCoordinator
	Data                      *DataComponents
	CoreData                  CoreComponentsHolder
	Crypto                    CryptoComponentsHolder
	State                     *StateComponents
	Network                   NetworkComponentsHolder
	Tries                     *TriesComponents
	CoreServiceContainer      serviceContainer.Core
	RequestedItemsHandler     dataRetriever.RequestedItemsHandler
	WhiteListHandler          process.WhiteListHandler
	WhiteListerVerifiedTxs    process.WhiteListHandler
	EpochStartNotifier        EpochStartNotifier
	EpochStart                *config.EpochStartConfig
	Rater                     sharding.PeerAccountListAndRatingHandler
	RatingsData               process.RatingsInfoHandler
	StartEpochNum             uint32
	SizeCheckDelta            uint32
	StateCheckpointModulus    uint
	MaxComputableRounds       uint64
	NumConcurrentResolverJobs int32
	MinSizeInBytes            uint32
	MaxSizeInBytes            uint32
	MaxRating                 uint32
	ValidatorPubkeyConverter  state.PubkeyConverter
	SystemSCConfig            *config.SystemSmartContractsConfig
	TxLogsProcessor           process.TransactionLogProcessor
	Version                   string
}

type processComponentsFactory struct {
	coreComponents            *CoreComponentsFactoryArgs
	accountsParser            genesis.AccountsParser
	smartContractParser       genesis.InitialSmartContractParser
	economicsData             *economics.EconomicsData
	nodesConfig               *sharding.NodesSetup
	gasSchedule               map[string]map[string]uint64
	rounder                   consensus.Rounder
	shardCoordinator          sharding.Coordinator
	nodesCoordinator          sharding.NodesCoordinator
	data                      *DataComponents
	coreData                  CoreComponentsHolder
	crypto                    CryptoComponentsHolder
	state                     *StateComponents
	network                   NetworkComponentsHolder
	tries                     *TriesComponents
	coreServiceContainer      serviceContainer.Core
	requestedItemsHandler     dataRetriever.RequestedItemsHandler
	whiteListHandler          process.WhiteListHandler
	whiteListerVerifiedTxs    process.WhiteListHandler
	epochStartNotifier        EpochStartNotifier
	startEpochNum             uint32
	rater                     sharding.PeerAccountListAndRatingHandler
	sizeCheckDelta            uint32
	stateCheckpointModulus    uint
	maxComputableRounds       uint64
	numConcurrentResolverJobs int32
	minSizeInBytes            uint32
	maxSizeInBytes            uint32
	maxRating                 uint32
	validatorPubkeyConverter  state.PubkeyConverter
	ratingsData               process.RatingsInfoHandler
	systemSCConfig            *config.SystemSmartContractsConfig
	txLogsProcessor           process.TransactionLogProcessor
	version                   string
}

// NewProcessComponentsFactory will return a new instance of processComponentsFactory
func NewProcessComponentsFactory(args ProcessComponentsFactoryArgs) (*processComponentsFactory, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &processComponentsFactory{
		coreComponents:            args.CoreComponents,
		accountsParser:            args.AccountsParser,
		smartContractParser:       args.SmartContractParser,
		economicsData:             args.EconomicsData,
		nodesConfig:               args.NodesConfig,
		gasSchedule:               args.GasSchedule,
		rounder:                   args.Rounder,
		shardCoordinator:          args.ShardCoordinator,
		nodesCoordinator:          args.NodesCoordinator,
		data:                      args.Data,
		coreData:                  args.CoreData,
		crypto:                    args.Crypto,
		state:                     args.State,
		network:                   args.Network,
		tries:                     args.Tries,
		coreServiceContainer:      args.CoreServiceContainer,
		requestedItemsHandler:     args.RequestedItemsHandler,
		whiteListHandler:          args.WhiteListHandler,
		whiteListerVerifiedTxs:    args.WhiteListerVerifiedTxs,
		epochStartNotifier:        args.EpochStartNotifier,
		rater:                     args.Rater,
		ratingsData:               args.RatingsData,
		sizeCheckDelta:            args.SizeCheckDelta,
		stateCheckpointModulus:    args.StateCheckpointModulus,
		startEpochNum:             args.StartEpochNum,
		maxComputableRounds:       args.MaxComputableRounds,
		numConcurrentResolverJobs: args.NumConcurrentResolverJobs,
		minSizeInBytes:            args.MinSizeInBytes,
		maxSizeInBytes:            args.MaxSizeInBytes,
		maxRating:                 args.MaxRating,
		validatorPubkeyConverter:  args.ValidatorPubkeyConverter,
		systemSCConfig:            args.SystemSCConfig,
		version:                   args.Version,
	}, nil
}

// Create will create and return a struct containing process components
func (pcf *processComponentsFactory) Create() (*processComponents, error) {
	argsHeaderSig := &headerCheck.ArgsHeaderSigVerifier{
		Marshalizer:       pcf.coreData.InternalMarshalizer(),
		Hasher:            pcf.coreData.Hasher(),
		NodesCoordinator:  pcf.nodesCoordinator,
		MultiSigVerifier:  pcf.crypto.MultiSigner(),
		SingleSigVerifier: pcf.crypto.TxSingleSigner(),
		KeyGen:            pcf.crypto.BlockSignKeyGen(),
	}
	headerSigVerifier, err := headerCheck.NewHeaderSigVerifier(argsHeaderSig)
	if err != nil {
		return nil, err
	}

	resolversContainerFactory, err := pcf.newResolverContainerFactory()
	if err != nil {
		return nil, err
	}

	resolversContainer, err := resolversContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	resolversFinder, err := containers.NewResolversFinder(resolversContainer, pcf.shardCoordinator)
	if err != nil {
		return nil, err
	}

	requestHandler, err := requestHandlers.NewResolverRequestHandler(
		resolversFinder,
		pcf.requestedItemsHandler,
		pcf.whiteListHandler,
		core.MaxTxsToRequest,
		pcf.shardCoordinator.SelfId(),
		time.Second,
	)
	if err != nil {
		return nil, err
	}

	validatorStatisticsProcessor, err := pcf.newValidatorStatisticsProcessor()
	if err != nil {
		return nil, err
	}

	validatorsProvider, err := peer.NewValidatorsProvider(
		validatorStatisticsProcessor,
		pcf.maxRating,
		pcf.validatorPubkeyConverter,
	)
	if err != nil {
		return nil, err
	}

	epochStartTrigger, err := pcf.newEpochStartTrigger(requestHandler)
	if err != nil {
		return nil, err
	}

	requestHandler.SetEpoch(epochStartTrigger.Epoch())

	err = dataRetriever.SetEpochHandlerToHdrResolver(resolversContainer, epochStartTrigger)
	if err != nil {
		return nil, err
	}

	validatorStatsRootHash, err := validatorStatisticsProcessor.RootHash()
	if err != nil {
		return nil, err
	}

	log.Trace("Validator stats created", "validatorStatsRootHash", validatorStatsRootHash)

	txLogsStorage := pcf.data.Store.GetStorer(dataRetriever.TxLogsUnit)
	txLogsProcessor, err := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      txLogsStorage,
		Marshalizer: pcf.coreData.InternalMarshalizer(),
	})
	if err != nil {
		return nil, err
	}

	pcf.txLogsProcessor = txLogsProcessor
	genesisBlocks, err := pcf.generateGenesisHeadersAndApplyInitialBalances()
	if err != nil {
		return nil, err
	}

	err = pcf.prepareGenesisBlock(genesisBlocks)
	if err != nil {
		return nil, err
	}

	bootStr := pcf.data.Store.GetStorer(dataRetriever.BootstrapUnit)
	bootStorer, err := bootstrapStorage.NewBootstrapStorer(pcf.coreData.InternalMarshalizer(), bootStr)
	if err != nil {
		return nil, err
	}

	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      pcf.coreData.Hasher(),
		Marshalizer: pcf.coreData.InternalMarshalizer(),
	}
	headerValidator, err := block.NewHeaderValidator(argsHeaderValidator)
	if err != nil {
		return nil, err
	}

	blockTracker, err := pcf.newBlockTracker(
		headerValidator,
		requestHandler,
		genesisBlocks,
	)
	if err != nil {
		return nil, err
	}

	_, err = poolsCleaner.NewMiniBlocksPoolsCleaner(
		pcf.data.Datapool.MiniBlocks(),
		pcf.rounder,
		pcf.shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	_, err = poolsCleaner.NewTxsPoolsCleaner(
		pcf.state.AddressPubkeyConverter,
		pcf.data.Datapool,
		pcf.rounder,
		pcf.shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	interceptorContainerFactory, blackListHandler, err := pcf.newInterceptorContainerFactory(
		headerSigVerifier,
		blockTracker,
		epochStartTrigger,
	)
	if err != nil {
		return nil, err
	}

	//TODO refactor all these factory calls
	interceptorsContainer, err := interceptorContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	var pendingMiniBlocksHandler process.PendingMiniBlocksHandler
	if pcf.shardCoordinator.SelfId() == core.MetachainShardId {
		pendingMiniBlocksHandler, err = pendingMb.NewPendingMiniBlocks()
		if err != nil {
			return nil, err
		}
	}

	forkDetector, err := pcf.newForkDetector(blackListHandler, blockTracker)
	if err != nil {
		return nil, err
	}

	blockProcessor, err := pcf.newBlockProcessor(
		requestHandler,
		forkDetector,
		epochStartTrigger,
		bootStorer,
		validatorStatisticsProcessor,
		headerValidator,
		blockTracker,
		pendingMiniBlocksHandler,
	)
	if err != nil {
		return nil, err
	}

	nodesSetupChecker, err := checking.NewNodesSetupChecker(
		pcf.accountsParser,
		pcf.economicsData.GenesisNodePrice(),
		pcf.validatorPubkeyConverter,
	)
	if err != nil {
		return nil, err
	}

	err = nodesSetupChecker.Check(pcf.nodesConfig.AllInitialNodes())
	if err != nil {
		return nil, err
	}

	return &processComponents{
		InterceptorsContainer:    interceptorsContainer,
		ResolversFinder:          resolversFinder,
		Rounder:                  pcf.rounder,
		ForkDetector:             forkDetector,
		BlockProcessor:           blockProcessor,
		EpochStartTrigger:        epochStartTrigger,
		BlackListHandler:         blackListHandler,
		BootStorer:               bootStorer,
		HeaderSigVerifier:        headerSigVerifier,
		ValidatorsStatistics:     validatorStatisticsProcessor,
		ValidatorsProvider:       validatorsProvider,
		BlockTracker:             blockTracker,
		PendingMiniBlocksHandler: pendingMiniBlocksHandler,
		RequestHandler:           requestHandler,
		TxLogsProcessor:          txLogsProcessor,
	}, nil
}

func (pcf *processComponentsFactory) newValidatorStatisticsProcessor() (process.ValidatorStatisticsProcessor, error) {

	storageService := pcf.data.Store

	var peerDataPool peer.DataPool = pcf.data.Datapool
	if pcf.shardCoordinator.SelfId() < pcf.shardCoordinator.NumberOfShards() {
		peerDataPool = pcf.data.Datapool
	}

	hardForkConfig := pcf.coreComponents.Config.Hardfork
	ratingEnabledEpoch := uint32(0)
	if hardForkConfig.MustImport {
		ratingEnabledEpoch = hardForkConfig.StartEpoch + hardForkConfig.ValidatorGracePeriodInEpochs
	}
	arguments := peer.ArgValidatorStatisticsProcessor{
		PeerAdapter:         pcf.state.PeerAccounts,
		PubkeyConv:          pcf.state.ValidatorPubkeyConverter,
		NodesCoordinator:    pcf.nodesCoordinator,
		ShardCoordinator:    pcf.shardCoordinator,
		DataPool:            peerDataPool,
		StorageService:      storageService,
		Marshalizer:         pcf.coreData.InternalMarshalizer(),
		Rater:               pcf.rater,
		MaxComputableRounds: pcf.maxComputableRounds,
		RewardsHandler:      pcf.economicsData,
		NodesSetup:          pcf.nodesConfig,
		RatingEnableEpoch:   ratingEnabledEpoch,
	}

	validatorStatisticsProcessor, err := peer.NewValidatorStatisticsProcessor(arguments)
	if err != nil {
		return nil, err
	}

	return validatorStatisticsProcessor, nil
}

func (pcf *processComponentsFactory) newEpochStartTrigger(requestHandler process.RequestHandler) (epochStart.TriggerHandler, error) {
	if pcf.shardCoordinator.SelfId() < pcf.shardCoordinator.NumberOfShards() {
		argsHeaderValidator := block.ArgsHeaderValidator{
			Hasher:      pcf.coreData.Hasher(),
			Marshalizer: pcf.coreData.InternalMarshalizer(),
		}
		headerValidator, err := block.NewHeaderValidator(argsHeaderValidator)
		if err != nil {
			return nil, err
		}

		argsPeerMiniBlockSyncer := shardchain.ArgPeerMiniBlockSyncer{
			MiniBlocksPool: pcf.data.Datapool.MiniBlocks(),
			Requesthandler: requestHandler,
		}

		peerMiniBlockSyncer, err := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlockSyncer)
		if err != nil {
			return nil, err
		}

		argEpochStart := &shardchain.ArgsShardEpochStartTrigger{
			Marshalizer:          pcf.coreData.InternalMarshalizer(),
			Hasher:               pcf.coreData.Hasher(),
			HeaderValidator:      headerValidator,
			Uint64Converter:      pcf.coreData.Uint64ByteSliceConverter(),
			DataPool:             pcf.data.Datapool,
			Storage:              pcf.data.Store,
			RequestHandler:       requestHandler,
			Epoch:                pcf.startEpochNum,
			EpochStartNotifier:   pcf.epochStartNotifier,
			Validity:             process.MetaBlockValidity,
			Finality:             process.BlockFinality,
			PeerMiniBlocksSyncer: peerMiniBlockSyncer,
		}
		epochStartTrigger, err := shardchain.NewEpochStartTrigger(argEpochStart)
		if err != nil {
			return nil, errors.New("error creating new start of epoch trigger" + err.Error())
		}
		err = epochStartTrigger.SetAppStatusHandler(pcf.coreData.StatusHandler())
		if err != nil {
			return nil, err
		}

		return epochStartTrigger, nil
	}

	if pcf.shardCoordinator.SelfId() == core.MetachainShardId {
		argEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
			GenesisTime:        time.Unix(pcf.nodesConfig.StartTime, 0),
			Settings:           &pcf.coreComponents.Config.EpochStartConfig,
			Epoch:              pcf.startEpochNum,
			EpochStartNotifier: pcf.epochStartNotifier,
			Storage:            pcf.data.Store,
			Marshalizer:        pcf.coreData.InternalMarshalizer(),
			Hasher:             pcf.coreData.Hasher(),
		}
		epochStartTrigger, err := metachain.NewEpochStartTrigger(argEpochStart)
		if err != nil {
			return nil, errors.New("error creating new start of epoch trigger" + err.Error())
		}
		err = epochStartTrigger.SetAppStatusHandler(pcf.coreData.StatusHandler())
		if err != nil {
			return nil, err
		}

		return epochStartTrigger, nil
	}

	return nil, errors.New("error creating new start of epoch trigger because of invalid shard id")
}

func (pcf *processComponentsFactory) generateGenesisHeadersAndApplyInitialBalances() (map[uint32]data.HeaderHandler, error) {
	arg := processGenesis.ArgsGenesisBlockCreator{
		GenesisTime:              uint64(pcf.nodesConfig.StartTime),
		StartEpochNum:            pcf.startEpochNum,
		Accounts:                 pcf.state.AccountsAdapter,
		PubkeyConv:               pcf.state.AddressPubkeyConverter,
		InitialNodesSetup:        pcf.nodesConfig,
		Economics:                pcf.economicsData,
		ShardCoordinator:         pcf.shardCoordinator,
		Store:                    pcf.data.Store,
		Blkc:                     pcf.data.Blkc,
		Marshalizer:              pcf.coreData.InternalMarshalizer(),
		Hasher:                   pcf.coreData.Hasher(),
		Uint64ByteSliceConverter: pcf.coreData.Uint64ByteSliceConverter(),
		DataPool:                 pcf.data.Datapool,
		AccountsParser:           pcf.accountsParser,
		SmartContractParser:      pcf.smartContractParser,
		ValidatorAccounts:        pcf.state.PeerAccounts,
		GasMap:                   pcf.gasSchedule,
		VirtualMachineConfig:     pcf.coreComponents.Config.VirtualMachineConfig,
		TxLogsProcessor:          pcf.txLogsProcessor,
		HardForkConfig:           pcf.coreComponents.Config.Hardfork,
		TrieStorageManagers:      pcf.tries.TrieStorageManagers,
		ChainID:                  pcf.coreData.ChainID(),
		SystemSCConfig:           *pcf.systemSCConfig,
	}

	gbc, err := processGenesis.NewGenesisBlockCreator(arg)
	if err != nil {
		return nil, err
	}

	return gbc.CreateGenesisBlocks()
}

func (pcf *processComponentsFactory) prepareGenesisBlock(genesisBlocks map[uint32]data.HeaderHandler) error {
	genesisBlock, ok := genesisBlocks[pcf.shardCoordinator.SelfId()]
	if !ok {
		return errors.New("genesis block does not exists")
	}

	genesisBlockHash, err := core.CalculateHash(pcf.coreData.InternalMarshalizer(), pcf.coreData.Hasher(), genesisBlock)
	if err != nil {
		return err
	}

	err = pcf.data.Blkc.SetGenesisHeader(genesisBlock)
	if err != nil {
		return err
	}

	pcf.data.Blkc.SetGenesisHeaderHash(genesisBlockHash)

	marshalizedBlock, err := pcf.coreData.InternalMarshalizer().Marshal(genesisBlock)
	if err != nil {
		return err
	}

	if pcf.shardCoordinator.SelfId() == core.MetachainShardId {
		errNotCritical := pcf.data.Store.Put(dataRetriever.MetaBlockUnit, genesisBlockHash, marshalizedBlock)
		if errNotCritical != nil {
			log.Error("error storing genesis metablock", "error", errNotCritical.Error())
		}
	} else {
		errNotCritical := pcf.data.Store.Put(dataRetriever.BlockHeaderUnit, genesisBlockHash, marshalizedBlock)
		if errNotCritical != nil {
			log.Error("error storing genesis shardblock", "error", errNotCritical.Error())
		}
	}

	return nil
}

func (pcf *processComponentsFactory) newBlockTracker(
	headerValidator process.HeaderConstructionValidator,
	requestHandler process.RequestHandler,
	genesisBlocks map[uint32]data.HeaderHandler,
) (process.BlockTracker, error) {
	argBaseTracker := track.ArgBaseTracker{
		Hasher:           pcf.coreData.Hasher(),
		HeaderValidator:  headerValidator,
		Marshalizer:      pcf.coreData.InternalMarshalizer(),
		RequestHandler:   requestHandler,
		Rounder:          pcf.rounder,
		ShardCoordinator: pcf.shardCoordinator,
		Store:            pcf.data.Store,
		StartHeaders:     genesisBlocks,
		PoolsHolder:      pcf.data.Datapool,
		WhitelistHandler: pcf.whiteListHandler,
	}

	if pcf.shardCoordinator.SelfId() < pcf.shardCoordinator.NumberOfShards() {
		arguments := track.ArgShardTracker{
			ArgBaseTracker: argBaseTracker,
		}

		return track.NewShardBlockTrack(arguments)
	}

	if pcf.shardCoordinator.SelfId() == core.MetachainShardId {
		arguments := track.ArgMetaTracker{
			ArgBaseTracker: argBaseTracker,
		}

		return track.NewMetaBlockTrack(arguments)
	}

	return nil, errors.New("could not create block tracker")
}

// -- Resolvers container Factory begin
func (pcf *processComponentsFactory) newResolverContainerFactory() (dataRetriever.ResolversContainerFactory, error) {
	if pcf.shardCoordinator.SelfId() < pcf.shardCoordinator.NumberOfShards() {
		return pcf.newShardResolverContainerFactory()
	}
	if pcf.shardCoordinator.SelfId() == core.MetachainShardId {
		return pcf.newMetaResolverContainerFactory()
	}

	return nil, errors.New("could not create interceptor and resolver container factory")
}

func (pcf *processComponentsFactory) newShardResolverContainerFactory() (dataRetriever.ResolversContainerFactory, error) {

	dataPacker, err := partitioning.NewSimpleDataPacker(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:           pcf.shardCoordinator,
		Messenger:                  pcf.network.NetworkMessenger(),
		Store:                      pcf.data.Store,
		Marshalizer:                pcf.coreData.InternalMarshalizer(),
		DataPools:                  pcf.data.Datapool,
		Uint64ByteSliceConverter:   pcf.coreData.Uint64ByteSliceConverter(),
		DataPacker:                 dataPacker,
		TriesContainer:             pcf.tries.TriesContainer,
		SizeCheckDelta:             pcf.sizeCheckDelta,
		InputAntifloodHandler:      pcf.network.InputAntiFloodHandler(),
		OutputAntifloodHandler:     pcf.network.OutputAntiFloodHandler(),
		NumConcurrentResolvingJobs: pcf.numConcurrentResolverJobs,
	}
	resolversContainerFactory, err := resolverscontainer.NewShardResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}

	return resolversContainerFactory, nil
}

func (pcf *processComponentsFactory) newMetaResolverContainerFactory() (dataRetriever.ResolversContainerFactory, error) {
	dataPacker, err := partitioning.NewSimpleDataPacker(pcf.coreData.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	resolversContainerFactoryArgs := resolverscontainer.FactoryArgs{
		ShardCoordinator:           pcf.shardCoordinator,
		Messenger:                  pcf.network.NetworkMessenger(),
		Store:                      pcf.data.Store,
		Marshalizer:                pcf.coreData.InternalMarshalizer(),
		DataPools:                  pcf.data.Datapool,
		Uint64ByteSliceConverter:   pcf.coreData.Uint64ByteSliceConverter(),
		DataPacker:                 dataPacker,
		TriesContainer:             pcf.tries.TriesContainer,
		SizeCheckDelta:             pcf.sizeCheckDelta,
		InputAntifloodHandler:      pcf.network.InputAntiFloodHandler(),
		OutputAntifloodHandler:     pcf.network.OutputAntiFloodHandler(),
		NumConcurrentResolvingJobs: pcf.numConcurrentResolverJobs,
	}
	resolversContainerFactory, err := resolverscontainer.NewMetaResolversContainerFactory(resolversContainerFactoryArgs)
	if err != nil {
		return nil, err
	}
	return resolversContainerFactory, nil
}

// -- Resolvers container Factory begin

// -- Interceptors containers start

func (pcf *processComponentsFactory) newInterceptorContainerFactory(
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {
	if pcf.shardCoordinator.SelfId() < pcf.shardCoordinator.NumberOfShards() {
		return pcf.newShardInterceptorContainerFactory(
			headerSigVerifier,
			validityAttester,
			epochStartTrigger,
		)
	}
	if pcf.shardCoordinator.SelfId() == core.MetachainShardId {
		return pcf.newMetaInterceptorContainerFactory(
			headerSigVerifier,
			validityAttester,
			epochStartTrigger,
		)
	}

	return nil, nil, errors.New("could not create interceptor container factory")
}

func (pcf *processComponentsFactory) newShardInterceptorContainerFactory(
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {
	headerBlackList := timecache.NewTimeCache(timeSpanForBadHeaders)
	shardInterceptorsContainerFactoryArgs := interceptorscontainer.ShardInterceptorsContainerFactoryArgs{
		Accounts:               pcf.state.AccountsAdapter,
		ShardCoordinator:       pcf.shardCoordinator,
		NodesCoordinator:       pcf.nodesCoordinator,
		Messenger:              pcf.network.NetworkMessenger(),
		Store:                  pcf.data.Store,
		ProtoMarshalizer:       pcf.coreData.InternalMarshalizer(),
		TxSignMarshalizer:      pcf.coreData.TxMarshalizer(),
		Hasher:                 pcf.coreData.Hasher(),
		KeyGen:                 pcf.crypto.TxSignKeyGen(),
		BlockSignKeyGen:        pcf.crypto.BlockSignKeyGen(),
		SingleSigner:           pcf.crypto.TxSingleSigner(),
		BlockSingleSigner:      pcf.crypto.BlockSigner(),
		MultiSigner:            pcf.crypto.MultiSigner(),
		DataPool:               pcf.data.Datapool,
		AddressPubkeyConverter: pcf.state.AddressPubkeyConverter,
		MaxTxNonceDeltaAllowed: core.MaxTxNonceDeltaAllowed,
		TxFeeHandler:           pcf.economicsData,
		BlackList:              headerBlackList,
		HeaderSigVerifier:      headerSigVerifier,
		ChainID:                []byte(pcf.coreData.ChainID()),
		SizeCheckDelta:         pcf.sizeCheckDelta,
		ValidityAttester:       validityAttester,
		EpochStartTrigger:      epochStartTrigger,
		WhiteListHandler:       pcf.whiteListHandler,
		WhiteListerVerifiedTxs: pcf.whiteListerVerifiedTxs,
		AntifloodHandler:       pcf.network.InputAntiFloodHandler(),
		NonceConverter:         pcf.coreData.Uint64ByteSliceConverter(),
	}
	interceptorContainerFactory, err := interceptorscontainer.NewShardInterceptorsContainerFactory(shardInterceptorsContainerFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	return interceptorContainerFactory, headerBlackList, nil
}

func (pcf *processComponentsFactory) newMetaInterceptorContainerFactory(
	headerSigVerifier process.InterceptedHeaderSigVerifier,
	validityAttester process.ValidityAttester,
	epochStartTrigger process.EpochStartTriggerHandler,
) (process.InterceptorsContainerFactory, process.BlackListHandler, error) {
	headerBlackList := timecache.NewTimeCache(timeSpanForBadHeaders)
	metaInterceptorsContainerFactoryArgs := interceptorscontainer.MetaInterceptorsContainerFactoryArgs{
		ShardCoordinator:       pcf.shardCoordinator,
		NodesCoordinator:       pcf.nodesCoordinator,
		Messenger:              pcf.network.NetworkMessenger(),
		Store:                  pcf.data.Store,
		ProtoMarshalizer:       pcf.coreData.InternalMarshalizer(),
		TxSignMarshalizer:      pcf.coreData.TxMarshalizer(),
		Hasher:                 pcf.coreData.Hasher(),
		MultiSigner:            pcf.crypto.MultiSigner(),
		DataPool:               pcf.data.Datapool,
		Accounts:               pcf.state.AccountsAdapter,
		AddressPubkeyConverter: pcf.state.AddressPubkeyConverter,
		SingleSigner:           pcf.crypto.TxSingleSigner(),
		BlockSingleSigner:      pcf.crypto.BlockSigner(),
		KeyGen:                 pcf.crypto.TxSignKeyGen(),
		BlockKeyGen:            pcf.crypto.BlockSignKeyGen(),
		MaxTxNonceDeltaAllowed: core.MaxTxNonceDeltaAllowed,
		TxFeeHandler:           pcf.economicsData,
		BlackList:              headerBlackList,
		HeaderSigVerifier:      headerSigVerifier,
		ChainID:                []byte(pcf.coreData.ChainID()),
		SizeCheckDelta:         pcf.sizeCheckDelta,
		ValidityAttester:       validityAttester,
		EpochStartTrigger:      epochStartTrigger,
		WhiteListHandler:       pcf.whiteListHandler,
		WhiteListerVerifiedTxs: pcf.whiteListerVerifiedTxs,
		AntifloodHandler:       pcf.network.InputAntiFloodHandler(),
		NonceConverter:         pcf.coreData.Uint64ByteSliceConverter(),
	}
	interceptorContainerFactory, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(metaInterceptorsContainerFactoryArgs)
	if err != nil {
		return nil, nil, err
	}

	return interceptorContainerFactory, headerBlackList, nil
}

// -- Interceptors containers end

func (pcf *processComponentsFactory) newForkDetector(
	headerBlackList process.BlackListHandler,
	blockTracker process.BlockTracker,
) (process.ForkDetector, error) {
	if pcf.shardCoordinator.SelfId() < pcf.shardCoordinator.NumberOfShards() {
		return sync.NewShardForkDetector(pcf.rounder, headerBlackList, blockTracker, pcf.nodesConfig.StartTime)
	}
	if pcf.shardCoordinator.SelfId() == core.MetachainShardId {
		return sync.NewMetaForkDetector(pcf.rounder, headerBlackList, blockTracker, pcf.nodesConfig.StartTime)
	}

	return nil, errors.New("could not create fork detector")
}

func checkArgs(args ProcessComponentsFactoryArgs) error {
	baseErrMessage := "error creating process components"
	if args.CoreComponents == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilCoreComponents)
	}
	if check.IfNil(args.AccountsParser) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilAccountsParser)
	}
	if check.IfNil(args.SmartContractParser) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilSmartContractParser)
	}
	if args.EconomicsData == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilEconomicsData)
	}
	if args.NodesConfig == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilNodesConfig)
	}
	if args.GasSchedule == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilGasSchedule)
	}
	if check.IfNil(args.Rounder) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilRounder)
	}
	if check.IfNil(args.ShardCoordinator) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilShardCoordinator)
	}
	if check.IfNil(args.NodesCoordinator) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilNodesCoordinator)
	}
	if args.Data == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilDataComponents)
	}
	if check.IfNil(args.CoreData) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilCoreComponentsHolder)
	}
	if check.IfNil(args.Crypto) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilCryptoComponentsHolder)
	}
	if args.State == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilStateComponents)
	}
	if check.IfNil(args.Network) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilNetworkComponentsHolder)
	}
	if args.Tries == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilTriesComponents)
	}
	if check.IfNil(args.CoreServiceContainer) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilCoreServiceContainer)
	}
	if check.IfNil(args.RequestedItemsHandler) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilRequestedItemHandler)
	}
	if check.IfNil(args.WhiteListHandler) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilWhiteListHandler)
	}
	if check.IfNil(args.WhiteListerVerifiedTxs) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilWhiteListVerifiedTxs)
	}
	if check.IfNil(args.EpochStartNotifier) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilEpochStartNotifier)
	}
	if args.EpochStart == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilEpochStartConfig)
	}
	if check.IfNil(args.EpochStartNotifier) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilEpochStartNotifier)
	}
	if check.IfNil(args.Rater) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilRater)
	}
	if check.IfNil(args.RatingsData) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilRatingData)
	}
	if check.IfNil(args.ValidatorPubkeyConverter) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilPubKeyConverter)
	}
	if args.SystemSCConfig == nil {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilSystemSCConfig)
	}
	if check.IfNil(args.TxLogsProcessor) {
		return fmt.Errorf("%s: %w", baseErrMessage, ErrNilTxLogsConfiguration)
	}
	return nil
}
