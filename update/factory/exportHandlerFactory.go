package factory

import (
	"path"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/files"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
	"github.com/ElrondNetwork/elrond-go/update/sync"
)

// ArgsExporter is the argument structure to create a new exporter
type ArgsExporter struct {
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	HeaderValidator          epochStart.HeaderValidator
	Uint64Converter          typeConverters.Uint64ByteSliceConverter
	DataPool                 dataRetriever.PoolsHolder
	StorageService           dataRetriever.StorageService
	RequestHandler           process.RequestHandler
	ShardCoordinator         sharding.Coordinator
	Messenger                p2p.Messenger
	ActiveTries              state.TriesHolder
	ExistingResolvers        dataRetriever.ResolversContainer
	ExportFolder             string
	ExportTriesStorageConfig config.StorageConfig
	ExportTriesCacheConfig   config.CacheConfig
	ExportStateStorageConfig config.StorageConfig
	WhiteListHandler         process.InterceptedDataWhiteList
	InterceptorsContainer    process.InterceptorsContainer
}

type exportHandlerFactory struct {
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	headerValidator          epochStart.HeaderValidator
	uint64Converter          typeConverters.Uint64ByteSliceConverter
	dataPool                 dataRetriever.PoolsHolder
	storageService           dataRetriever.StorageService
	requestHandler           process.RequestHandler
	shardCoordinator         sharding.Coordinator
	messenger                p2p.Messenger
	activeTries              state.TriesHolder
	exportFolder             string
	exportTriesStorageConfig config.StorageConfig
	exportTriesCacheConfig   config.CacheConfig
	exportStateStorageConfig config.StorageConfig
	whiteListHandler         process.InterceptedDataWhiteList
	interceptorsContainer    process.InterceptorsContainer
	existingResolvers        dataRetriever.ResolversContainer
}

// NewExportHandlerFactory creates an exporter factory
func NewExportHandlerFactory(args ArgsExporter) (*exportHandlerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.HeaderValidator) {
		return nil, update.ErrNilHeaderValidator
	}
	if check.IfNil(args.Uint64Converter) {
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
	if check.IfNil(args.ActiveTries) {
		return nil, update.ErrNilActiveTries
	}
	if check.IfNil(args.WhiteListHandler) {
		return nil, update.ErrNilWhiteListHandler
	}
	if check.IfNil(args.InterceptorsContainer) {
		return nil, update.ErrNilInterceptorsContainer
	}
	if check.IfNil(args.ExistingResolvers) {
		return nil, update.ErrNilResolverContainer
	}

	e := &exportHandlerFactory{
		marshalizer:              args.Marshalizer,
		hasher:                   args.Hasher,
		headerValidator:          args.HeaderValidator,
		uint64Converter:          args.Uint64Converter,
		dataPool:                 args.DataPool,
		storageService:           args.StorageService,
		requestHandler:           args.RequestHandler,
		shardCoordinator:         args.ShardCoordinator,
		messenger:                args.Messenger,
		activeTries:              args.ActiveTries,
		exportFolder:             args.ExportFolder,
		exportTriesStorageConfig: args.ExportTriesStorageConfig,
		exportTriesCacheConfig:   args.ExportTriesCacheConfig,
		exportStateStorageConfig: args.ExportStateStorageConfig,
		interceptorsContainer:    args.InterceptorsContainer,
		whiteListHandler:         args.WhiteListHandler,
		existingResolvers:        args.ExistingResolvers,
	}

	return e, nil
}

// Create makes a new export handler
func (e *exportHandlerFactory) Create() (update.ExportHandler, error) {
	argsEpochTrigger := shardchain.ArgsShardEpochStartTrigger{
		Marshalizer:        e.marshalizer,
		Hasher:             e.hasher,
		HeaderValidator:    e.headerValidator,
		Uint64Converter:    e.uint64Converter,
		DataPool:           e.dataPool,
		Storage:            e.storageService,
		RequestHandler:     e.requestHandler,
		EpochStartNotifier: notifier.NewEpochStartSubscriptionHandler(),
		Epoch:              0,
		Validity:           process.MetaBlockValidity,
		Finality:           process.BlockFinality,
	}
	epochHandler, err := shardchain.NewEpochStartTrigger(&argsEpochTrigger)
	if err != nil {
		return nil, err
	}

	argsDataTrieFactory := ArgsNewDataTrieFactory{
		StorageConfig:    e.exportTriesStorageConfig,
		SyncFolder:       e.exportFolder,
		Marshalizer:      e.marshalizer,
		Hasher:           e.hasher,
		ShardCoordinator: e.shardCoordinator,
	}
	dataTriesContainerFactory, err := NewDataTrieFactory(argsDataTrieFactory)
	if err != nil {
		return nil, err
	}
	dataTries, err := dataTriesContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	argsResolvers := ArgsNewResolversContainerFactory{
		ShardCoordinator:  e.shardCoordinator,
		Messenger:         e.messenger,
		Marshalizer:       e.marshalizer,
		DataTrieContainer: dataTries,
		ExistingResolvers: e.existingResolvers,
	}
	resolversContainerFactory, err := NewResolversContainerFactory(argsResolvers)
	if err != nil {
		return nil, err
	}
	resolversContainer, err := resolversContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	argsTrieSyncers := ArgsNewTrieSyncersContainerFactory{
		CacheConfig:        e.exportTriesCacheConfig,
		SyncFolder:         e.exportFolder,
		ResolversContainer: resolversContainer,
		DataTrieContainer:  dataTries,
		ShardCoordinator:   e.shardCoordinator,
	}
	trieSyncContainersFactory, err := NewTrieSyncersContainerFactory(argsTrieSyncers)
	if err != nil {
		return nil, err
	}
	trieSyncers, err := trieSyncContainersFactory.Create()
	if err != nil {
		return nil, err
	}

	argsNewHeadersSync := sync.ArgsNewHeadersSyncHandler{
		StorageService:  e.storageService,
		Cache:           e.dataPool.Headers(),
		Marshalizer:     e.marshalizer,
		EpochHandler:    epochHandler,
		RequestHandler:  e.requestHandler,
		Uint64Converter: e.uint64Converter,
	}
	epochStartHeadersSyncer, err := sync.NewHeadersSyncHandler(argsNewHeadersSync)
	if err != nil {
		return nil, err
	}

	argsNewEpochStartTrieSyncer := sync.ArgsNewSyncTriesHandler{
		TrieSyncers: trieSyncers,
		ActiveTries: e.activeTries,
	}
	_, err = sync.NewSyncTriesHandler(argsNewEpochStartTrieSyncer)
	if err != nil {
		return nil, err
	}

	argsMiniBlockSyncer := sync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        e.storageService.GetStorer(dataRetriever.MiniBlockUnit),
		Cache:          e.dataPool.MiniBlocks(),
		Marshalizer:    e.marshalizer,
		RequestHandler: e.requestHandler,
	}
	epochStartMiniBlocksSyncer, err := sync.NewPendingMiniBlocksSyncer(argsMiniBlockSyncer)
	if err != nil {
		return nil, err
	}

	argsPendingTransactions := sync.ArgsNewPendingTransactionsSyncer{
		DataPools:      e.dataPool,
		Storages:       e.storageService,
		Marshalizer:    e.marshalizer,
		RequestHandler: e.requestHandler,
	}
	epochStartTransactionsSyncer, err := sync.NewPendingTransactionsSyncer(argsPendingTransactions)
	if err != nil {
		return nil, err
	}

	argsSyncState := sync.ArgsNewSyncState{
		Headers:      epochStartHeadersSyncer,
		Tries:        sync.NewNilSyncTries(), // TODO change to the actual trie syncer
		MiniBlocks:   epochStartMiniBlocksSyncer,
		Transactions: epochStartTransactionsSyncer,
	}
	stateSyncer, err := sync.NewSyncState(argsSyncState)
	if err != nil {
		return nil, err
	}

	exportStore, err := createFinalExportStorage(e.exportStateStorageConfig, e.exportFolder)
	if err != nil {
		return nil, err
	}

	argsWriter := files.ArgsNewMultiFileWriter{
		ExportFolder: e.exportFolder,
		ExportStore:  exportStore,
	}
	writer, err := files.NewMultiFileWriter(argsWriter)
	if err != nil {
		return nil, err
	}

	argsExporter := genesis.ArgsNewStateExporter{
		ShardCoordinator: e.shardCoordinator,
		StateSyncer:      stateSyncer,
		Marshalizer:      e.marshalizer,
		Writer:           writer,
	}
	exportHandler, err := genesis.NewStateExporter(argsExporter)
	if err != nil {
		return nil, err
	}

	e.setWhiteListHandlerToInterceptors()

	return exportHandler, nil
}

func createFinalExportStorage(storageConfig config.StorageConfig, folder string) (storage.Storer, error) {
	dbConfig := storageFactory.GetDBFromConfig(storageConfig.DB)
	dbConfig.FilePath = path.Join(folder, storageConfig.DB.FilePath)
	accountsTrieStorage, err := storageUnit.NewStorageUnitFromConf(
		storageFactory.GetCacherFromConfig(storageConfig.Cache),
		dbConfig,
		storageFactory.GetBloomFromConfig(storageConfig.Bloom),
	)
	if err != nil {
		return nil, err
	}

	return accountsTrieStorage, nil
}

func (e *exportHandlerFactory) setWhiteListHandlerToInterceptors() {
	e.interceptorsContainer.Iterate(func(key string, interceptor process.Interceptor) bool {
		interceptor.SetIsDataForCurrentShardVerifier(e.whiteListHandler)
		return true
	})
}

// IsInterfaceNil returns true if underlying object is nil
func (e *exportHandlerFactory) IsInterfaceNil() bool {
	return e == nil
}
