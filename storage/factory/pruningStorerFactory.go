package factory

import (
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
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
	generalConfig      *config.Config
	shardCoordinator   sharding.Coordinator
	pathManager        storage.PathManagerHandler
	epochStartNotifier storage.EpochStartNotifier
	currentEpoch       uint32
}

// NewStorageServiceFactory will return a new instance of StorageServiceFactory
func NewStorageServiceFactory(
	config *config.Config,
	shardCoordinator sharding.Coordinator,
	pathManager storage.PathManagerHandler,
	epochStartNotifier storage.EpochStartNotifier,
	currentEpoch uint32,
) (*StorageServiceFactory, error) {
	if config == nil {
		return nil, storage.ErrNilConfig
	}
	if config.StoragePruning.NumEpochsToKeep < minimumNumberOfEpochsToKeep && !config.StoragePruning.FullArchive {
		return nil, storage.ErrInvalidNumberOfEpochsToSave
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

	return &StorageServiceFactory{
		generalConfig:      config,
		shardCoordinator:   shardCoordinator,
		pathManager:        pathManager,
		epochStartNotifier: epochStartNotifier,
		currentEpoch:       currentEpoch,
	}, nil
}

// CreateForShard will return the storage service which contains all storers needed for a shard
func (psf *StorageServiceFactory) CreateForShard() (dataRetriever.StorageService, error) {
	var headerUnit *pruning.PruningStorer
	var peerBlockUnit *pruning.PruningStorer
	var miniBlockUnit *pruning.PruningStorer
	var txUnit *pruning.PruningStorer
	var metachainHeaderUnit *pruning.PruningStorer
	var unsignedTxUnit *pruning.PruningStorer
	var rewardTxUnit *pruning.PruningStorer
	var metaHdrHashNonceUnit *pruning.PruningStorer
	var shardHdrHashNonceUnit *pruning.PruningStorer
	var bootstrapUnit *pruning.PruningStorer
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
	txUnit, err = pruning.NewPruningStorer(txUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitStorerArgs := psf.createPruningStorerArgs(psf.generalConfig.UnsignedTransactionStorage)
	unsignedTxUnit, err = pruning.NewPruningStorer(unsignedTxUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	rewardTxUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.RewardTxStorage)
	rewardTxUnit, err = pruning.NewPruningStorer(rewardTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, rewardTxUnit)

	miniBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MiniBlocksStorage)
	miniBlockUnit, err = pruning.NewPruningStorer(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	peerBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.PeerBlockBodyStorage)
	peerBlockUnit, err = pruning.NewPruningStorer(peerBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, peerBlockUnit)

	headerUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BlockHeaderStorage)
	headerUnit, err = pruning.NewPruningStorer(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	metaChainHeaderUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MetaBlockStorage)
	metachainHeaderUnit, err = pruning.NewPruningStorer(metaChainHeaderUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metachainHeaderUnit)

	metaHdrHashNonceUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MetaHdrNonceHashStorage)
	metaHdrHashNonceUnit, err = pruning.NewPruningStorer(metaHdrHashNonceUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	shardHdrHashNonceUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.ShardHdrNonceHashStorage)
	shardHdrHashNonceUnit, err = pruning.NewShardedPruningStorer(shardHdrHashNonceUnitArgs, psf.shardCoordinator.SelfId())
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, shardHdrHashNonceUnit)

	heartbeatDbConfig := GetDBFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.DB)
	shardId := core.GetShardIdString(psf.shardCoordinator.SelfId())
	dbPath := psf.pathManager.PathForStatic(shardId, psf.generalConfig.Heartbeat.HeartbeatStorage.DB.FilePath)
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
	shardId = core.GetShardIdString(psf.shardCoordinator.SelfId())
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

	bootstrapUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BootstrapStorage)
	bootstrapUnit, err = pruning.NewPruningStorer(bootstrapUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, bootstrapUnit)

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

	return store, err
}

// CreateForMeta will return the storage service which contains all storers needed for metachain
func (psf *StorageServiceFactory) CreateForMeta() (dataRetriever.StorageService, error) {
	var metaBlockUnit *pruning.PruningStorer
	var headerUnit *pruning.PruningStorer
	var txUnit *pruning.PruningStorer
	var metaHdrHashNonceUnit *pruning.PruningStorer
	var miniBlockUnit *pruning.PruningStorer
	var unsignedTxUnit *pruning.PruningStorer
	var miniBlockHeadersUnit *pruning.PruningStorer
	var shardHdrHashNonceUnits []*pruning.PruningStorer
	var bootstrapUnit *pruning.PruningStorer
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
	metaBlockUnit, err = pruning.NewPruningStorer(metaBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaBlockUnit)

	headerUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BlockHeaderStorage)
	headerUnit, err = pruning.NewPruningStorer(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	metaHdrHashNonceUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MetaHdrNonceHashStorage)
	metaHdrHashNonceUnit, err = pruning.NewPruningStorer(metaHdrHashNonceUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	shardHdrHashNonceUnits = make([]*pruning.PruningStorer, psf.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.ShardHdrNonceHashStorage)
		shardHdrHashNonceUnits[i], err = pruning.NewShardedPruningStorer(shardHdrHashNonceUnitArgs, i)
		if err != nil {
			return nil, err
		}
		successfullyCreatedStorers = append(successfullyCreatedStorers, shardHdrHashNonceUnits[i])
	}

	shardId := core.GetShardIdString(psf.shardCoordinator.SelfId())
	heartbeatDbConfig := GetDBFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.DB)
	dbPath := psf.pathManager.PathForStatic(shardId, psf.generalConfig.Heartbeat.HeartbeatStorage.DB.FilePath)
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
	shardId = core.GetShardIdString(psf.shardCoordinator.SelfId())
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

	txUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.TxStorage)
	txUnit, err = pruning.NewPruningStorer(txUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.UnsignedTransactionStorage)
	unsignedTxUnit, err = pruning.NewPruningStorer(unsignedTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	miniBlockUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MiniBlocksStorage)
	miniBlockUnit, err = pruning.NewPruningStorer(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	miniBlockHeadersUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.MiniBlockHeadersStorage)
	miniBlockHeadersUnit, err = pruning.NewPruningStorer(miniBlockHeadersUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockHeadersUnit)

	bootstrapUnitArgs := psf.createPruningStorerArgs(psf.generalConfig.BootstrapStorage)
	bootstrapUnit, err = pruning.NewPruningStorer(bootstrapUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, bootstrapUnit)

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, unsignedTxUnit)
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)
	store.AddStorer(dataRetriever.MiniBlockHeaderUnit, miniBlockHeadersUnit)
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnits[i])
	}
	store.AddStorer(dataRetriever.HeartbeatUnit, heartbeatStorageUnit)
	store.AddStorer(dataRetriever.BootstrapUnit, bootstrapUnit)
	store.AddStorer(dataRetriever.StatusMetricsUnit, statusMetricsStorageUnit)

	return store, err
}

func (psf *StorageServiceFactory) createPruningStorerArgs(storageConfig config.StorageConfig) *pruning.StorerArgs {
	fullArchiveMode := psf.generalConfig.StoragePruning.FullArchive
	numOfEpochsToKeep := uint32(psf.generalConfig.StoragePruning.NumEpochsToKeep)
	numOfActivePersisters := uint32(psf.generalConfig.StoragePruning.NumActivePersisters)
	pruningEnabled := psf.generalConfig.StoragePruning.Enabled
	shardId := core.GetShardIdString(psf.shardCoordinator.SelfId())
	dbPath := filepath.Join(psf.pathManager.PathForEpoch(shardId, psf.currentEpoch, storageConfig.DB.FilePath))
	args := &pruning.StorerArgs{
		Identifier:            storageConfig.DB.FilePath,
		PruningEnabled:        pruningEnabled,
		StartingEpoch:         psf.currentEpoch,
		FullArchive:           fullArchiveMode,
		ShardCoordinator:      psf.shardCoordinator,
		CacheConf:             GetCacherFromConfig(storageConfig.Cache),
		PathManager:           psf.pathManager,
		DbPath:                dbPath,
		PersisterFactory:      NewPersisterFactory(storageConfig.DB),
		BloomFilterConf:       GetBloomFromConfig(storageConfig.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}

	return args
}
