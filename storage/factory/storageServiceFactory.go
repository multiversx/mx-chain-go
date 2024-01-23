package factory

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/clean"
	"github.com/multiversx/mx-chain-go/storage/databaseremover/disabled"
	"github.com/multiversx/mx-chain-go/storage/databaseremover/factory"
	storageDisabled "github.com/multiversx/mx-chain-go/storage/disabled"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("storage/factory")

const (
	minimumNumberOfActivePersisters = 1
	minimumNumberOfEpochsToKeep     = 2
)

// StorageServiceType defines the type of StorageService
type StorageServiceType string

const (
	// BootstrapStorageService is used when the node is bootstrapping
	BootstrapStorageService StorageServiceType = "bootstrap"

	// ProcessStorageService is used in normal processing
	ProcessStorageService StorageServiceType = "process"

	// ImportDBStorageService is used for the import-db storage service
	ImportDBStorageService StorageServiceType = "import-db"
)

// StorageServiceFactory handles the creation of storage services for both meta and shards
type StorageServiceFactory struct {
	generalConfig                 config.Config
	prefsConfig                   config.PreferencesConfig
	shardCoordinator              storage.ShardCoordinator
	pathManager                   storage.PathManagerHandler
	epochStartNotifier            epochStart.EpochStartNotifier
	oldDataCleanerProvider        clean.OldDataCleanerProvider
	createTrieEpochRootHashStorer bool
	currentEpoch                  uint32
	storageType                   StorageServiceType
	nodeProcessingMode            common.NodeProcessingMode
	snapshotsEnabled              bool
	repopulateTokensSupplies      bool
	stateStatsHandler             common.StateStatisticsHandler
	chainRunType                  common.ChainRunType
	additionalStorageServiceCreator process.AdditionalStorageServiceCreator
}

// StorageServiceFactoryArgs holds the arguments needed for creating a new storage service factory
type StorageServiceFactoryArgs struct {
	Config                        config.Config
	PrefsConfig                   config.PreferencesConfig
	ShardCoordinator              storage.ShardCoordinator
	PathManager                   storage.PathManagerHandler
	EpochStartNotifier            epochStart.EpochStartNotifier
	NodeTypeProvider              NodeTypeProviderHandler
	StorageType                   StorageServiceType
	ManagedPeersHolder            storage.ManagedPeersHolder
	CurrentEpoch                  uint32
	CreateTrieEpochRootHashStorer bool
	NodeProcessingMode            common.NodeProcessingMode
	RepopulateTokensSupplies      bool
	StateStatsHandler             common.StateStatisticsHandler
	ChainRunType                  common.ChainRunType
	AdditionalStorageServiceCreator process.AdditionalStorageServiceCreator
}

// NewStorageServiceFactory will return a new instance of StorageServiceFactory
func NewStorageServiceFactory(args StorageServiceFactoryArgs) (*StorageServiceFactory, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	argsOldDataCleanerProvider := clean.ArgOldDataCleanerProvider{
		NodeTypeProvider:    args.NodeTypeProvider,
		PruningStorerConfig: args.Config.StoragePruning,
		ManagedPeersHolder:  args.ManagedPeersHolder,
	}
	oldDataCleanProvider, err := clean.NewOldDataCleanerProvider(argsOldDataCleanerProvider)
	if err != nil {
		return nil, err
	}
	if args.Config.StoragePruning.NumEpochsToKeep < minimumNumberOfEpochsToKeep && oldDataCleanProvider.ShouldClean() {
		return nil, storage.ErrInvalidNumberOfEpochsToSave
	}

	return &StorageServiceFactory{
		generalConfig:                 args.Config,
		prefsConfig:                   args.PrefsConfig,
		shardCoordinator:              args.ShardCoordinator,
		pathManager:                   args.PathManager,
		epochStartNotifier:            args.EpochStartNotifier,
		currentEpoch:                  args.CurrentEpoch,
		createTrieEpochRootHashStorer: args.CreateTrieEpochRootHashStorer,
		oldDataCleanerProvider:        oldDataCleanProvider,
		storageType:                   args.StorageType,
		nodeProcessingMode:            args.NodeProcessingMode,
		snapshotsEnabled:              args.Config.StateTriesConfig.SnapshotsEnabled,
		repopulateTokensSupplies:      args.RepopulateTokensSupplies,
		stateStatsHandler:             args.StateStatsHandler,
		chainRunType:                  args.ChainRunType,
		additionalStorageServiceCreator: args.AdditionalStorageServiceCreator,
	}, nil
}

func checkArgs(args StorageServiceFactoryArgs) error {
	if args.Config.StoragePruning.NumActivePersisters < minimumNumberOfActivePersisters {
		return storage.ErrInvalidNumberOfActivePersisters
	}
	if check.IfNil(args.ShardCoordinator) {
		return storage.ErrNilShardCoordinator
	}
	if check.IfNil(args.PathManager) {
		return storage.ErrNilPathManager
	}
	if check.IfNil(args.EpochStartNotifier) {
		return storage.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.StateStatsHandler) {
		return statistics.ErrNilStateStatsHandler
	}
	if check.IfNil(args.AdditionalStorageServiceCreator) {
		return errors.ErrNilAdditionalStorageServiceCreator
	}

	return nil
}

// TODO: refactor this function, split it into multiple ones
func (psf *StorageServiceFactory) createAndAddBaseStorageUnits(
	store dataRetriever.StorageService,
	customDatabaseRemover storage.CustomDatabaseRemoverHandler,
	shardID string,
) error {
	disabledCustomDatabaseRemover := disabled.NewDisabledCustomDatabaseRemover()

	txUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.TxStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	txUnit, err := psf.createPruningPersister(txUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for TxStorage", err)
	}
	store.AddStorer(dataRetriever.TransactionUnit, txUnit)

	unsignedTxUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.UnsignedTransactionStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	unsignedTxUnit, err := psf.createPruningPersister(unsignedTxUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for UnsignedTransactionStorage", err)
	}
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, unsignedTxUnit)

	rewardTxUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.RewardTxStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	rewardTxUnit, err := psf.createPruningPersister(rewardTxUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for RewardTxStorage", err)
	}
	store.AddStorer(dataRetriever.RewardTransactionUnit, rewardTxUnit)

	receiptsUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.ReceiptsStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	receiptsUnit, err := psf.createPruningPersister(receiptsUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for ReceiptsStorage", err)
	}
	store.AddStorer(dataRetriever.ReceiptsUnit, receiptsUnit)

	scheduledSCRsUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.ScheduledSCRsStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	scheduledSCRsUnit, err := psf.createPruningPersister(scheduledSCRsUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for ScheduledSCRsStorage", err)
	}
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, scheduledSCRsUnit)

	bootstrapUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.BootstrapStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	bootstrapUnit, err := psf.createPruningPersister(bootstrapUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for BootstrapStorage", err)
	}
	store.AddStorer(dataRetriever.BootstrapUnit, bootstrapUnit)

	miniBlockUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.MiniBlocksStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	miniBlockUnit, err := psf.createPruningPersister(miniBlockUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for MiniBlocksStorage", err)
	}
	store.AddStorer(dataRetriever.MiniBlockUnit, miniBlockUnit)

	metaBlockUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.MetaBlockStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	metaBlockUnit, err := psf.createPruningPersister(metaBlockUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for MetaBlockStorage", err)
	}
	store.AddStorer(dataRetriever.MetaBlockUnit, metaBlockUnit)

	// metaHdrHashNonce is static
	metaHdrHashNonceUnitConfig := GetDBFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.DB)
	dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.MetaHdrNonceHashStorage.DB.FilePath)
	metaHdrHashNonceUnitConfig.FilePath = dbPath

	dbConfigHandlerInstance := NewDBConfigHandler(psf.generalConfig.MetaHdrNonceHashStorage.DB)
	metaHdrHashNoncePersisterCreator, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return err
	}

	metaHdrHashNonceUnit, err := storageunit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.MetaHdrNonceHashStorage.Cache),
		metaHdrHashNonceUnitConfig,
		metaHdrHashNoncePersisterCreator,
	)
	if err != nil {
		return fmt.Errorf("%w for MetaHdrNonceHashStorage", err)
	}
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, metaHdrHashNonceUnit)

	headerUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.BlockHeaderStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	headerUnit, err := psf.createPruningPersister(headerUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for BlockHeaderStorage", err)
	}
	store.AddStorer(dataRetriever.BlockHeaderUnit, headerUnit)

	userAccountsUnit, err := psf.createTrieStorer(psf.generalConfig.AccountsTrieStorage, customDatabaseRemover)
	if err != nil {
		return fmt.Errorf("%w for AccountsTrieStorage", err)
	}
	store.AddStorer(dataRetriever.UserAccountsUnit, userAccountsUnit)

	statusMetricsDbConfig := GetDBFromConfig(psf.generalConfig.StatusMetricsStorage.DB)
	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath = psf.pathManager.PathForStatic(shardId, psf.generalConfig.StatusMetricsStorage.DB.FilePath)
	statusMetricsDbConfig.FilePath = dbPath

	dbConfigHandlerInstance = NewDBConfigHandler(psf.generalConfig.StatusMetricsStorage.DB)
	statusMetricsPersisterCreator, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return err
	}

	statusMetricsStorageUnit, err := storageunit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.StatusMetricsStorage.Cache),
		statusMetricsDbConfig,
		statusMetricsPersisterCreator,
	)
	if err != nil {
		return fmt.Errorf("%w for StatusMetricsStorage", err)
	}
	store.AddStorer(dataRetriever.StatusMetricsUnit, statusMetricsStorageUnit)

	trieEpochRootHashStorageUnit, err := psf.createTrieEpochRootHashStorerIfNeeded()
	if err != nil {
		return err
	}
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit, trieEpochRootHashStorageUnit)

	return nil
}

func (psf *StorageServiceFactory) createAndAddStorageUnitsForSovereign(
	store dataRetriever.StorageService,
	shardID string,
) error {
	disabledCustomDatabaseRemover := disabled.NewDisabledCustomDatabaseRemover()

	extendedHeaderUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.SovereignConfig.ExtendedShardHeaderStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return err
	}
	extendedHeaderUnit, err := psf.createPruningPersister(extendedHeaderUnitArgs)
	if err != nil {
		return fmt.Errorf("%w for ExtendedShardHeaderStorage", err)
	}
	store.AddStorer(dataRetriever.ExtendedShardHeadersUnit, extendedHeaderUnit)

	extendedShardHdrHashNonceConfig := GetDBFromConfig(psf.generalConfig.SovereignConfig.ExtendedShardHdrNonceHashStorage.DB)
	dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.SovereignConfig.ExtendedShardHdrNonceHashStorage.DB.FilePath) + shardID
	extendedShardHdrHashNonceConfig.FilePath = dbPath

	extendedHeaderConfig := psf.generalConfig.SovereignConfig.ExtendedShardHeaderStorage
	dbConfigExtendedHeader := NewDBConfigHandler(extendedHeaderConfig.DB)
	extendedHeaderPersisterCreator, err := NewPersisterFactory(dbConfigExtendedHeader)
	if err != nil {
		return err
	}

	extendedShardHdrHashNonceUnit, err := storageunit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.SovereignConfig.ExtendedShardHdrNonceHashStorage.Cache),
		extendedShardHdrHashNonceConfig,
		extendedHeaderPersisterCreator,
	)
	if err != nil {
		return fmt.Errorf("%w for ExtendedShardHdrNonceHashStorage", err)
	}
	store.AddStorer(dataRetriever.ExtendedShardHeadersNonceHashDataUnit, extendedShardHdrHashNonceUnit)

	return nil
}

// CreateForShard will return the storage service which contains all storers needed for a shard
func (psf *StorageServiceFactory) CreateForShard() (dataRetriever.StorageService, error) {
	// TODO: if there will be a differentiation between the creation or opening of a DB, the DBs could be destroyed on a defer
	// in case of a failure while creating (not opening).

	disabledCustomDatabaseRemover := disabled.NewDisabledCustomDatabaseRemover()
	customDatabaseRemover, err := factory.CreateCustomDatabaseRemover(psf.generalConfig.StoragePruning)
	if err != nil {
		return nil, err
	}

	shardID := core.GetShardIDString(psf.shardCoordinator.SelfId())

	// shardHdrHashNonce storer is static
	shardHdrHashNonceConfig := GetDBFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.DB)
	dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.ShardHdrNonceHashStorage.DB.FilePath) + shardID
	shardHdrHashNonceConfig.FilePath = dbPath

	dbConfigHandlerInstance := NewDBConfigHandler(psf.generalConfig.ShardHdrNonceHashStorage.DB)
	shardHdrHashNoncePersisterCreator, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return nil, err
	}

	shardHdrHashNonceUnit, err := storageunit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Cache),
		shardHdrHashNonceConfig,
		shardHdrHashNoncePersisterCreator,
	)
	if err != nil {
		return nil, fmt.Errorf("%w for ShardHdrNonceHashStorage", err)
	}

	store := dataRetriever.NewChainStorer()
	err = psf.createAndAddBaseStorageUnits(store, customDatabaseRemover, shardID)
	if err != nil {
		return nil, err
	}

	// TODO: this should be refactored to not get a function as a parameter
	err = psf.additionalStorageServiceCreator.CreateAdditionalStorageUnits(psf.createAndAddStorageUnitsForSovereign, store, shardID)
	if err != nil {
		return nil, err
	}

	peerAccountsUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.PeerAccountsTrieStorage, customDatabaseRemover)
	if err != nil {
		return nil, err
	}
	peerAccountsUnit, err := psf.createTrieUnit(psf.generalConfig.PeerAccountsTrieStorage, peerAccountsUnitArgs)
	if err != nil {
		return nil, fmt.Errorf("%w for PeerAccountsTrieStorage", err)
	}
	store.AddStorer(dataRetriever.PeerAccountsUnit, peerAccountsUnit)

	peerBlockUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.PeerBlockBodyStorage, disabledCustomDatabaseRemover)
	if err != nil {
		return nil, err
	}
	peerBlockUnit, err := psf.createPruningPersister(peerBlockUnitArgs)
	if err != nil {
		return nil, fmt.Errorf("%w for PeerBlockBodyStorage", err)
	}
	store.AddStorer(dataRetriever.PeerChangesUnit, peerBlockUnit)

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(psf.shardCoordinator.SelfId())
	store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnit)

	err = psf.setUpDbLookupExtensions(store)
	if err != nil {
		return nil, err
	}

	err = psf.setUpLogsAndEventsStorer(store)
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
	// TODO: if there will be a differentiation between the creation or opening of a DB, the DBs could be destroyed on a defer
	// in case of a failure while creating (not opening)

	customDatabaseRemover, err := factory.CreateCustomDatabaseRemover(psf.generalConfig.StoragePruning)
	if err != nil {
		return nil, err
	}
	shardID := core.GetShardIDString(core.MetachainShardId)

	shardHdrHashNonceUnits := make([]*storageunit.Unit, psf.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		shardHdrHashNonceConfig := GetDBFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.DB)
		shardID = core.GetShardIDString(core.MetachainShardId)
		dbPath := psf.pathManager.PathForStatic(shardID, psf.generalConfig.ShardHdrNonceHashStorage.DB.FilePath) + fmt.Sprintf("%d", i)
		shardHdrHashNonceConfig.FilePath = dbPath

		dbConfigHandlerInstance := NewDBConfigHandler(psf.generalConfig.ShardHdrNonceHashStorage.DB)
		shardHdrHashNoncePersisterCreator, errLoop := NewPersisterFactory(dbConfigHandlerInstance)
		if errLoop != nil {
			return nil, errLoop
		}

		shardHdrHashNonceUnits[i], errLoop = storageunit.NewStorageUnitFromConf(
			GetCacherFromConfig(psf.generalConfig.ShardHdrNonceHashStorage.Cache),
			shardHdrHashNonceConfig,
			shardHdrHashNoncePersisterCreator,
		)
		if errLoop != nil {
			return nil, fmt.Errorf("%w for ShardHdrNonceHashStorage on shard %d", errLoop, i)
		}
	}

	store := dataRetriever.NewChainStorer()
	err = psf.createAndAddBaseStorageUnits(store, customDatabaseRemover, shardID)
	if err != nil {
		return nil, err
	}

	peerAccountsUnit, err := psf.createTrieStorer(psf.generalConfig.PeerAccountsTrieStorage, customDatabaseRemover)
	if err != nil {
		return nil, err
	}
	store.AddStorer(dataRetriever.PeerAccountsUnit, peerAccountsUnit)

	for i := uint32(0); i < psf.shardCoordinator.NumberOfShards(); i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		store.AddStorer(hdrNonceHashDataUnit, shardHdrHashNonceUnits[i])
	}

	err = psf.setUpDbLookupExtensions(store)
	if err != nil {
		return nil, err
	}

	err = psf.setUpLogsAndEventsStorer(store)
	if err != nil {
		return nil, err
	}

	err = psf.initOldDatabasesCleaningIfNeeded(store)
	if err != nil {
		return nil, err
	}

	return store, err
}

func (psf *StorageServiceFactory) createTrieStorer(
	storageConfig config.StorageConfig,
	customDatabaseRemover storage.CustomDatabaseRemoverHandler,
) (storage.Storer, error) {
	accountsUnitArgs, err := psf.createPruningStorerArgs(storageConfig, customDatabaseRemover)
	if err != nil {
		return nil, err
	}

	if psf.storageType == ProcessStorageService && psf.nodeProcessingMode == common.Normal {
		accountsUnitArgs.PersistersTracker = pruning.NewTriePersisterTracker(accountsUnitArgs.EpochsData)
	}

	return psf.createTrieUnit(storageConfig, accountsUnitArgs)
}

func (psf *StorageServiceFactory) createTrieUnit(
	storageConfig config.StorageConfig,
	pruningStorageArgs pruning.StorerArgs,
) (storage.Storer, error) {
	if psf.storageType == ImportDBStorageService {
		return storageDisabled.NewStorer(), nil
	}

	if !psf.snapshotsEnabled {
		return psf.createTriePersister(storageConfig)
	}

	return psf.createTriePruningPersister(pruningStorageArgs)
}

func (psf *StorageServiceFactory) setUpLogsAndEventsStorer(chainStorer *dataRetriever.ChainStorer) error {
	var txLogsUnit storage.Storer
	txLogsUnit = storageDisabled.NewStorer()

	// Should not create logs and events storer in the next case:
	// - LogsAndEvents.Enabled = false and DbLookupExtensions.Enabled = false
	// If we have DbLookupExtensions ACTIVE node by default should save logs no matter if is enabled or not
	shouldCreateStorer := psf.generalConfig.LogsAndEvents.SaveInStorageEnabled || psf.generalConfig.DbLookupExtensions.Enabled
	if shouldCreateStorer {
		var err error
		txLogsUnitArgs, err := psf.createPruningStorerArgs(psf.generalConfig.LogsAndEvents.TxLogsStorage, disabled.NewDisabledCustomDatabaseRemover())
		if err != nil {
			return err
		}
		txLogsUnit, err = psf.createPruningPersister(txLogsUnitArgs)
		if err != nil {
			return fmt.Errorf("%w for LogsAndEvents.TxLogsStorage", err)
		}
	}

	chainStorer.AddStorer(dataRetriever.TxLogsUnit, txLogsUnit)

	return nil
}

func (psf *StorageServiceFactory) setUpDbLookupExtensions(chainStorer *dataRetriever.ChainStorer) error {
	if !psf.generalConfig.DbLookupExtensions.Enabled {
		return nil
	}

	shardID := core.GetShardIDString(psf.shardCoordinator.SelfId())

	// Create the eventsHashesByTxHash (PRUNING) storer
	eventsHashesByTxHashConfig := psf.generalConfig.DbLookupExtensions.ResultsHashesByTxHashStorageConfig
	eventsHashesByTxHashStorerArgs, err := psf.createPruningStorerArgs(eventsHashesByTxHashConfig, disabled.NewDisabledCustomDatabaseRemover())
	if err != nil {
		return err
	}
	eventsHashesByTxHashPruningStorer, err := psf.createPruningPersister(eventsHashesByTxHashStorerArgs)
	if err != nil {
		return fmt.Errorf("%w for DbLookupExtensions.ResultsHashesByTxHashStorageConfig", err)
	}

	chainStorer.AddStorer(dataRetriever.ResultsHashesByTxHashUnit, eventsHashesByTxHashPruningStorer)

	// Create the miniblocksMetadata (PRUNING) storer
	miniblocksMetadataConfig := psf.generalConfig.DbLookupExtensions.MiniblocksMetadataStorageConfig
	miniblocksMetadataPruningStorerArgs, err := psf.createPruningStorerArgs(miniblocksMetadataConfig, disabled.NewDisabledCustomDatabaseRemover())
	if err != nil {
		return err
	}
	miniblocksMetadataPruningStorer, err := psf.createPruningPersister(miniblocksMetadataPruningStorerArgs)
	if err != nil {
		return fmt.Errorf("%w for DbLookupExtensions.MiniblocksMetadataStorageConfig", err)
	}

	chainStorer.AddStorer(dataRetriever.MiniblocksMetadataUnit, miniblocksMetadataPruningStorer)

	// Create the miniblocksHashByTxHash (STATIC) storer
	miniblockHashByTxHashConfig := psf.generalConfig.DbLookupExtensions.MiniblockHashByTxHashStorageConfig
	miniblockHashByTxHashDbConfig := GetDBFromConfig(miniblockHashByTxHashConfig.DB)
	miniblockHashByTxHashDbConfig.FilePath = psf.pathManager.PathForStatic(shardID, miniblockHashByTxHashConfig.DB.FilePath)
	miniblockHashByTxHashCacherConfig := GetCacherFromConfig(miniblockHashByTxHashConfig.Cache)

	dbConfigHandlerInstance := NewDBConfigHandler(miniblockHashByTxHashConfig.DB)
	miniblockHashByTxHashPersisterCreator, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return err
	}

	miniblockHashByTxHashUnit, err := storageunit.NewStorageUnitFromConf(
		miniblockHashByTxHashCacherConfig,
		miniblockHashByTxHashDbConfig,
		miniblockHashByTxHashPersisterCreator,
	)
	if err != nil {
		return fmt.Errorf("%w for DbLookupExtensions.MiniblockHashByTxHashStorageConfig", err)
	}

	chainStorer.AddStorer(dataRetriever.MiniblockHashByTxHashUnit, miniblockHashByTxHashUnit)

	// Create the blockHashByRound (STATIC) storer
	blockHashByRoundConfig := psf.generalConfig.DbLookupExtensions.RoundHashStorageConfig
	blockHashByRoundDBConfig := GetDBFromConfig(blockHashByRoundConfig.DB)
	blockHashByRoundDBConfig.FilePath = psf.pathManager.PathForStatic(shardID, blockHashByRoundConfig.DB.FilePath)
	blockHashByRoundCacherConfig := GetCacherFromConfig(blockHashByRoundConfig.Cache)

	dbConfigHandlerInstance = NewDBConfigHandler(blockHashByRoundConfig.DB)
	blockHashByRoundPersisterCreator, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return err
	}

	blockHashByRoundUnit, err := storageunit.NewStorageUnitFromConf(
		blockHashByRoundCacherConfig,
		blockHashByRoundDBConfig,
		blockHashByRoundPersisterCreator,
	)
	if err != nil {
		return fmt.Errorf("%w for DbLookupExtensions.RoundHashStorageConfig", err)
	}

	chainStorer.AddStorer(dataRetriever.RoundHdrHashDataUnit, blockHashByRoundUnit)

	// Create the epochByHash (STATIC) storer
	epochByHashConfig := psf.generalConfig.DbLookupExtensions.EpochByHashStorageConfig
	epochByHashDbConfig := GetDBFromConfig(epochByHashConfig.DB)
	epochByHashDbConfig.FilePath = psf.pathManager.PathForStatic(shardID, epochByHashConfig.DB.FilePath)
	epochByHashCacherConfig := GetCacherFromConfig(epochByHashConfig.Cache)

	dbConfigHandlerInstance = NewDBConfigHandler(epochByHashConfig.DB)
	epochByHashPersisterCreator, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return err
	}

	epochByHashUnit, err := storageunit.NewStorageUnitFromConf(
		epochByHashCacherConfig,
		epochByHashDbConfig,
		epochByHashPersisterCreator,
	)
	if err != nil {
		return fmt.Errorf("%w for DbLookupExtensions.EpochByHashStorageConfig", err)
	}

	chainStorer.AddStorer(dataRetriever.EpochByHashUnit, epochByHashUnit)

	return psf.setUpEsdtSuppliesStorer(chainStorer, shardID)
}

func (psf *StorageServiceFactory) setUpEsdtSuppliesStorer(chainStorer *dataRetriever.ChainStorer, shardIDStr string) error {
	esdtSuppliesUnit, err := psf.createEsdtSuppliesUnit(shardIDStr)
	if err != nil {
		return fmt.Errorf("%w for DbLookupExtensions.ESDTSuppliesStorageConfig", err)
	}

	if psf.repopulateTokensSupplies {
		// if the flag is set, then we need to clear the storer at this point. The easiest way is to destroy it and then create it again
		err = esdtSuppliesUnit.DestroyUnit()
		if err != nil {
			return err
		}

		time.Sleep(time.Second) // making sure the unit was properly closed and destroyed
		esdtSuppliesUnit, err = psf.createEsdtSuppliesUnit(shardIDStr)
		if err != nil {
			return err
		}
	}

	chainStorer.AddStorer(dataRetriever.ESDTSuppliesUnit, esdtSuppliesUnit)
	return nil
}

func (psf *StorageServiceFactory) createEsdtSuppliesUnit(shardIDStr string) (storage.Storer, error) {
	esdtSuppliesConfig := psf.generalConfig.DbLookupExtensions.ESDTSuppliesStorageConfig
	esdtSuppliesDbConfig := GetDBFromConfig(esdtSuppliesConfig.DB)
	esdtSuppliesDbConfig.FilePath = psf.pathManager.PathForStatic(shardIDStr, esdtSuppliesConfig.DB.FilePath)
	esdtSuppliesCacherConfig := GetCacherFromConfig(esdtSuppliesConfig.Cache)

	dbConfigHandlerInstance := NewDBConfigHandler(esdtSuppliesConfig.DB)
	esdtSuppliesPersisterCreator, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return nil, err
	}

	return storageunit.NewStorageUnitFromConf(
		esdtSuppliesCacherConfig, esdtSuppliesDbConfig,
		esdtSuppliesPersisterCreator)
}

func (psf *StorageServiceFactory) createPruningStorerArgs(
	storageConfig config.StorageConfig,
	customDatabaseRemover storage.CustomDatabaseRemoverHandler,
) (pruning.StorerArgs, error) {
	numOfEpochsToKeep := uint32(psf.generalConfig.StoragePruning.NumEpochsToKeep)
	numOfActivePersisters := uint32(psf.generalConfig.StoragePruning.NumActivePersisters)
	pruningEnabled := psf.generalConfig.StoragePruning.Enabled
	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := filepath.Join(psf.pathManager.PathForEpoch(shardId, psf.currentEpoch, storageConfig.DB.FilePath))
	epochsData := pruning.EpochArgs{
		StartingEpoch:         psf.currentEpoch,
		NumOfEpochsToKeep:     numOfEpochsToKeep,
		NumOfActivePersisters: numOfActivePersisters,
	}

	dbConfigHandlerInstance := NewDBConfigHandler(storageConfig.DB)
	persisterFactory, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return pruning.StorerArgs{}, err
	}

	args := pruning.StorerArgs{
		Identifier:                storageConfig.DB.FilePath,
		PruningEnabled:            pruningEnabled,
		OldDataCleanerProvider:    psf.oldDataCleanerProvider,
		CustomDatabaseRemover:     customDatabaseRemover,
		ShardCoordinator:          psf.shardCoordinator,
		CacheConf:                 GetCacherFromConfig(storageConfig.Cache),
		PathManager:               psf.pathManager,
		DbPath:                    dbPath,
		PersisterFactory:          persisterFactory,
		Notifier:                  psf.epochStartNotifier,
		MaxBatchSize:              storageConfig.DB.MaxBatchSize,
		EnabledDbLookupExtensions: psf.generalConfig.DbLookupExtensions.Enabled,
		PersistersTracker:         pruning.NewPersistersTracker(epochsData),
		EpochsData:                epochsData,
		StateStatsHandler:         psf.stateStatsHandler,
	}

	return args, nil
}

func (psf *StorageServiceFactory) createTrieEpochRootHashStorerIfNeeded() (storage.Storer, error) {
	if !psf.createTrieEpochRootHashStorer {
		return storageunit.NewNilStorer(), nil
	}

	trieEpochRootHashDbConfig := GetDBFromConfig(psf.generalConfig.TrieEpochRootHashStorage.DB)
	shardId := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := psf.pathManager.PathForStatic(shardId, psf.generalConfig.TrieEpochRootHashStorage.DB.FilePath)
	trieEpochRootHashDbConfig.FilePath = dbPath

	dbConfigHandlerInstance := NewDBConfigHandler(psf.generalConfig.TrieEpochRootHashStorage.DB)
	esdtSuppliesPersisterCreator, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return nil, err
	}

	trieEpochRootHashStorageUnit, err := storageunit.NewStorageUnitFromConf(
		GetCacherFromConfig(psf.generalConfig.TrieEpochRootHashStorage.Cache),
		trieEpochRootHashDbConfig,
		esdtSuppliesPersisterCreator,
	)
	if err != nil {
		return nil, fmt.Errorf("%w for TrieEpochRootHashStorage", err)
	}

	return trieEpochRootHashStorageUnit, nil
}

func (psf *StorageServiceFactory) createTriePersister(
	storageConfig config.StorageConfig,
) (storage.Storer, error) {
	trieDBConfig := GetDBFromConfig(storageConfig.DB)
	shardID := core.GetShardIDString(psf.shardCoordinator.SelfId())
	dbPath := psf.pathManager.PathForStatic(shardID, storageConfig.DB.FilePath)
	trieDBConfig.FilePath = dbPath

	dbConfigHandlerInstance := NewDBConfigHandler(storageConfig.DB)
	persisterFactory, err := NewPersisterFactory(dbConfigHandlerInstance)
	if err != nil {
		return nil, err
	}

	return storageunit.NewStorageUnitFromConf(
		GetCacherFromConfig(storageConfig.Cache),
		trieDBConfig,
		persisterFactory)
}

func (psf *StorageServiceFactory) createTriePruningPersister(arg pruning.StorerArgs) (storage.Storer, error) {
	isFullArchive := psf.prefsConfig.FullArchive
	isDBLookupExtension := psf.generalConfig.DbLookupExtensions.Enabled
	if !isFullArchive && !isDBLookupExtension {
		return pruning.NewTriePruningStorer(arg)
	}

	numOldActivePersisters := psf.getNumActivePersistersForFullHistoryStorer(isFullArchive, isDBLookupExtension)
	historyArgs := pruning.FullHistoryStorerArgs{
		StorerArgs:               arg,
		NumOfOldActivePersisters: numOldActivePersisters,
	}

	return pruning.NewFullHistoryTriePruningStorer(historyArgs)
}

func (psf *StorageServiceFactory) createPruningPersister(arg pruning.StorerArgs) (storage.Storer, error) {
	isFullArchive := psf.prefsConfig.FullArchive
	isDBLookupExtension := psf.generalConfig.DbLookupExtensions.Enabled
	if !isFullArchive && !isDBLookupExtension {
		return pruning.NewPruningStorer(arg)
	}

	numOldActivePersisters := psf.getNumActivePersistersForFullHistoryStorer(isFullArchive, isDBLookupExtension)
	historyArgs := pruning.FullHistoryStorerArgs{
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
