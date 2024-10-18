package factory

import (
	"fmt"
	"math"
	"os"
	"path"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/debug/factory"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/epochStart/shardchain"
	mxFactory "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/genesis"
	"github.com/multiversx/mx-chain-go/update/storing"
	"github.com/multiversx/mx-chain-go/update/sync"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("update/factory")

// ArgsExporter is the argument structure to create a new exporter
type ArgsExporter struct {
	CoreComponents                   process.CoreComponentsHolder
	CryptoComponents                 process.CryptoComponentsHolder
	StatusCoreComponents             process.StatusCoreComponentsHolder
	NetworkComponents                mxFactory.NetworkComponentsHolder
	HeaderValidator                  epochStart.HeaderValidator
	DataPool                         dataRetriever.PoolsHolder
	StorageService                   dataRetriever.StorageService
	RequestHandler                   process.RequestHandler
	ShardCoordinator                 sharding.Coordinator
	ActiveAccountsDBs                map[state.AccountsDbIdentifier]state.AccountsAdapter
	ExistingResolvers                dataRetriever.ResolversContainer
	ExistingRequesters               dataRetriever.RequestersContainer
	ExportFolder                     string
	ExportTriesStorageConfig         config.StorageConfig
	ExportStateStorageConfig         config.StorageConfig
	ExportStateKeysConfig            config.StorageConfig
	MaxTrieLevelInMemory             uint
	WhiteListHandler                 process.WhiteListHandler
	WhiteListerVerifiedTxs           process.WhiteListHandler
	MainInterceptorsContainer        process.InterceptorsContainer
	FullArchiveInterceptorsContainer process.InterceptorsContainer
	NodesCoordinator                 nodesCoordinator.NodesCoordinator
	HeaderSigVerifier                process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier          process.HeaderIntegrityVerifier
	ValidityAttester                 process.ValidityAttester
	RoundHandler                     process.RoundHandler
	InterceptorDebugConfig           config.InterceptorResolverDebugConfig
	MaxHardCapForMissingNodes        int
	NumConcurrentTrieSyncers         int
	TrieSyncerVersion                int
	CheckNodesOnDisk                 bool
	NodeOperationMode                common.NodeOperation
}

type exportHandlerFactory struct {
	coreComponents                   process.CoreComponentsHolder
	cryptoComponents                 process.CryptoComponentsHolder
	statusCoreComponents             process.StatusCoreComponentsHolder
	networkComponents                mxFactory.NetworkComponentsHolder
	headerValidator                  epochStart.HeaderValidator
	dataPool                         dataRetriever.PoolsHolder
	storageService                   dataRetriever.StorageService
	requestHandler                   process.RequestHandler
	shardCoordinator                 sharding.Coordinator
	activeAccountsDBs                map[state.AccountsDbIdentifier]state.AccountsAdapter
	exportFolder                     string
	exportTriesStorageConfig         config.StorageConfig
	exportStateStorageConfig         config.StorageConfig
	exportStateKeysConfig            config.StorageConfig
	maxTrieLevelInMemory             uint
	whiteListHandler                 process.WhiteListHandler
	whiteListerVerifiedTxs           process.WhiteListHandler
	mainInterceptorsContainer        process.InterceptorsContainer
	fullArchiveInterceptorsContainer process.InterceptorsContainer
	existingResolvers                dataRetriever.ResolversContainer
	existingRequesters               dataRetriever.RequestersContainer
	epochStartTrigger                epochStart.TriggerHandler
	accounts                         state.AccountsAdapter
	nodesCoordinator                 nodesCoordinator.NodesCoordinator
	headerSigVerifier                process.InterceptedHeaderSigVerifier
	headerIntegrityVerifier          process.HeaderIntegrityVerifier
	validityAttester                 process.ValidityAttester
	resolverContainer                dataRetriever.ResolversContainer
	requestersContainer              dataRetriever.RequestersContainer
	roundHandler                     process.RoundHandler
	interceptorDebugConfig           config.InterceptorResolverDebugConfig
	maxHardCapForMissingNodes        int
	numConcurrentTrieSyncers         int
	trieSyncerVersion                int
	checkNodesOnDisk                 bool
	nodeOperationMode                common.NodeOperation
}

// NewExportHandlerFactory creates an exporter factory
func NewExportHandlerFactory(args ArgsExporter) (*exportHandlerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.CoreComponents) {
		return nil, update.ErrNilCoreComponents
	}
	if check.IfNil(args.NetworkComponents) {
		return nil, update.ErrNilNetworkComponents
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, update.ErrNilCryptoComponents
	}
	if check.IfNil(args.CoreComponents.Hasher()) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.CoreComponents.InternalMarshalizer()) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.HeaderValidator) {
		return nil, update.ErrNilHeaderValidator
	}
	if check.IfNil(args.CoreComponents.Uint64ByteSliceConverter()) {
		return nil, update.ErrNilUint64Converter
	}
	if check.IfNil(args.DataPool) {
		return nil, update.ErrNilDataPoolHolder
	}
	if check.IfNil(args.StorageService) {
		return nil, update.ErrNilStorage
	}
	if check.IfNil(args.RequestHandler) {
		return nil, update.ErrNilRequestHandler
	}
	if check.IfNil(args.NetworkComponents) {
		return nil, update.ErrNilMessenger
	}
	if args.ActiveAccountsDBs == nil {
		return nil, update.ErrNilAccounts
	}
	if check.IfNil(args.WhiteListHandler) {
		return nil, update.ErrNilWhiteListHandler
	}
	if check.IfNil(args.WhiteListerVerifiedTxs) {
		return nil, update.ErrNilWhiteListHandler
	}
	if check.IfNil(args.MainInterceptorsContainer) {
		return nil, fmt.Errorf("%w on main network", update.ErrNilInterceptorsContainer)
	}
	if check.IfNil(args.FullArchiveInterceptorsContainer) {
		return nil, fmt.Errorf("%w on full archive network", update.ErrNilInterceptorsContainer)
	}
	if check.IfNil(args.ExistingResolvers) {
		return nil, update.ErrNilResolverContainer
	}
	if check.IfNil(args.ExistingRequesters) {
		return nil, update.ErrNilRequestersContainer
	}
	multiSigner, err := args.CryptoComponents.GetMultiSigner(0)
	if err != nil {
		return nil, err
	}
	if check.IfNil(multiSigner) {
		return nil, update.ErrNilMultiSigner
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, update.ErrNilNodesCoordinator
	}
	if check.IfNil(args.CryptoComponents.TxSingleSigner()) {
		return nil, update.ErrNilSingleSigner
	}
	if check.IfNil(args.CoreComponents.AddressPubKeyConverter()) {
		return nil, update.ErrNilPubKeyConverter
	}
	if check.IfNil(args.CoreComponents.ValidatorPubKeyConverter()) {
		return nil, update.ErrNilPubKeyConverter
	}
	if check.IfNil(args.CryptoComponents.BlockSignKeyGen()) {
		return nil, update.ErrNilBlockKeyGen
	}
	if check.IfNil(args.CryptoComponents.TxSignKeyGen()) {
		return nil, update.ErrNilKeyGenerator
	}
	if check.IfNil(args.CryptoComponents.BlockSigner()) {
		return nil, update.ErrNilBlockSigner
	}
	if check.IfNil(args.HeaderSigVerifier) {
		return nil, update.ErrNilHeaderSigVerifier
	}
	if check.IfNil(args.HeaderIntegrityVerifier) {
		return nil, update.ErrNilHeaderIntegrityVerifier
	}
	if check.IfNil(args.ValidityAttester) {
		return nil, update.ErrNilValidityAttester
	}
	if check.IfNil(args.CoreComponents.TxMarshalizer()) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.RoundHandler) {
		return nil, update.ErrNilRoundHandler
	}
	if check.IfNil(args.CoreComponents.TxSignHasher()) {
		return nil, update.ErrNilHasher
	}
	if args.MaxHardCapForMissingNodes < 1 {
		return nil, update.ErrInvalidMaxHardCapForMissingNodes
	}
	if args.NumConcurrentTrieSyncers < 1 {
		return nil, update.ErrInvalidNumConcurrentTrieSyncers
	}
	err = trie.CheckTrieSyncerVersion(args.TrieSyncerVersion)
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.StatusCoreComponents) {
		return nil, update.ErrNilStatusCoreComponentsHolder
	}
	if check.IfNil(args.StatusCoreComponents.AppStatusHandler()) {
		return nil, update.ErrNilAppStatusHandler
	}

	e := &exportHandlerFactory{
		coreComponents:                   args.CoreComponents,
		cryptoComponents:                 args.CryptoComponents,
		networkComponents:                args.NetworkComponents,
		headerValidator:                  args.HeaderValidator,
		dataPool:                         args.DataPool,
		storageService:                   args.StorageService,
		requestHandler:                   args.RequestHandler,
		shardCoordinator:                 args.ShardCoordinator,
		activeAccountsDBs:                args.ActiveAccountsDBs,
		exportFolder:                     args.ExportFolder,
		exportTriesStorageConfig:         args.ExportTriesStorageConfig,
		exportStateStorageConfig:         args.ExportStateStorageConfig,
		exportStateKeysConfig:            args.ExportStateKeysConfig,
		mainInterceptorsContainer:        args.MainInterceptorsContainer,
		fullArchiveInterceptorsContainer: args.FullArchiveInterceptorsContainer,
		whiteListHandler:                 args.WhiteListHandler,
		whiteListerVerifiedTxs:           args.WhiteListerVerifiedTxs,
		existingResolvers:                args.ExistingResolvers,
		existingRequesters:               args.ExistingRequesters,
		accounts:                         args.ActiveAccountsDBs[state.UserAccountsState],
		nodesCoordinator:                 args.NodesCoordinator,
		headerSigVerifier:                args.HeaderSigVerifier,
		headerIntegrityVerifier:          args.HeaderIntegrityVerifier,
		validityAttester:                 args.ValidityAttester,
		maxTrieLevelInMemory:             args.MaxTrieLevelInMemory,
		roundHandler:                     args.RoundHandler,
		interceptorDebugConfig:           args.InterceptorDebugConfig,
		maxHardCapForMissingNodes:        args.MaxHardCapForMissingNodes,
		numConcurrentTrieSyncers:         args.NumConcurrentTrieSyncers,
		trieSyncerVersion:                args.TrieSyncerVersion,
		checkNodesOnDisk:                 args.CheckNodesOnDisk,
		statusCoreComponents:             args.StatusCoreComponents,
		nodeOperationMode:                args.NodeOperationMode,
	}

	return e, nil
}

// Create makes a new export handler
func (e *exportHandlerFactory) Create() (update.ExportHandler, error) {
	err := e.prepareFolders(e.exportFolder)
	if err != nil {
		return nil, err
	}

	// TODO reuse the debugger when the one used for regular resolvers & interceptors will be moved inside the status components
	debugger, errNotCritical := factory.NewInterceptorDebuggerFactory(e.interceptorDebugConfig)
	if errNotCritical != nil {
		log.Warn("error creating hardfork debugger", "error", errNotCritical)
	}

	argsPeerMiniBlocksSyncer := shardchain.ArgPeerMiniBlockSyncer{
		MiniBlocksPool:     e.dataPool.MiniBlocks(),
		ValidatorsInfoPool: e.dataPool.ValidatorsInfo(),
		RequestHandler:     e.requestHandler,
	}
	peerMiniBlocksSyncer, err := shardchain.NewPeerMiniBlockSyncer(argsPeerMiniBlocksSyncer)
	if err != nil {
		return nil, err
	}
	argsEpochTrigger := shardchain.ArgsShardEpochStartTrigger{
		Marshalizer:          e.coreComponents.InternalMarshalizer(),
		Hasher:               e.coreComponents.Hasher(),
		HeaderValidator:      e.headerValidator,
		Uint64Converter:      e.coreComponents.Uint64ByteSliceConverter(),
		DataPool:             e.dataPool,
		Storage:              e.storageService,
		RequestHandler:       e.requestHandler,
		EpochStartNotifier:   notifier.NewEpochStartSubscriptionHandler(),
		Epoch:                0,
		Validity:             process.MetaBlockValidity,
		Finality:             process.BlockFinality,
		PeerMiniBlocksSyncer: peerMiniBlocksSyncer,
		RoundHandler:         e.roundHandler,
		AppStatusHandler:     e.statusCoreComponents.AppStatusHandler(),
		EnableEpochsHandler:  e.coreComponents.EnableEpochsHandler(),
	}
	epochHandler, err := shardchain.NewEpochStartTrigger(&argsEpochTrigger)
	if err != nil {
		return nil, err
	}

	argsDataTrieFactory := ArgsNewDataTrieFactory{
		StorageConfig:        e.exportTriesStorageConfig,
		SyncFolder:           e.exportFolder,
		Marshalizer:          e.coreComponents.InternalMarshalizer(),
		Hasher:               e.coreComponents.Hasher(),
		ShardCoordinator:     e.shardCoordinator,
		MaxTrieLevelInMemory: e.maxTrieLevelInMemory,
		EnableEpochsHandler:  e.coreComponents.EnableEpochsHandler(),
		StateStatsCollector:  e.statusCoreComponents.StateStatsHandler(),
	}
	dataTriesContainerFactory, err := NewDataTrieFactory(argsDataTrieFactory)
	if err != nil {
		return nil, err
	}

	trieStorageManager := dataTriesContainerFactory.TrieStorageManager()
	defer func() {
		if err != nil {
			if !check.IfNil(trieStorageManager) {
				_ = trieStorageManager.Close()
			}
		}
	}()

	dataTries, err := dataTriesContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	argsResolvers := ArgsNewResolversContainerFactory{
		ShardCoordinator:           e.shardCoordinator,
		MainMessenger:              e.networkComponents.NetworkMessenger(),
		FullArchiveMessenger:       e.networkComponents.FullArchiveNetworkMessenger(),
		Marshalizer:                e.coreComponents.InternalMarshalizer(),
		DataTrieContainer:          dataTries,
		ExistingResolvers:          e.existingResolvers,
		NumConcurrentResolvingJobs: 100,
		InputAntifloodHandler:      e.networkComponents.InputAntiFloodHandler(),
		OutputAntifloodHandler:     e.networkComponents.OutputAntiFloodHandler(),
	}
	resolversFactory, err := NewResolversContainerFactory(argsResolvers)
	if err != nil {
		return nil, err
	}
	e.resolverContainer, err = resolversFactory.Create()
	if err != nil {
		return nil, err
	}

	e.resolverContainer.Iterate(func(key string, resolver dataRetriever.Resolver) bool {
		errNotCritical = resolver.SetDebugHandler(debugger)
		if errNotCritical != nil {
			log.Warn("error setting debugger", "resolver", key, "error", errNotCritical)
		}

		return true
	})

	argsRequesters := ArgsRequestersContainerFactory{
		ShardCoordinator:       e.shardCoordinator,
		MainMessenger:          e.networkComponents.NetworkMessenger(),
		FullArchiveMessenger:   e.networkComponents.FullArchiveNetworkMessenger(),
		Marshaller:             e.coreComponents.InternalMarshalizer(),
		ExistingRequesters:     e.existingRequesters,
		OutputAntifloodHandler: e.networkComponents.OutputAntiFloodHandler(),
		PeersRatingHandler:     e.networkComponents.PeersRatingHandler(),
	}
	requestersFactory, err := NewRequestersContainerFactory(argsRequesters)
	if err != nil {
		return nil, err
	}
	e.requestersContainer, err = requestersFactory.Create()
	if err != nil {
		return nil, err
	}

	e.requestersContainer.Iterate(func(key string, requester dataRetriever.Requester) bool {
		errNotCritical = requester.SetDebugHandler(debugger)
		if errNotCritical != nil {
			log.Warn("error setting debugger", "requester", key, "error", errNotCritical)
		}

		return true
	})

	argsAccountsSyncers := ArgsNewAccountsDBSyncersContainerFactory{
		TrieCacher:                e.dataPool.TrieNodes(),
		RequestHandler:            e.requestHandler,
		ShardCoordinator:          e.shardCoordinator,
		Hasher:                    e.coreComponents.Hasher(),
		Marshalizer:               e.coreComponents.InternalMarshalizer(),
		TrieStorageManager:        trieStorageManager,
		TimoutGettingTrieNode:     common.TimeoutGettingTrieNodesInHardfork,
		MaxTrieLevelInMemory:      e.maxTrieLevelInMemory,
		MaxHardCapForMissingNodes: e.maxHardCapForMissingNodes,
		NumConcurrentTrieSyncers:  e.numConcurrentTrieSyncers,
		TrieSyncerVersion:         e.trieSyncerVersion,
		CheckNodesOnDisk:          e.checkNodesOnDisk,
		AddressPubKeyConverter:    e.coreComponents.AddressPubKeyConverter(),
		EnableEpochsHandler:       e.coreComponents.EnableEpochsHandler(),
	}
	accountsDBSyncerFactory, err := NewAccountsDBSContainerFactory(argsAccountsSyncers)
	if err != nil {
		return nil, err
	}
	accountsDBSyncerContainer, err := accountsDBSyncerFactory.Create()
	if err != nil {
		return nil, err
	}

	argsNewHeadersSync := sync.ArgsNewHeadersSyncHandler{
		StorageService:   e.storageService,
		Cache:            e.dataPool.Headers(),
		Marshalizer:      e.coreComponents.InternalMarshalizer(),
		Hasher:           e.coreComponents.Hasher(),
		EpochHandler:     epochHandler,
		RequestHandler:   e.requestHandler,
		Uint64Converter:  e.coreComponents.Uint64ByteSliceConverter(),
		ShardCoordinator: e.shardCoordinator,
	}
	epochStartHeadersSyncer, err := sync.NewHeadersSyncHandler(argsNewHeadersSync)
	if err != nil {
		return nil, err
	}

	argsNewSyncAccountsDBsHandler := sync.ArgsNewSyncAccountsDBsHandler{
		AccountsDBsSyncers: accountsDBSyncerContainer,
		ActiveAccountsDBs:  e.activeAccountsDBs,
	}
	epochStartTrieSyncer, err := sync.NewSyncAccountsDBsHandler(argsNewSyncAccountsDBsHandler)
	if err != nil {
		return nil, err
	}

	storer, err := e.storageService.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return nil, err
	}

	argsMiniBlockSyncer := sync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        storer,
		Cache:          e.dataPool.MiniBlocks(),
		Marshalizer:    e.coreComponents.InternalMarshalizer(),
		RequestHandler: e.requestHandler,
	}
	epochStartMiniBlocksSyncer, err := sync.NewPendingMiniBlocksSyncer(argsMiniBlockSyncer)
	if err != nil {
		return nil, err
	}

	argsPendingTransactions := sync.ArgsNewTransactionsSyncer{
		DataPools:      e.dataPool,
		Storages:       e.storageService,
		Marshaller:     e.coreComponents.InternalMarshalizer(),
		RequestHandler: e.requestHandler,
	}
	epochStartTransactionsSyncer, err := sync.NewTransactionsSyncer(argsPendingTransactions)
	if err != nil {
		return nil, err
	}

	argsSyncState := sync.ArgsNewSyncState{
		Headers:      epochStartHeadersSyncer,
		Tries:        epochStartTrieSyncer,
		MiniBlocks:   epochStartMiniBlocksSyncer,
		Transactions: epochStartTransactionsSyncer,
	}
	stateSyncer, err := sync.NewSyncState(argsSyncState)
	if err != nil {
		return nil, err
	}

	var keysStorer storage.Storer
	var keysVals storage.Storer

	defer func() {
		if err != nil {
			if !check.IfNil(keysStorer) {
				_ = keysStorer.Close()
			}
			if !check.IfNil(keysVals) {
				_ = keysVals.Close()
			}
		}
	}()

	keysStorer, err = createStorer(e.exportStateKeysConfig, e.exportFolder)
	if err != nil {
		return nil, fmt.Errorf("%w while creating keys storer", err)
	}
	keysVals, err = createStorer(e.exportStateStorageConfig, e.exportFolder)
	if err != nil {
		return nil, fmt.Errorf("%w while creating keys-values storer", err)
	}

	arg := storing.ArgHardforkStorer{
		KeysStore:   keysStorer,
		KeyValue:    keysVals,
		Marshalizer: e.coreComponents.InternalMarshalizer(),
	}
	hs, err := storing.NewHardforkStorer(arg)
	if err != nil {
		return nil, fmt.Errorf("%w while creating hardfork storer", err)
	}

	argsExporter := genesis.ArgsNewStateExporter{
		ShardCoordinator:         e.shardCoordinator,
		StateSyncer:              stateSyncer,
		Marshalizer:              e.coreComponents.InternalMarshalizer(),
		HardforkStorer:           hs,
		Hasher:                   e.coreComponents.Hasher(),
		ExportFolder:             e.exportFolder,
		ValidatorPubKeyConverter: e.coreComponents.ValidatorPubKeyConverter(),
		AddressPubKeyConverter:   e.coreComponents.AddressPubKeyConverter(),
		GenesisNodesSetupHandler: e.coreComponents.GenesisNodesSetup(),
	}
	exportHandler, err := genesis.NewStateExporter(argsExporter)
	if err != nil {
		return nil, err
	}

	e.epochStartTrigger = epochHandler
	err = e.createInterceptors()
	if err != nil {
		return nil, err
	}

	e.mainInterceptorsContainer.Iterate(func(key string, interceptor process.Interceptor) bool {
		errNotCritical = interceptor.SetInterceptedDebugHandler(debugger)
		if errNotCritical != nil {
			log.Warn("error setting debugger", "interceptor", key, "error", errNotCritical)
		}

		return true
	})

	return exportHandler, nil
}

func (e *exportHandlerFactory) prepareFolders(folder string) error {
	err := os.RemoveAll(folder)
	if err != nil {
		return err
	}

	return os.MkdirAll(folder, os.ModePerm)
}

func (e *exportHandlerFactory) createInterceptors() error {
	argsInterceptors := ArgsNewFullSyncInterceptorsContainerFactory{
		CoreComponents:                   e.coreComponents,
		CryptoComponents:                 e.cryptoComponents,
		Accounts:                         e.accounts,
		ShardCoordinator:                 e.shardCoordinator,
		NodesCoordinator:                 e.nodesCoordinator,
		MainMessenger:                    e.networkComponents.NetworkMessenger(),
		FullArchiveMessenger:             e.networkComponents.FullArchiveNetworkMessenger(),
		Store:                            e.storageService,
		DataPool:                         e.dataPool,
		MaxTxNonceDeltaAllowed:           math.MaxInt32,
		TxFeeHandler:                     &disabled.FeeHandler{},
		BlockBlackList:                   cache.NewTimeCache(time.Second),
		HeaderSigVerifier:                e.headerSigVerifier,
		HeaderIntegrityVerifier:          e.headerIntegrityVerifier,
		SizeCheckDelta:                   math.MaxUint32,
		ValidityAttester:                 e.validityAttester,
		EpochStartTrigger:                e.epochStartTrigger,
		WhiteListHandler:                 e.whiteListHandler,
		WhiteListerVerifiedTxs:           e.whiteListerVerifiedTxs,
		MainInterceptorsContainer:        e.mainInterceptorsContainer,
		FullArchiveInterceptorsContainer: e.fullArchiveInterceptorsContainer,
		AntifloodHandler:                 e.networkComponents.InputAntiFloodHandler(),
		NodeOperationMode:                e.nodeOperationMode,
	}
	fullSyncInterceptors, err := NewFullSyncInterceptorsContainerFactory(argsInterceptors)
	if err != nil {
		return err
	}

	mainInterceptorsContainer, fullArchiveInterceptorsContainer, err := fullSyncInterceptors.Create()
	if err != nil {
		return err
	}

	e.mainInterceptorsContainer = mainInterceptorsContainer
	e.fullArchiveInterceptorsContainer = fullArchiveInterceptorsContainer
	return nil
}

func createStorer(storageConfig config.StorageConfig, folder string) (storage.Storer, error) {
	dbConfig := storageFactory.GetDBFromConfig(storageConfig.DB)
	dbConfig.FilePath = path.Join(folder, storageConfig.DB.FilePath)

	persisterFactory, err := storageFactory.NewPersisterFactory(storageConfig.DB)
	if err != nil {
		return nil, err
	}

	accountsTrieStorage, err := storageunit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(storageConfig.Cache),
		dbConfig,
		persisterFactory,
	)
	if err != nil {
		return nil, err
	}

	return accountsTrieStorage, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *exportHandlerFactory) IsInterfaceNil() bool {
	return e == nil
}
