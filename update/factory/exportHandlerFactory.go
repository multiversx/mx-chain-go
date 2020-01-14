package factory

import (
	"path"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
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
)

type ArgsExporter struct {
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	HeaderValidator          epochStart.HeaderValidator
	Uint64Converter          typeConverters.Uint64ByteSliceConverter
	DataPool                 dataRetriever.PoolsHolder
	StorageService           dataRetriever.StorageService
	RequestHandler           update.RequestHandler
	ShardCoordinator         sharding.Coordinator
	Messenger                p2p.Messenger
	ActiveTries              update.DataTriesContainer
	ExportFolder             string
	ExportTriesStorageConfig config.StorageConfig
	ExportTriesCacheConfig   config.CacheConfig
	ExportStateStorageConfig config.StorageConfig
}

type exportHandlerFactory struct {
	marshalizer              marshal.Marshalizer
	hasher                   hashing.Hasher
	headerValidator          epochStart.HeaderValidator
	uint64Converter          typeConverters.Uint64ByteSliceConverter
	dataPool                 dataRetriever.PoolsHolder
	storageService           dataRetriever.StorageService
	requestHandler           update.RequestHandler
	shardCoordinator         sharding.Coordinator
	messenger                p2p.Messenger
	activeTries              update.DataTriesContainer
	exportFolder             string
	exportTriesStorageConfig config.StorageConfig
	exportTriesCacheConfig   config.CacheConfig
	exportStateStorageConfig config.StorageConfig
}

func NewExportHandlerFactory(args ArgsExporter) (*exportHandlerFactory, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, sharding.ErrNilShardCoordinator
	}
	if check.IfNil(args.Hasher) {
		return nil, sharding.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, data.ErrNilMarshalizer
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
	}

	return e, nil
}

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
		Finality:           process.MetaBlockFinality,
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

	argsSyncState := genesis.ArgsNewSyncState{
		Hasher:           e.hasher,
		Marshalizer:      e.marshalizer,
		ShardCoordinator: e.shardCoordinator,
		TrieSyncers:      trieSyncers,
		EpochHandler:     epochHandler,
		Storages:         e.storageService,
		DataPools:        e.dataPool,
		RequestHandler:   e.requestHandler,
		HeaderValidator:  e.headerValidator,
		ActiveDataTries:  e.activeTries,
	}
	stateSyncer, err := genesis.NewSyncState(argsSyncState)
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

	return exportHandler, nil
}

func createFinalExportStorage(storageConfig config.StorageConfig, folder string) (storage.Storer, error) {
	dbConfig := storageFactory.GetDBFromConfig(storageConfig.DB)
	dbConfig.FilePath = path.Join(folder, "syncTries")
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

func (e *exportHandlerFactory) IsInterfaceNil() bool {
	return e == nil
}
