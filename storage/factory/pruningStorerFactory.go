package factory

import (
	"fmt"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/clean"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.GetOrCreate("storage/factory")

const (
	minimumNumberOfActivePersisters = 1
	minimumNumberOfEpochsToKeep     = 2
)

// StorageServiceFactory handles the creation of storage services for both meta and shards
type StorageServiceFactory struct {
	generalConfig                 *config.Config
	prefsConfig                   *config.PreferencesConfig
	shardCoordinator              storage.ShardCoordinator
	pathManager                   storage.PathManagerHandler
	epochStartNotifier            storage.EpochStartNotifier
	oldDataCleanerProvider        clean.OldDataCleanerProvider
	createTrieEpochRootHashStorer bool
	currentEpoch                  uint32
}

// NewStorageServiceFactory will return a new instance of StorageServiceFactory
func NewStorageServiceFactory(
	config *config.Config,
	prefsConfig *config.PreferencesConfig,
	shardCoordinator storage.ShardCoordinator,
	pathManager storage.PathManagerHandler,
	epochStartNotifier storage.EpochStartNotifier,
	nodeTypeProvider NodeTypeProviderHandler,
	currentEpoch uint32,
	createTrieEpochRootHashStorer bool,
) (*StorageServiceFactory, error) {
	if config == nil {
		return nil, fmt.Errorf("%w for config.Config", storage.ErrNilConfig)
	}
	if prefsConfig == nil {
		return nil, fmt.Errorf("%w for config.PreferencesConfig", storage.ErrNilConfig)
	}
	if config.StoragePruning.NumActivePersisters < minimumNumberOfActivePersisters {
		return nil, storage.ErrInvalidNumberOfActivePersisters
	}
	if check.IfNil(shardCoordinator) {
		return nil, storage.ErrNilShardCoordinator
	}
	if check.IfNil(pathManager) {
		return nil, storage.ErrNilPathManager
	}
	if check.IfNil(epochStartNotifier) {
		return nil, storage.ErrNilEpochStartNotifier
	}

	oldDataCleanProvider, err := clean.NewOldDataCleanerProvider(
		nodeTypeProvider,
		config.StoragePruning,
	)
	if err != nil {
		return nil, err
	}
	if config.StoragePruning.NumEpochsToKeep < minimumNumberOfEpochsToKeep && oldDataCleanProvider.ShouldClean() {
		return nil, storage.ErrInvalidNumberOfEpochsToSave
	}

	return &StorageServiceFactory{
		generalConfig:                 config,
		prefsConfig:                   prefsConfig,
		shardCoordinator:              shardCoordinator,
		pathManager:                   pathManager,
		epochStartNotifier:            epochStartNotifier,
		currentEpoch:                  currentEpoch,
		createTrieEpochRootHashStorer: createTrieEpochRootHashStorer,
		oldDataCleanerProvider:        oldDataCleanProvider,
	}, nil
}

// CreateForShard will return the storage service which contains all storers needed for a shard
func (psf *StorageServiceFactory) CreateForShard() (dataRetriever.StorageService, error) {
	var headerUnit storage.Storer
	var peerBlockUnit storage.Storer
	var miniBlockUnit storage.Storer
	var txUnit storage.Storer
	var metachainHeaderUnit storage.Storer
	var unsignedTxUnit storage.Storer
	var rewardTxUnit storage.Storer
	var bootstrapUnit storage.Storer
	var receiptsUnit storage.Storer
	var err error

	successfullyCreatedStorers := make([]storage.Storer, 0)
	defer func() {
		// cleanup
		if err != nil {
			for _, storer := range successfullyCreatedStorers {
				_ = storer.DestroyUnit()
			}
		}
	}()

	txUnitStorerArgs := psf.createPruningStorerArgs(psf.generalConfig.TxStorage)
	txUnit, err = psf.createPruningPersister(txUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitStorerArgs := psf.createPruningStorerArgs(psf.generalConfig.UnsignedTransactionStorage)
	unsignedTxUnit, err = psf.createPruningPersister(unsignedTxUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	rewardTxUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.RewardTxStorage)
	rewardTxUnit, err = psf.createPruningPersister(rewardTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, rewardTxUnit)

	miniBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MiniBlocksStorage)
	miniBlockUnit, err = psf.createPruningPersister(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	peerBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.PeerBlockBodyStorage)
	peerBlockUnit, err = psf.createPruningPersister(peerBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, peerBlockUnit)

	headerUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BlockHeaderStorage)
	headerUnit, err = psf.createPruningPersister(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	metaChainHeaderUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MetaBlockStorage)
	metachainHeaderUnit, err = psf.createPruningPersister(metaChainHeaderUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metachainHeaderUnit)

	// metaHdrHashNonce is static
	metaHdrHashNonceUnitConfig := GetDBFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.DB)
	shardID := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.MetaHdrNonceHashStorage.DB.FilePath)
	metaHdrHashNonceUnitConfig.FilePath = dbPath
	metaHdrHashNonceUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Cache),
		metaHdrHashNonceUnitConfig,
		GetBloomFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	// shardHdrHashNonce storer is static
	shardHdrHashNonceConfig := GetDBFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.DB)
	shardID = core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardID, psf.generalConfig.ShardHdrNonceHashStorage.DB.FilePath) + shardID
	shardHdrHashNonceConfig.FilePath = dbPath
	shardHdrHashNonceUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Cache),
		shardHdrHashNonceConfig,
		GetBloomFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, shardHdrHashNonceUnit)

	heartbeatDbConfig := GetDBFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.DB)
	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.Heartbeat.HeartbeatStorage.DB.FilePath)
	heartbeatDbConfig.FilePath = dbPath
	heartbeatStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Cache),
		heartbeatDbConfig,
		GetBloomFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, heartbeatStorageUnit)

	statusMetricsDbConfig := GetDBFromConfig(psf.generalConfig.StatusMetricsStorage.DB)
	shardId = core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.StatusMetricsStorage.DB.FilePath)
	statusMetricsDbConfig.FilePath = dbPath
	statusMetricsStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.StatusMetricsStorage.Cache),
		statusMetricsDbConfig,
		GetBloomFromConfig(psf.generalConfig.StatusMetricsStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, statusMetricsStorageUnit)

	trieEpochRootHashStorageUnit, err := psf.createTrieEpochRootHashStorerIfNeeded()
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, trieEpochRootHashStorageUnit)

	bootstrapUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BootstrapStorage)
	bootstrapUnit, err = psf.createPruningPersister(bootstrapUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, bootstrapUnit)

	receiptsUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.ReceiptsStorage)
	receiptsUnit, err = psf.createPruningPersister(receiptsUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, receiptsUnit)

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.PeerChangesUnit, peerBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaBlockUnit, metachainHeaderUnit)
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, unsignedTxUnit)
	store.AddStorer(dataRetriever.RewardTransactionUnit, rewardTxUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(psf.shardCoordinator.SelfId())
	store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnit)
	store.AddStorer(dataRetriever.HeartbeatUnit, heartbeatStorageUnit)
	store.AddStorer(dataRetriever.BootstrapUnit, bootstrapUnit)
	store.AddStorer(dataRetriever.StatusMetricsUnit, statusMetricsStorageUnit)
	store.AddStorer(dataRetriever.ReceiptsUnit, receiptsUnit)
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit, trieEpochRootHashStorageUnit)

	createdStorers, err := psf.setupDbLookupExtensions(store)
	successfullyCreatedStorers = append(successfullyCreatedStorers, createdStorers...)
	if err != nil {
		return nil, err
	}

	createdStorers, err = psf.setupLogsAndEventsStorer(store)
	successfullyCreatedStorers = append(successfullyCreatedStorers, createdStorers...)
	if err != nil {
		return nil, err
	}

	err = psf.initOldDatabasesCleaningIfNeeded(store)
	if err != nil {
		return nil, err
	}

	return store, err
}

// CreateForMeta will return the storage service which contains all storers needed for metachain
func (psf *StorageServiceFactory) CreateForMeta() (dataRetriever.StorageService, error) {
	var metaBlockUnit storage.Storer
	var headerUnit storage.Storer
	var txUnit storage.Storer
	var miniBlockUnit storage.Storer
	var unsignedTxUnit storage.Storer
	var rewardTxUnit storage.Storer
	var bootstrapUnit storage.Storer
	var receiptsUnit storage.Storer
	var err error

	successfullyCreatedStorers := make([]storage.Storer, 0)

	defer func() {
		// cleanup
		if err != nil {
			log.Error("create meta store", "error", err.Error())
			for _, storer := range successfullyCreatedStorers {
				_ = storer.DestroyUnit()
			}
		}
	}()

	metaBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MetaBlockStorage)
	metaBlockUnit, err = psf.createPruningPersister(metaBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaBlockUnit)

	headerUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BlockHeaderStorage)
	headerUnit, err = psf.createPruningPersister(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	// metaHdrHashNonce is static
	metaHdrHashNonceUnitConfig := GetDBFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.DB)
	shardID := core.GetShardIDString(core.MetachainShardId)
	dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.MetaHdrNonceHashStorage.DB.FilePath)
	metaHdrHashNonceUnitConfig.FilePath = dbPath
	metaHdrHashNonceUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Cache),
		metaHdrHashNonceUnitConfig,
		GetBloomFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	shardHdrHashNonceUnits := make([]*storageUnit.Unit, psf.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceConfig := GetDBFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.DB)
		shardID = core.GetShardIDString(core.MetachainShardId)
		dbPath = psf.pathManager.PathForStatic(shardID, psf.generalConfig.ShardHdrNonceHashStorage.DB.FilePath) + fmt.Sprintf("%d", i)
		shardHdrHashNonceConfig.FilePath = dbPath
		shardHdrHashNonceUnits[i], err = storageUnit.NewStorageUnitFromConf(
			GetCacherFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Cache),
			shardHdrHashNonceConfig,
			GetBloomFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Bloom))
		if err != nil {
			return nil, err
		}

		successfullyCreatedStorers = append(successfullyCreatedStorers, shardHdrHashNonceUnits[i])
	}

	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	heartbeatDbConfig := GetDBFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.DB)
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.Heartbeat.HeartbeatStorage.DB.FilePath)
	heartbeatDbConfig.FilePath = dbPath
	heartbeatStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Cache),
		heartbeatDbConfig,
		GetBloomFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, heartbeatStorageUnit)

	statusMetricsDbConfig := GetDBFromConfig(psf.generalConfig.StatusMetricsStorage.DB)
	shardId = core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.StatusMetricsStorage.DB.FilePath)
	statusMetricsDbConfig.FilePath = dbPath
	statusMetricsStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.StatusMetricsStorage.Cache),
		statusMetricsDbConfig,
		GetBloomFromConfig(psf.generalConfig.StatusMetricsStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, statusMetricsStorageUnit)

	trieEpochRootHashStorageUnit, err := psf.createTrieEpochRootHashStorerIfNeeded()
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, trieEpochRootHashStorageUnit)

	txUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.TxStorage)
	txUnit, err = psf.createPruningPersister(txUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.UnsignedTransactionStorage)
	unsignedTxUnit, err = psf.createPruningPersister(unsignedTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	rewardTxUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.RewardTxStorage)
	rewardTxUnit, err = psf.createPruningPersister(rewardTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, rewardTxUnit)

	miniBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MiniBlocksStorage)
	miniBlockUnit, err = psf.createPruningPersister(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	bootstrapUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BootstrapStorage)
	bootstrapUnit, err = psf.createPruningPersister(bootstrapUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, bootstrapUnit)

	receiptsUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.ReceiptsStorage)
	receiptsUnit, err = psf.createPruningPersister(receiptsUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, receiptsUnit)

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, unsignedTxUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.RewardTransactionUnit, rewardTxUnit)
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnits[i])
	}
	store.AddStorer(dataRetriever.HeartbeatUnit, heartbeatStorageUnit)
	store.AddStorer(dataRetriever.BootstrapUnit, bootstrapUnit)
	store.AddStorer(dataRetriever.StatusMetricsUnit, statusMetricsStorageUnit)
	store.AddStorer(dataRetriever.ReceiptsUnit, receiptsUnit)
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit, trieEpochRootHashStorageUnit)

	createdStorers, err := psf.setupDbLookupExtensions(store)
	successfullyCreatedStorers = append(successfullyCreatedStorers, createdStorers...)
	if err != nil {
		return nil, err
	}

	createdStorers, err = psf.setupLogsAndEventsStorer(store)
	successfullyCreatedStorers = append(successfullyCreatedStorers, createdStorers...)
	if err != nil {
		return nil, err
	}

	err = psf.initOldDatabasesCleaningIfNeeded(store)
	if err != nil {
		return nil, err
	}

	return store, err
}

func (psf *StorageServiceFactory) setupLogsAndEventsStorer(chainStorer *dataRetriever.ChainStorer) ([]storage.Storer, error) {
	createdStorers := make([]storage.Storer, 0)

	// Should not create logs and events storer in the next case:
	// - LogsAndEvents.Enabled = false and DbLookupExtensions.Enabled = false
	// If we have DbLookupExtensions ACTIVE node by default should save logs no matter if is enabled or not
	shouldCreateStorer := psf.generalConfig.LogsAndEvents.SaveInStorageEnabled || psf.generalConfig.DbLookupExtensions.Enabled
	if !shouldCreateStorer {
		return createdStorers, nil
	}

	txLogsUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.LogsAndEvents.TxLogsStorage)
	txLogsUnit, err := psf.createPruningPersister(txLogsUnitArgs)
	if err != nil {
		return createdStorers, err
	}

	createdStorers = append(createdStorers, txLogsUnit)
	chainStorer.AddStorer(dataRetriever.TxLogsUnit, txLogsUnit)

	return createdStorers, nil
}

func (psf *StorageServiceFactory) setupDbLookupExtensions(chainStorer *dataRetriever.ChainStorer) ([]storage.Storer, error) {
	createdStorers := make([]storage.Storer, 0)

	if !psf.generalConfig.DbLookupExtensions.Enabled {
		return createdStorers, nil
	}

	shardID := core.GetShardIDString(psf.shardCoordinator.SelfId())

	// Create the eventsHashesByTxHash (PRUNING) storer
	eventsHashesByTxHashConfig := psf.generalConfig.DbLookupExtensions.ResultsHashesByTxHashStorageConfig
	eventsHashesByTxHashStorerArgs := psf.createPruningStorerArgs(eventsHashesByTxHashConfig)
	eventsHashesByTxHashPruningStorer, err := psf.createPruningPersister(eventsHashesByTxHashStorerArgs)
	if err != nil {
		return createdStorers, err
	}

	createdStorers = append(createdStorers, eventsHashesByTxHashPruningStorer)
	chainStorer.AddStorer(dataRetriever.ResultsHashesByTxHashUnit, eventsHashesByTxHashPruningStorer)

	// Create the miniblocksMetadata (PRUNING) storer
	miniblocksMetadataConfig := psf.generalConfig.DbLookupExtensions.MiniblocksMetadataStorageConfig
	miniblocksMetadataPruningStorerArgs := psf.createPruningStorerArgs(miniblocksMetadataConfig)
	miniblocksMetadataPruningStorer, err := psf.createPruningPersister(miniblocksMetadataPruningStorerArgs)
	if err != nil {
		return createdStorers, err
	}

	createdStorers = append(createdStorers, miniblocksMetadataPruningStorer)
	chainStorer.AddStorer(dataRetriever.MiniblocksMetadataUnit, miniblocksMetadataPruningStorer)

	// Create the miniblocksHashByTxHash (STATIC) storer
	miniblockHashByTxHashConfig := psf.generalConfig.DbLookupExtensions.MiniblockHashByTxHashStorageConfig
	miniblockHashByTxHashDbConfig := GetDBFromConfig(miniblockHashByTxHashConfig.DB)
	miniblockHashByTxHashDbConfig.FilePath = psf.pathManager.PathForStatic(shardID, miniblockHashByTxHashConfig.DB.FilePath)
	miniblockHashByTxHashCacherConfig := GetCacherFromConfig(miniblockHashByTxHashConfig.Cache)
	miniblockHashByTxHashBloomFilter := GetBloomFromConfig(miniblockHashByTxHashConfig.Bloom)
	miniblockHashByTxHashUnit, err := storageUnit.NewStorageUnitFromConf(miniblockHashByTxHashCacherConfig, miniblockHashByTxHashDbConfig, miniblockHashByTxHashBloomFilter)
	if err != nil {
		return createdStorers, err
	}

	createdStorers = append(createdStorers, miniblockHashByTxHashUnit)
	chainStorer.AddStorer(dataRetriever.MiniblockHashByTxHashUnit, miniblockHashByTxHashUnit)

	// Create the blockHashByRound (STATIC) storer
	blockHashByRoundConfig := psf.generalConfig.DbLookupExtensions.RoundHashStorageConfig
	blockHashByRoundDBConfig := GetDBFromConfig(blockHashByRoundConfig.DB)
	blockHashByRoundDBConfig.FilePath = psf.pathManager.PathForStatic(shardID, blockHashByRoundConfig.DB.FilePath)
	blockHashByRoundCacherConfig := GetCacherFromConfig(blockHashByRoundConfig.Cache)
	blockHashByRoundBloomFilter := GetBloomFromConfig(blockHashByRoundConfig.Bloom)
	blockHashByRoundUnit, err := storageUnit.NewStorageUnitFromConf(blockHashByRoundCacherConfig, blockHashByRoundDBConfig, blockHashByRoundBloomFilter)
	if err != nil {
		return createdStorers, err
	}

	createdStorers = append(createdStorers, blockHashByRoundUnit)
	chainStorer.AddStorer(dataRetriever.RoundHdrHashDataUnit, blockHashByRoundUnit)

	// Create the epochByHash (STATIC) storer
	epochByHashConfig := psf.generalConfig.DbLookupExtensions.EpochByHashStorageConfig
	epochByHashDbConfig := GetDBFromConfig(epochByHashConfig.DB)
	epochByHashDbConfig.FilePath = psf.pathManager.PathForStatic(shardID, epochByHashConfig.DB.FilePath)
	epochByHashCacherConfig := GetCacherFromConfig(epochByHashConfig.Cache)
	epochByHashBloomFilter := GetBloomFromConfig(epochByHashConfig.Bloom)
	epochByHashUnit, err := storageUnit.NewStorageUnitFromConf(epochByHashCacherConfig, epochByHashDbConfig, epochByHashBloomFilter)
	if err != nil {
		return createdStorers, err
	}

	createdStorers = append(createdStorers, epochByHashUnit)
	chainStorer.AddStorer(dataRetriever.EpochByHashUnit, epochByHashUnit)

	esdtSuppliesConfig := psf.generalConfig.DbLookupExtensions.ESDTSuppliesStorageConfig
	esdtSuppliesDbConfig := GetDBFromConfig(esdtSuppliesConfig.DB)
	esdtSuppliesDbConfig.FilePath = psf.pathManager.PathForStatic(shardID, esdtSuppliesConfig.DB.FilePath)
	esdtSuppliesCacherConfig := GetCacherFromConfig(esdtSuppliesConfig.Cache)
	esdtSuppliesBloomFilter := GetBloomFromConfig(esdtSuppliesConfig.Bloom)
	esdtSuppliesUnit, err := storageUnit.NewStorageUnitFromConf(esdtSuppliesCacherConfig, esdtSuppliesDbConfig, esdtSuppliesBloomFilter)
	if err != nil {
		return createdStorers, err
	}

	createdStorers = append(createdStorers, esdtSuppliesUnit)
	chainStorer.AddStorer(dataRetriever.ESDTSuppliesUnit, esdtSuppliesUnit)

	return createdStorers, nil
}

func (psf *StorageServiceFactory) createPruningStorerArgs(storageConfig config.StorageConfig) *pruning.StorerArgs {
	numOfEpochsToKeep := uint32(psf.generalConfig.StoragePruning.NumEpochsToKeep)
	numOfActivePersisters := uint32(psf.generalConfig.StoragePruning.NumActivePersisters)
	pruningEnabled := psf.generalConfig.StoragePruning.Enabled
	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := filepath.Join(psf.pathManager.PathForEpoch(shardId, psf.currentEpoch, storageConfig.DB.FilePath))
	args := &pruning.StorerArgs{
		Identifier:                storageConfig.DB.FilePath,
		PruningEnabled:            pruningEnabled,
		StartingEpoch:             psf.currentEpoch,
		OldDataCleanerProvider:    psf.oldDataCleanerProvider,
		ShardCoordinator:          psf.shardCoordinator,
		CacheConf:                 GetCacherFromConfig(storageConfig.Cache),
		PathManager:               psf.pathManager,
		DbPath:                    dbPath,
		PersisterFactory:          NewPersisterFactory(storageConfig.DB),
		BloomFilterConf:           GetBloomFromConfig(storageConfig.Bloom),
		NumOfEpochsToKeep:         numOfEpochsToKeep,
		NumOfActivePersisters:     numOfActivePersisters,
		Notifier:                  psf.epochStartNotifier,
		MaxBatchSize:              storageConfig.DB.MaxBatchSize,
		EnabledDbLookupExtensions: psf.generalConfig.DbLookupExtensions.Enabled,
	}

	return args
}

func (psf *StorageServiceFactory) createTrieEpochRootHashStorerIfNeeded() (storage.Storer, error) {
	if !psf.createTrieEpochRootHashStorer {
		return storageUnit.NewNilStorer(), nil
	}

	trieEpochRootHashDbConfig := GetDBFromConfig(psf.generalConfig.TrieEpochRootHashStorage.DB)
	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := psf.pathManager.PathForStatic(shardId, psf.generalConfig.TrieEpochRootHashStorage.DB.FilePath)
	trieEpochRootHashDbConfig.FilePath = dbPath
	trieEpochRootHashStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.TrieEpochRootHashStorage.Cache),
		trieEpochRootHashDbConfig,
		GetBloomFromConfig(psf.generalConfig.TrieEpochRootHashStorage.Bloom))
	if err != nil {
		return nil, err
	}

	return trieEpochRootHashStorageUnit, nil
}

func (psf *StorageServiceFactory) createPruningPersister(arg *pruning.StorerArgs) (storage.Storer, error) {
	isFullArchive := psf.prefsConfig.FullArchive
	isDBLookupExtenstion := psf.generalConfig.DbLookupExtensions.Enabled
	if !isFullArchive && !isDBLookupExtenstion {
		return pruning.NewPruningStorer(arg)
	}

	numOldActivePersisters := psf.getNumActivePersistersForFullHistoryStorer(isFullArchive, isDBLookupExtenstion)
	historyArgs := &pruning.FullHistoryStorerArgs{
		StorerArgs:               arg,
		NumOfOldActivePersisters: numOldActivePersisters,
	}

	return pruning.NewFullHistoryPruningStorer(historyArgs)
}

func (psf *StorageServiceFactory) getNumActivePersistersForFullHistoryStorer(isFullArchive bool, isDBLookupExtension bool) uint32 {
	if isFullArchive && !isDBLookupExtension {
		return psf.generalConfig.StoragePruning.FullArchiveNumActivePersisters
	}

	if !isFullArchive && isDBLookupExtension {
		return psf.generalConfig.DbLookupExtensions.DbLookupMaxActivePersisters
	}

	if psf.generalConfig.DbLookupExtensions.DbLookupMaxActivePersisters != psf.generalConfig.StoragePruning.FullArchiveNumActivePersisters {
		log.Warn("node is started with both Full Archive and DB Lookup Extension modes and have different values " +
			"for the number of active persisters. It will use NumOfOldActivePersisters from full archive's settings")
	}

	return psf.generalConfig.StoragePruning.FullArchiveNumActivePersisters
}

func (psf *StorageServiceFactory) initOldDatabasesCleaningIfNeeded(store dataRetriever.StorageService) error {
	isFullArchive := psf.prefsConfig.FullArchive
	if isFullArchive {
		return nil
	}
	_, err := clean.NewOldDatabaseCleaner(clean.ArgsOldDatabaseCleaner{
		DatabasePath:           psf.pathManager.DatabasePath(),
		StorageListProvider:    store,
		EpochStartNotifier:     psf.epochStartNotifier,
		OldDataCleanerProvider: psf.oldDataCleanerProvider,
	})

	return err
}
