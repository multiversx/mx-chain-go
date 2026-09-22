package factory

import (
	"fmt"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

const cacheSize = 10

// ArgsNewOpenStorageUnits defines the arguments in order to open a set of storage units from disk
type ArgsNewOpenStorageUnits struct {
	BootstrapDataProvider     BootstrapDataProviderHandler
	LatestStorageDataProvider storage.LatestStorageDataProviderHandler
	DefaultEpochString        string
	DefaultShardString        string
	RecoveryCheckpointEnabled bool
}

type openStorageUnits struct {
	bootstrapDataProvider     BootstrapDataProviderHandler
	latestStorageDataProvider storage.LatestStorageDataProviderHandler
	defaultEpochString        string
	defaultShardString        string
	recoveryCheckpointEnabled bool
}

// NewStorageUnitOpenHandler creates an openStorageUnits component
func NewStorageUnitOpenHandler(args ArgsNewOpenStorageUnits) (*openStorageUnits, error) {
	if check.IfNil(args.BootstrapDataProvider) {
		return nil, storage.ErrNilBootstrapDataProvider
	}
	if check.IfNil(args.LatestStorageDataProvider) {
		return nil, storage.ErrNilLatestStorageDataProvider
	}

	o := &openStorageUnits{
		defaultEpochString:        args.DefaultEpochString,
		defaultShardString:        args.DefaultShardString,
		bootstrapDataProvider:     args.BootstrapDataProvider,
		latestStorageDataProvider: args.LatestStorageDataProvider,
		recoveryCheckpointEnabled: args.RecoveryCheckpointEnabled,
	}

	return o, nil
}

// GetMostRecentStorageUnit will open bootstrap storage unit
func (o *openStorageUnits) GetMostRecentStorageUnit(dbConfig config.DBConfig) (storage.Storer, error) {
	persisterFactory, err := NewPersisterFactory(dbConfig)
	if err != nil {
		return nil, err
	}

	persisterPath, err := o.getMostRecentPersisterPath(dbConfig, persisterFactory)
	if err != nil {
		return nil, err
	}

	persister, err := persisterFactory.CreateWithRetries(persisterPath)
	if err != nil {
		return nil, err
	}

	cacher, err := cache.NewLRUCache(cacheSize)
	if err != nil {
		return nil, err
	}

	storer, err := storageunit.NewStorageUnit(cacher, persister)
	if err != nil {
		return nil, err
	}

	return storer, nil
}

func (o *openStorageUnits) getMostRecentPersisterPath(dbConfig config.DBConfig, persisterFactory storage.PersisterFactory) (string, error) {
	if o.recoveryCheckpointEnabled {
		selected, err := o.latestStorageDataProvider.Get()
		if err != nil {
			return "", err
		}
		parentDir := o.latestStorageDataProvider.GetParentDirectory()
		pathWithoutShard := o.getPathWithoutShard(parentDir, selected.Epoch)
		return o.getPersisterPath(pathWithoutShard, core.GetShardIDString(selected.ShardID), dbConfig), nil
	}

	parentDir, lastEpoch, err := o.latestStorageDataProvider.GetParentDirAndLastEpoch()
	if err != nil {
		return "", err
	}
	pathWithoutShard := o.getPathWithoutShard(parentDir, lastEpoch)
	shardIdsStr, err := o.latestStorageDataProvider.GetShardsFromDirectory(pathWithoutShard)
	if err != nil {
		return "", err
	}

	mostRecentShard, err := o.getMostUpToDateDirectory(dbConfig, pathWithoutShard, shardIdsStr, persisterFactory)
	if err != nil {
		return "", err
	}
	return o.getPersisterPath(pathWithoutShard, mostRecentShard, dbConfig), nil
}

func (o *openStorageUnits) getPathWithoutShard(parentDir string, epoch uint32) string {
	return filepath.Join(
		parentDir,
		fmt.Sprintf("%s_%d", o.defaultEpochString, epoch),
	)
}

func (o *openStorageUnits) getPersisterPath(pathWithoutShard string, shardID string, dbConfig config.DBConfig) string {
	return filepath.Join(
		pathWithoutShard,
		fmt.Sprintf("%s_%s", o.defaultShardString, shardID),
		dbConfig.FilePath,
	)
}

// OpenDB opens or creates a given DB
func (o *openStorageUnits) OpenDB(dbConfig config.DBConfig, shardID uint32, epoch uint32) (storage.Storer, error) {
	parentDir := o.latestStorageDataProvider.GetParentDirectory()
	pathWithoutShard := o.getPathWithoutShard(parentDir, epoch)
	persisterPath := o.getPersisterPath(pathWithoutShard, fmt.Sprintf("%d", shardID), dbConfig)
	persisterFactory, err := NewPersisterFactory(dbConfig)
	if err != nil {
		return nil, err
	}

	persister, err := persisterFactory.CreateWithRetries(persisterPath)
	if err != nil {
		return nil, err
	}

	lruCache, err := cache.NewLRUCache(cacheSize)
	if err != nil {
		return nil, err
	}

	return storageunit.NewStorageUnit(lruCache, persister)
}

func (o *openStorageUnits) getMostUpToDateDirectory(
	dbConfig config.DBConfig,
	pathWithoutShard string,
	shardIdsStr []string,
	persisterFactory storage.PersisterFactory,
) (string, error) {
	var mostRecentShard string
	highestRoundInStoredShards := int64(0)

	for _, shardIdStr := range shardIdsStr {
		persisterPath := filepath.Join(
			pathWithoutShard,
			fmt.Sprintf("%s_%s", o.defaultShardString, shardIdStr),
			dbConfig.FilePath,
		)

		bootstrapData, errGet := o.loadBootstrapDataForShard(persisterFactory, persisterPath)
		if errGet != nil {
			continue
		}

		if bootstrapData.LastRound > highestRoundInStoredShards {
			highestRoundInStoredShards = bootstrapData.LastRound
			mostRecentShard = shardIdStr
		}
	}

	if len(mostRecentShard) == 0 {
		return "", storage.ErrBootstrapDataNotFoundInStorage
	}

	return mostRecentShard, nil
}

func (o *openStorageUnits) loadBootstrapDataForShard(
	persisterFactory storage.PersisterFactory,
	persisterPath string,
) (*bootstrapStorage.BootstrapData, error) {
	bootstrapData, storer, err := o.bootstrapDataProvider.LoadForPath(persisterFactory, persisterPath)
	defer func() {
		if storer != nil {
			errClose := storer.Close()
			if errClose != nil {
				log.Debug("openStorageunits: error closing storer",
					"persister path", persisterPath,
					"error", errClose)
			}
		}
	}()

	return bootstrapData, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (o *openStorageUnits) IsInterfaceNil() bool {
	return o == nil
}
