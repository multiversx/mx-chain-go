package factory

import (
	"path/filepath"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/pruning"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.GetOrCreate("storage/factory")

// StorageServiceFactory handles the creation of storage services for both meta and shards
type StorageServiceFactory struct {
	generalConfig      *config.Config
	shardCoordinator   sharding.Coordinator
	uniqueID           string
	epochStartNotifier storage.EpochStartNotifier
}

// NewStorageServiceFactory will return a new instance of StorageServiceFactory
func NewStorageServiceFactory(
	config *config.Config,
	shardCoordinator sharding.Coordinator,
	uniqueID string,
	epochStartNotifier storage.EpochStartNotifier,
) (*StorageServiceFactory, error) {
	if config == nil {
		return nil, storage.ErrNilConfig
	}
	if check.IfNil(shardCoordinator) {
		return nil, storage.ErrNilShardCoordinator
	}
	if len(uniqueID) == 0 {
		return nil, storage.ErrEmptyUniqueIdentifier
	}
	if check.IfNil(epochStartNotifier) {
		return nil, storage.ErrNilEpochStartNotifier
	}

	return &StorageServiceFactory{
		generalConfig:      config,
		shardCoordinator:   shardCoordinator,
		uniqueID:           uniqueID,
		epochStartNotifier: epochStartNotifier,
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

	txUnitStorerArgs := psf.createPruningStorerArgs("txUnit", psf.generalConfig.TxStorage)
	txUnit, err = pruning.NewPruningStorer(txUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitStorerArgs := psf.createPruningStorerArgs("unsignedTxUnit", psf.generalConfig.UnsignedTransactionStorage)
	unsignedTxUnit, err = pruning.NewPruningStorer(unsignedTxUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	rewardTxUnitArgs := psf.createPruningStorerArgs("rewardTxUnit", psf.generalConfig.RewardTxStorage)
	rewardTxUnit, err = pruning.NewPruningStorer(rewardTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, rewardTxUnit)

	miniBlockUnitArgs := psf.createPruningStorerArgs("miniBlockUnit", psf.generalConfig.MiniBlocksStorage)
	miniBlockUnit, err = pruning.NewPruningStorer(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	peerBlockUnitArgs := psf.createPruningStorerArgs("peerBlockUnit", psf.generalConfig.PeerBlockBodyStorage)
	peerBlockUnit, err = pruning.NewPruningStorer(peerBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, peerBlockUnit)

	headerUnitArgs := psf.createPruningStorerArgs("headerUnit", psf.generalConfig.BlockHeaderStorage)
	headerUnit, err = pruning.NewPruningStorer(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	metaChainHeaderUnitArgs := psf.createPruningStorerArgs("metachainHeaderUnit", psf.generalConfig.MetaBlockStorage)
	metachainHeaderUnit, err = pruning.NewPruningStorer(metaChainHeaderUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metachainHeaderUnit)

	metaHdrHashNonceUnitArgs := psf.createPruningStorerArgs("metaHdrHashNonceUnit", psf.generalConfig.MetaHdrNonceHashStorage)
	metaHdrHashNonceUnit, err = pruning.NewPruningStorer(metaHdrHashNonceUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	shardHdrHashNonceUnitArgs := psf.createPruningStorerArgs("shardHrHashNonceUnit", psf.generalConfig.ShardHdrNonceHashStorage)
	shardHdrHashNonceUnit, err = pruning.NewShardedPruningStorer(shardHdrHashNonceUnitArgs, psf.shardCoordinator.SelfId())
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, shardHdrHashNonceUnit)

	// TODO: the path will be fetched from a path naming component
	heartbeatUniqueID := strings.Replace(psf.uniqueID, "Epoch_0", "Static", 1)
	heartbeatStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Cache),
		GetDBFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.DB, heartbeatUniqueID),
		GetBloomFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, heartbeatStorageUnit)

	bootstrapUnitArgs := psf.createPruningStorerArgs("bootstrapUnit", psf.generalConfig.BootstrapStorage)
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

	return store, err
}

// CreateForMeta will return the storage service which contains all storers needed for metachain
func (psf *StorageServiceFactory) CreateForMeta() (dataRetriever.StorageService, error) {
	var shardDataUnit *pruning.PruningStorer
	var metaBlockUnit *pruning.PruningStorer
	var headerUnit *pruning.PruningStorer
	var txUnit *pruning.PruningStorer
	var peerDataUnit *pruning.PruningStorer
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

	metaBlockUnitArgs := psf.createPruningStorerArgs("metaBlockUnit", psf.generalConfig.MetaBlockStorage)
	metaBlockUnit, err = pruning.NewPruningStorer(metaBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaBlockUnit)

	shardDataUnitArgs := psf.createPruningStorerArgs("shardDataUnit", psf.generalConfig.ShardDataStorage)
	shardDataUnit, err = pruning.NewPruningStorer(shardDataUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, shardDataUnit)

	peerDataUnitArgs := psf.createPruningStorerArgs("peerDataUnit", psf.generalConfig.PeerDataStorage)
	peerDataUnit, err = pruning.NewPruningStorer(peerDataUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, peerDataUnit)

	headerUnitArgs := psf.createPruningStorerArgs("headerUnit", psf.generalConfig.BlockHeaderStorage)
	headerUnit, err = pruning.NewPruningStorer(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	metaHdrHashNonceUnitArgs := psf.createPruningStorerArgs("metaHdrHashNonceUnit", psf.generalConfig.MetaHdrNonceHashStorage)
	metaHdrHashNonceUnit, err = pruning.NewPruningStorer(metaHdrHashNonceUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	shardHdrHashNonceUnits = make([]*pruning.PruningStorer, psf.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceUnitArgs := psf.createPruningStorerArgs("shardHdrHashNonceUnit",
			psf.generalConfig.ShardHdrNonceHashStorage)
		shardHdrHashNonceUnits[i], err = pruning.NewShardedPruningStorer(shardHdrHashNonceUnitArgs, i)
		if err != nil {
			return nil, err
		}
		successfullyCreatedStorers = append(successfullyCreatedStorers, shardHdrHashNonceUnits[i])
	}

	// TODO: the path will be fetched from a path naming component
	heartbeatUniqueID := strings.Replace(psf.uniqueID, "Epoch_0", "Static", 1)
	heartbeatStorageUnit, err := storageUnit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Cache),
		GetDBFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.DB, heartbeatUniqueID),
		GetBloomFromConfig(psf.generalConfig.Heartbeat.HeartbeatStorage.Bloom))
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, heartbeatStorageUnit)

	txUnitArgs := psf.createPruningStorerArgs("txUnit", psf.generalConfig.TxStorage)
	txUnit, err = pruning.NewPruningStorer(txUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitArgs := psf.createPruningStorerArgs("unsignedTxUnit", psf.generalConfig.UnsignedTransactionStorage)
	unsignedTxUnit, err = pruning.NewPruningStorer(unsignedTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	miniBlockUnitArgs := psf.createPruningStorerArgs("miniBlockUnit", psf.generalConfig.MiniBlocksStorage)
	miniBlockUnit, err = pruning.NewPruningStorer(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	miniBlockHeadersUnitArgs := psf.createPruningStorerArgs("miniBlockHeadersUnit", psf.generalConfig.MiniBlockHeadersStorage)
	miniBlockHeadersUnit, err = pruning.NewPruningStorer(miniBlockHeadersUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockHeadersUnit)

	bootstrapUnitArgs := psf.createPruningStorerArgs("bootstrapUnit", psf.generalConfig.BootstrapStorage)
	bootstrapUnit, err = pruning.NewPruningStorer(bootstrapUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, bootstrapUnit)

	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)
	store.AddStorer(dataRetriever.MetaShardDataUnit, shardDataUnit)
	store.AddStorer(dataRetriever.MetaPeerDataUnit, peerDataUnit)
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

	return store, err
}

func (psf *StorageServiceFactory) createPruningStorerArgs(identifier string, storageConfig config.StorageConfig) *pruning.PruningStorerArgs {
	fullArchiveMode := psf.generalConfig.StoragePruning.FullArchive
	numOfEpochsToKeep := uint32(psf.generalConfig.StoragePruning.NumEpochsToKeep)
	numOfActivePersisters := uint32(psf.generalConfig.StoragePruning.NumActivePersisters)

	args := &pruning.PruningStorerArgs{
		Identifier:            identifier,
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(storageConfig.Cache),
		DbPath:                filepath.Join(psf.uniqueID, storageConfig.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(storageConfig.DB),
		BloomFilterConf:       GetBloomFromConfig(storageConfig.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}

	return args
}
