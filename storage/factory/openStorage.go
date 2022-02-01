package factory

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

const cacheSize = 10

// ArgsNewOpenStorageUnits defines the arguments in order to open a set of storage units from disk
type ArgsNewOpenStorageUnits struct {
	BootstrapDataProvider     BootstrapDataProviderHandler
	LatestStorageDataProvider storage.LatestStorageDataProviderHandler
	DefaultEpochString        string
	DefaultShardString        string
}

type openStorageUnits struct {
	bootstrapDataProvider     BootstrapDataProviderHandler
	latestStorageDataProvider storage.LatestStorageDataProviderHandler
	defaultEpochString        string
	defaultShardString        string
}

// NewStorageUnitOpenHandler creates an openStorageUnits component
func NewStorageUnitOpenHandler(args ArgsNewOpenStorageUnits) (*openStorageUnits, error) {
	o := &openStorageUnits{
		defaultEpochString:        args.DefaultEpochString,
		defaultShardString:        args.DefaultShardString,
		bootstrapDataProvider:     args.BootstrapDataProvider,
		latestStorageDataProvider: args.LatestStorageDataProvider,
	}

	return o, nil
}

// GetMostRecentStorageUnit will open bootstrap storage unit
func (o *openStorageUnits) GetMostRecentStorageUnit(dbConfig config.DBConfig) (storage.Storer, error) {
	parentDir, lastEpoch, err := o.latestStorageDataProvider.GetParentDirAndLastEpoch()
	if err != nil {
		return nil, err
	}

	persisterFactory := NewPersisterFactory(dbConfig)
	pathWithoutShard := o.getPathWithoutShard(parentDir, lastEpoch)
	shardIdsStr, err := o.latestStorageDataProvider.GetShardsFromDirectory(pathWithoutShard)
	if err != nil {
		return nil, err
	}

	mostRecentShard, err := o.getMostUpToDateDirectory(dbConfig, pathWithoutShard, shardIdsStr, persisterFactory)
	if err != nil {
		return nil, err
	}

	persisterPath := o.getPersisterPath(pathWithoutShard, mostRecentShard, dbConfig)

	persister, err := createDB(persisterFactory, persisterPath)
	if err != nil {
		return nil, err
	}

	cacher, err := lrucache.NewCache(cacheSize)
	if err != nil {
		return nil, err
	}

	storer, err := storageUnit.NewStorageUnit(cacher, persister)
	if err != nil {
		return nil, err
	}

	return storer, nil
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
	persisterFactory := NewPersisterFactory(dbConfig)

	persister, err := createDB(persisterFactory, persisterPath)
	if err != nil {
		return nil, err
	}

	cache, err := lrucache.NewCache(cacheSize)
	if err != nil {
		return nil, err
	}

	return storageUnit.NewStorageUnit(cache, persister)
}

func createDB(persisterFactory *PersisterFactory, persisterPath string) (storage.Persister, error) {
	var persister storage.Persister
	var err error
	for i := 0; i < common.MaxRetriesToCreateDB; i++ {
		persister, err = persisterFactory.Create(persisterPath)
		if err == nil {
			return persister, nil
		}
		log.Warn("Create Persister failed", "path", persisterPath, "error", err)
		//TODO: extract this in a parameter and inject it
		time.Sleep(common.SleepTimeBetweenCreateDBRetries)
	}
	return nil, err
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
