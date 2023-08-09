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
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/p2p"
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
	CoreComponents            process.CoreComponentsHolder
	CryptoComponents          process.CryptoComponentsHolder
	StatusCoreComponents      process.StatusCoreComponentsHolder
	HeaderValidator           epochStart.HeaderValidator
	DataPool                  dataRetriever.PoolsHolder
	StorageService            dataRetriever.StorageService
	RequestHandler            process.RequestHandler
	ShardCoordinator          sharding.Coordinator
	Messenger                 p2p.Messenger
	ActiveAccountsDBs         map[state.AccountsDbIdentifier]state.AccountsAdapter
	ExistingResolvers         dataRetriever.ResolversContainer
	ExistingRequesters        dataRetriever.RequestersContainer
	ExportFolder              string
	ExportTriesStorageConfig  config.StorageConfig
	ExportStateStorageConfig  config.StorageConfig
	ExportStateKeysConfig     config.StorageConfig
	MaxTrieLevelInMemory      uint
	WhiteListHandler          process.WhiteListHandler
	WhiteListerVerifiedTxs    process.WhiteListHandler
	InterceptorsContainer     process.InterceptorsContainer
	NodesCoordinator          nodesCoordinator.NodesCoordinator
	HeaderSigVerifier         process.InterceptedHeaderSigVerifier
	HeaderIntegrityVerifier   process.HeaderIntegrityVerifier
	ValidityAttester          process.ValidityAttester
	InputAntifloodHandler     process.P2PAntifloodHandler
	OutputAntifloodHandler    process.P2PAntifloodHandler
	RoundHandler              process.RoundHandler
	PeersRatingHandler        dataRetriever.PeersRatingHandler
	InterceptorDebugConfig    config.InterceptorResolverDebugConfig
	MaxHardCapForMissingNodes int
	NumConcurrentTrieSyncers  int
	TrieSyncerVersion         int
	CheckNodesOnDisk          bool

	ShardCoordinatorFactory sharding.ShardCoordinatorFactory
}

type exportHandlerFactory struct {
	CoreComponents            process.CoreComponentsHolder
	CryptoComponents          process.CryptoComponentsHolder
	statusCoreComponents      process.StatusCoreComponentsHolder
	headerValidator           epochStart.HeaderValidator
	dataPool                  dataRetriever.PoolsHolder
	storageService            dataRetriever.StorageService
	requestHandler            process.RequestHandler
	shardCoordinator          sharding.Coordinator
	messenger                 p2p.Messenger
	activeAccountsDBs         map[state.AccountsDbIdentifier]state.AccountsAdapter
	exportFolder              string
	exportTriesStorageConfig  config.StorageConfig
	exportStateStorageConfig  config.StorageConfig
	exportStateKeysConfig     config.StorageConfig
	maxTrieLevelInMemory      uint
	whiteListHandler          process.WhiteListHandler
	whiteListerVerifiedTxs    process.WhiteListHandler
	interceptorsContainer     process.InterceptorsContainer
	existingResolvers         dataRetriever.ResolversContainer
	existingRequesters        dataRetriever.RequestersContainer
	epochStartTrigger         epochStart.TriggerHandler
	accounts                  state.AccountsAdapter
	nodesCoordinator          nodesCoordinator.NodesCoordinator
	headerSigVerifier         process.InterceptedHeaderSigVerifier
	headerIntegrityVerifier   process.HeaderIntegrityVerifier
	validityAttester          process.ValidityAttester
	resolverContainer         dataRetriever.ResolversContainer
	requestersContainer       dataRetriever.RequestersContainer
	inputAntifloodHandler     process.P2PAntifloodHandler
	outputAntifloodHandler    process.P2PAntifloodHandler
	roundHandler              process.RoundHandler
	peersRatingHandler        dataRetriever.PeersRatingHandler
	interceptorDebugConfig    config.InterceptorResolverDebugConfig
	maxHardCapForMissingNodes int
	numConcurrentTrieSyncers  int
	trieSyncerVersion         int
	checkNodesOnDisk          bool

	shardCoordinatorFactory sharding.ShardCoordinatorFactory
}

// NewExportHandlerFactory creates an exporter factory
func NewExportHandlerFactory(args ArgsExporter) (*exportHandlerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.CoreComponents) {
		return nil, update.ErrNilCoreComponents
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
	if check.IfNil(args.Messenger) {
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
	if check.IfNil(args.InterceptorsContainer) {
		return nil, update.ErrNilInterceptorsContainer
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
	if check.IfNil(args.InputAntifloodHandler) {
		return nil, update.ErrNilAntiFloodHandler
	}
	if check.IfNil(args.OutputAntifloodHandler) {
		return nil, update.ErrNilAntiFloodHandler
	}
	if check.IfNil(args.RoundHandler) {
		return nil, update.ErrNilRoundHandler
	}
	if check.IfNil(args.PeersRatingHandler) {
		return nil, update.ErrNilPeersRatingHandler
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
	if check.IfNil(args.ShardCoordinatorFactory) {
		return nil, errors.ErrNilShardCoordinatorFactory
	}

	e := &exportHandlerFactory{
		CoreComponents:            args.CoreComponents,
		CryptoComponents:          args.CryptoComponents,
		headerValidator:           args.HeaderValidator,
		dataPool:                  args.DataPool,
		storageService:            args.StorageService,
		requestHandler:            args.RequestHandler,
		shardCoordinator:          args.ShardCoordinator,
		messenger:                 args.Messenger,
		activeAccountsDBs:         args.ActiveAccountsDBs,
		exportFolder:              args.ExportFolder,
		exportTriesStorageConfig:  args.ExportTriesStorageConfig,
		exportStateStorageConfig:  args.ExportStateStorageConfig,
		exportStateKeysConfig:     args.ExportStateKeysConfig,
		interceptorsContainer:     args.InterceptorsContainer,
		whiteListHandler:          args.WhiteListHandler,
		whiteListerVerifiedTxs:    args.WhiteListerVerifiedTxs,
		existingResolvers:         args.ExistingResolvers,
		existingRequesters:        args.ExistingRequesters,
		accounts:                  args.ActiveAccountsDBs[state.UserAccountsState],
		nodesCoordinator:          args.NodesCoordinator,
		headerSigVerifier:         args.HeaderSigVerifier,
		headerIntegrityVerifier:   args.HeaderIntegrityVerifier,
		validityAttester:          args.ValidityAttester,
		inputAntifloodHandler:     args.InputAntifloodHandler,
		outputAntifloodHandler:    args.OutputAntifloodHandler,
		maxTrieLevelInMemory:      args.MaxTrieLevelInMemory,
		roundHandler:              args.RoundHandler,
		peersRatingHandler:        args.PeersRatingHandler,
		interceptorDebugConfig:    args.InterceptorDebugConfig,
		maxHardCapForMissingNodes: args.MaxHardCapForMissingNodes,
		numConcurrentTrieSyncers:  args.NumConcurrentTrieSyncers,
		trieSyncerVersion:         args.TrieSyncerVersion,
		checkNodesOnDisk:          args.CheckNodesOnDisk,
		statusCoreComponents:      args.StatusCoreComponents,
		shardCoordinatorFactory:   args.ShardCoordinatorFactory,
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
		Marshalizer:          e.CoreComponents.InternalMarshalizer(),
		Hasher:               e.CoreComponents.Hasher(),
		HeaderValidator:      e.headerValidator,
		Uint64Converter:      e.CoreComponents.Uint64ByteSliceConverter(),
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
		EnableEpochsHandler:  e.CoreComponents.EnableEpochsHandler(),
	}
	epochHandler, err := shardchain.NewEpochStartTrigger(&argsEpochTrigger)
	if err != nil {
		return nil, err
	}

	argsDataTrieFactory := ArgsNewDataTrieFactory{
		StorageConfig:        e.exportTriesStorageConfig,
		SyncFolder:           e.exportFolder,
		Marshalizer:          e.CoreComponents.InternalMarshalizer(),
		Hasher:               e.CoreComponents.Hasher(),
		ShardCoordinator:     e.shardCoordinator,
		MaxTrieLevelInMemory: e.maxTrieLevelInMemory,
		EnableEpochsHandler:  e.CoreComponents.EnableEpochsHandler(),
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
		Messenger:                  e.messenger,
		Marshalizer:                e.CoreComponents.InternalMarshalizer(),
		DataTrieContainer:          dataTries,
		ExistingResolvers:          e.existingResolvers,
		NumConcurrentResolvingJobs: 100,
		InputAntifloodHandler:      e.inputAntifloodHandler,
		OutputAntifloodHandler:     e.outputAntifloodHandler,
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
		ShardCoordinator:        e.shardCoordinator,
		Messenger:               e.messenger,
		Marshaller:              e.CoreComponents.InternalMarshalizer(),
		ExistingRequesters:      e.existingRequesters,
		OutputAntifloodHandler:  e.outputAntifloodHandler,
		PeersRatingHandler:      e.peersRatingHandler,
		ShardCoordinatorFactory: e.shardCoordinatorFactory,
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
		Hasher:                    e.CoreComponents.Hasher(),
		Marshalizer:               e.CoreComponents.InternalMarshalizer(),
		TrieStorageManager:        trieStorageManager,
		TimoutGettingTrieNode:     common.TimeoutGettingTrieNodesInHardfork,
		MaxTrieLevelInMemory:      e.maxTrieLevelInMemory,
		MaxHardCapForMissingNodes: e.maxHardCapForMissingNodes,
		NumConcurrentTrieSyncers:  e.numConcurrentTrieSyncers,
		TrieSyncerVersion:         e.trieSyncerVersion,
		CheckNodesOnDisk:          e.checkNodesOnDisk,
		AddressPubKeyConverter:    e.CoreComponents.AddressPubKeyConverter(),
		EnableEpochsHandler:       e.CoreComponents.EnableEpochsHandler(),
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
		Marshalizer:      e.CoreComponents.InternalMarshalizer(),
		Hasher:           e.CoreComponents.Hasher(),
		EpochHandler:     epochHandler,
		RequestHandler:   e.requestHandler,
		Uint64Converter:  e.CoreComponents.Uint64ByteSliceConverter(),
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
		Marshalizer:    e.CoreComponents.InternalMarshalizer(),
		RequestHandler: e.requestHandler,
	}
	epochStartMiniBlocksSyncer, err := sync.NewPendingMiniBlocksSyncer(argsMiniBlockSyncer)
	if err != nil {
		return nil, err
	}

	argsPendingTransactions := sync.ArgsNewTransactionsSyncer{
		DataPools:      e.dataPool,
		Storages:       e.storageService,
		Marshaller:     e.CoreComponents.InternalMarshalizer(),
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
		Marshalizer: e.CoreComponents.InternalMarshalizer(),
	}
	hs, err := storing.NewHardforkStorer(arg)
	if err != nil {
		return nil, fmt.Errorf("%w while creating hardfork storer", err)
	}

	argsExporter := genesis.ArgsNewStateExporter{
		ShardCoordinator:         e.shardCoordinator,
		StateSyncer:              stateSyncer,
		Marshalizer:              e.CoreComponents.InternalMarshalizer(),
		HardforkStorer:           hs,
		Hasher:                   e.CoreComponents.Hasher(),
		ExportFolder:             e.exportFolder,
		ValidatorPubKeyConverter: e.CoreComponents.ValidatorPubKeyConverter(),
		AddressPubKeyConverter:   e.CoreComponents.AddressPubKeyConverter(),
		GenesisNodesSetupHandler: e.CoreComponents.GenesisNodesSetup(),
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

	e.interceptorsContainer.Iterate(func(key string, interceptor process.Interceptor) bool {
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
		CoreComponents:          e.CoreComponents,
		CryptoComponents:        e.CryptoComponents,
		Accounts:                e.accounts,
		ShardCoordinator:        e.shardCoordinator,
		NodesCoordinator:        e.nodesCoordinator,
		Messenger:               e.messenger,
		Store:                   e.storageService,
		DataPool:                e.dataPool,
		MaxTxNonceDeltaAllowed:  math.MaxInt32,
		TxFeeHandler:            &disabled.FeeHandler{},
		BlockBlackList:          cache.NewTimeCache(time.Second),
		HeaderSigVerifier:       e.headerSigVerifier,
		HeaderIntegrityVerifier: e.headerIntegrityVerifier,
		SizeCheckDelta:          math.MaxUint32,
		ValidityAttester:        e.validityAttester,
		EpochStartTrigger:       e.epochStartTrigger,
		WhiteListHandler:        e.whiteListHandler,
		WhiteListerVerifiedTxs:  e.whiteListerVerifiedTxs,
		InterceptorsContainer:   e.interceptorsContainer,
		AntifloodHandler:        e.inputAntifloodHandler,
		ShardCoordinatorFactory: e.shardCoordinatorFactory,
	}
	fullSyncInterceptors, err := NewFullSyncInterceptorsContainerFactory(argsInterceptors)
	if err != nil {
		return err
	}

	interceptorsContainer, err := fullSyncInterceptors.Create()
	if err != nil {
		return err
	}

	e.interceptorsContainer = interceptorsContainer
	return nil
}

func createStorer(storageConfig config.StorageConfig, folder string) (storage.Storer, error) {
	dbConfig := storageFactory.GetDBFromConfig(storageConfig.DB)
	dbConfig.FilePath = path.Join(folder, storageConfig.DB.FilePath)
	accountsTrieStorage, err := storageunit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(storageConfig.Cache),
		dbConfig,
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
