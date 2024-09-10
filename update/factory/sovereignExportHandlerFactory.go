package factory

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/debug/factory"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/epochStart/shardchain"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/genesis"
	"github.com/multiversx/mx-chain-go/update/storing"
	"github.com/multiversx/mx-chain-go/update/sync"
)

type sovereignExportHandlerFactory struct {
	*exportHandlerFactory
}

func NewSovereignExportHandlerFactory(exportHandlerFactory *exportHandlerFactory) (*sovereignExportHandlerFactory, error) {
	return &sovereignExportHandlerFactory{
		exportHandlerFactory,
	}, nil
}

// Create makes a new export handler
func (e *sovereignExportHandlerFactory) Create() (update.ExportHandler, error) {
	err := e.prepareFolders(e.exportFolder)
	if err != nil {
		return nil, err
	}

	// TODO reuse the debugger when the one used for regular resolvers & interceptors will be moved inside the status components
	debugger, errNotCritical := factory.NewInterceptorDebuggerFactory(e.interceptorDebugConfig)
	if errNotCritical != nil {
		log.Warn("error creating hardfork debugger", "error", errNotCritical)
	}

	// TODO: SOVEREIGN CHECK IF WE EVEN NEED THIS
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
	// TODO: CHECK IF WE NEED SOVEREIGN COMPONENT
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
	// TODO: SOVEREIGN COMPONENT
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
	// TODO: SOVEREIGN COMPONENT
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
		MainMessenger:           e.networkComponents.NetworkMessenger(),
		FullArchiveMessenger:    e.networkComponents.FullArchiveNetworkMessenger(),
		Marshaller:              e.coreComponents.InternalMarshalizer(),
		ExistingRequesters:      e.existingRequesters,
		OutputAntifloodHandler:  e.networkComponents.OutputAntiFloodHandler(),
		PeersRatingHandler:      e.networkComponents.PeersRatingHandler(),
		ShardCoordinatorFactory: e.shardCoordinatorFactory,
	}
	// TODO: SOVEREIGN COMPONENT
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
	// TODO: SOVEREIGN COMPONENT
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

	// TODO: SOVEREIGN COMPONENT
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

func (e *sovereignExportHandlerFactory) createInterceptors() error {
	fullSyncInterceptors, err := e.createFullSyncInterceptors()
	if err != nil {
		return err
	}

	sovFullySyncInterceptors, err := NewSovereignFullSyncInterceptorsContainerFactory(fullSyncInterceptors)
	if err != nil {
		return err
	}

	mainInterceptorsContainer, fullArchiveInterceptorsContainer, err := sovFullySyncInterceptors.Create()
	if err != nil {
		return err
	}

	e.mainInterceptorsContainer = mainInterceptorsContainer
	e.fullArchiveInterceptorsContainer = fullArchiveInterceptorsContainer
	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *sovereignExportHandlerFactory) IsInterfaceNil() bool {
	return e == nil
}
