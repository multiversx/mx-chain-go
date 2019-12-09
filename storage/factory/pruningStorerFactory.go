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

	fullArchiveMode := psf.generalConfig.StoragePruning.FullArchive
	numOfEpochsToKeep := uint32(psf.generalConfig.StoragePruning.NumEpochsToKeep)
	numOfActivePersisters := uint32(psf.generalConfig.StoragePruning.NumActivePersisters)

	txUnitStorerArgs := &pruning.PruningStorerArgs{
		Identifier:            "txUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.TxStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.TxStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.TxStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.TxStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	txUnit, err = pruning.NewPruningStorer(txUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitStorerArgs := &pruning.PruningStorerArgs{
		Identifier:            "unsignedTxUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.UnsignedTransactionStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.UnsignedTransactionStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.UnsignedTransactionStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.UnsignedTransactionStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	unsignedTxUnit, err = pruning.NewPruningStorer(unsignedTxUnitStorerArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	rewardTxUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "rewardTxUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.RewardTxStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.RewardTxStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.RewardTxStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.RewardTxStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	rewardTxUnit, err = pruning.NewPruningStorer(rewardTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, rewardTxUnit)

	miniBlockUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "miniBlockUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.MiniBlocksStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.MiniBlocksStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.MiniBlocksStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.MiniBlocksStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	miniBlockUnit, err = pruning.NewPruningStorer(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	peerBlockUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "peerBlockUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.PeerBlockBodyStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.PeerBlockBodyStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.PeerBlockBodyStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.PeerBlockBodyStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	peerBlockUnit, err = pruning.NewPruningStorer(peerBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, peerBlockUnit)

	headerUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "headerUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.BlockHeaderStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.BlockHeaderStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.BlockHeaderStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.BlockHeaderStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	headerUnit, err = pruning.NewPruningStorer(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	metaChainHeaderUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "metachainHeaderUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.MetaBlockStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.MetaBlockStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.MetaBlockStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.MetaBlockStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	metachainHeaderUnit, err = pruning.NewPruningStorer(metaChainHeaderUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metachainHeaderUnit)

	metaHdrHashNonceUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "metaHdrHashNonceUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.MetaHdrNonceHashStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.MetaHdrNonceHashStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	metaHdrHashNonceUnit, err = pruning.NewPruningStorer(metaHdrHashNonceUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	shardHdrHashNonceUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "shardHrHashNonceUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.ShardHdrNonceHashStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.ShardHdrNonceHashStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
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

	bootstrapUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "bootstrapUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.BootstrapStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.BootstrapStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.BootstrapStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.BootstrapStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
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
			if shardHdrHashNonceUnits != nil {
				for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
					if shardHdrHashNonceUnits[i] != nil {
						_ = shardHdrHashNonceUnits[i].DestroyUnit()
					}
				}
			}
		}
	}()

	fullArchiveMode := psf.generalConfig.StoragePruning.FullArchive
	numOfEpochsToKeep := uint32(psf.generalConfig.StoragePruning.NumEpochsToKeep)
	numOfActivePersisters := uint32(psf.generalConfig.StoragePruning.NumActivePersisters)

	metaBlockUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "metaBlockUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.MetaBlockStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.MetaBlockStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.MetaBlockStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.MetaBlockStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	metaBlockUnit, err = pruning.NewPruningStorer(metaBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaBlockUnit)

	shardDataUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "shardDataUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.ShardDataStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.ShardDataStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.ShardDataStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.ShardDataStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	shardDataUnit, err = pruning.NewPruningStorer(shardDataUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, shardDataUnit)

	peerDataUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "peerDataUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.PeerDataStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.PeerDataStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.PeerDataStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.PeerDataStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	peerDataUnit, err = pruning.NewPruningStorer(peerDataUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, peerDataUnit)

	headerUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "headerUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.BlockHeaderStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.BlockHeaderStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.BlockHeaderStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.BlockHeaderStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	headerUnit, err = pruning.NewPruningStorer(headerUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, headerUnit)

	metaHdrHashNonceUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "metaHdrHashNonceUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.MetaHdrNonceHashStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.MetaHdrNonceHashStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	metaHdrHashNonceUnit, err = pruning.NewPruningStorer(metaHdrHashNonceUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, metaHdrHashNonceUnit)

	shardHdrHashNonceUnits = make([]*pruning.PruningStorer, psf.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceUnitArgs := &pruning.PruningStorerArgs{
			Identifier:            "shardHdrHashNonceUnit",
			FullArchive:           fullArchiveMode,
			CacheConf:             GetCacherFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Cache),
			DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.ShardHdrNonceHashStorage.DB.FilePath),
			PersisterFactory:      NewPersisterFactory(psf.generalConfig.ShardHdrNonceHashStorage.DB),
			BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Bloom),
			NumOfEpochsToKeep:     numOfEpochsToKeep,
			NumOfActivePersisters: numOfActivePersisters,
			Notifier:              psf.epochStartNotifier,
		}
		shardHdrHashNonceUnits[i], err = pruning.NewShardedPruningStorer(shardHdrHashNonceUnitArgs, i)
		if err != nil {
			return nil, err
		}
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

	txUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "txUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.TxStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.TxStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.TxStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.TxStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	txUnit, err = pruning.NewPruningStorer(txUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, txUnit)

	unsignedTxUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "unsignedTxUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.UnsignedTransactionStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.UnsignedTransactionStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.UnsignedTransactionStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.UnsignedTransactionStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	unsignedTxUnit, err = pruning.NewPruningStorer(unsignedTxUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, unsignedTxUnit)

	miniBlockUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "miniBlockUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.MiniBlocksStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.MiniBlocksStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.MiniBlocksStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.MiniBlocksStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	miniBlockUnit, err = pruning.NewPruningStorer(miniBlockUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockUnit)

	miniBlockHeadersUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "miniBlockHeadersUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.MiniBlockHeadersStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.MiniBlockHeadersStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.MiniBlockHeadersStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.MiniBlockHeadersStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
	miniBlockHeadersUnit, err = pruning.NewPruningStorer(miniBlockHeadersUnitArgs)
	if err != nil {
		return nil, err
	}
	successfullyCreatedStorers = append(successfullyCreatedStorers, miniBlockHeadersUnit)

	bootstrapUnitArgs := &pruning.PruningStorerArgs{
		Identifier:            "bootstrapUnit",
		FullArchive:           fullArchiveMode,
		CacheConf:             GetCacherFromConfig(psf.generalConfig.BootstrapStorage.Cache),
		DbPath:                filepath.Join(psf.uniqueID, psf.generalConfig.BootstrapStorage.DB.FilePath),
		PersisterFactory:      NewPersisterFactory(psf.generalConfig.BootstrapStorage.DB),
		BloomFilterConf:       GetBloomFromConfig(psf.generalConfig.BootstrapStorage.Bloom),
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
		Notifier:              psf.epochStartNotifier,
	}
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
