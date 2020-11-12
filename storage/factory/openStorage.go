package factory

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// ArgsNewOpenStorageUnits defines the arguments in order to open a set of storage units from disk
type ArgsNewOpenStorageUnits struct {
	GeneralConfig             config.Config
	BootstrapDataProvider     BootstrapDataProviderHandler
	LatestStorageDataProvider storage.LatestStorageDataProviderHandler
	DefaultEpochString        string
	DefaultShardString        string
}

type openStorageUnits struct {
	generalConfig             config.Config
	bootstrapDataProvider     BootstrapDataProviderHandler
	latestStorageDataProvider storage.LatestStorageDataProviderHandler
	defaultEpochString        string
	defaultShardString        string
}

// NewStorageUnitOpenHandler creates an openStorageUnits component
func NewStorageUnitOpenHandler(args ArgsNewOpenStorageUnits) (*openStorageUnits, error) {
	o := &openStorageUnits{
		generalConfig:             args.GeneralConfig,
		defaultEpochString:        args.DefaultEpochString,
		defaultShardString:        args.DefaultShardString,
		bootstrapDataProvider:     args.BootstrapDataProvider,
		latestStorageDataProvider: args.LatestStorageDataProvider,
	}

	return o, nil
}

// GetMostRecentBootstrapStorageUnit will open bootstrap storage unit
func (o *openStorageUnits) GetMostRecentBootstrapStorageUnit() (storage.Storer, error) {
	parentDir, lastEpoch, err := o.latestStorageDataProvider.GetParentDirAndLastEpoch()
	if err != nil {
		return nil, err
	}

	// TODO: refactor this - as it works with bootstrap storage unit only
	persisterFactory := NewPersisterFactory(o.generalConfig.BootstrapStorage.DB)
	pathWithoutShard := filepath.Join(
		parentDir,
		fmt.Sprintf("%s_%d", o.defaultEpochString, lastEpoch),
	)
	shardIdsStr, err := o.latestStorageDataProvider.GetShardsFromDirectory(pathWithoutShard)
	if err != nil {
		return nil, err
	}

	mostRecentShard, err := o.getMostUpToDateDirectory(pathWithoutShard, shardIdsStr, persisterFactory)
	if err != nil {
		return nil, err
	}

	persisterPath := filepath.Join(
		pathWithoutShard,
		fmt.Sprintf("%s_%s", o.defaultShardString, mostRecentShard),
		o.generalConfig.BootstrapStorage.DB.FilePath,
	)

	persister, err := createDB(persisterFactory, persisterPath)
	if err != nil {
		return nil, err
	}

	cacher, err := lrucache.NewCache(10)
	if err != nil {
		return nil, err
	}

	storer, err := storageUnit.NewStorageUnit(cacher, persister)
	if err != nil {
		return nil, err
	}

	return storer, nil
}

func createDB(persisterFactory *PersisterFactory, persisterPath string) (storage.Persister, error) {
	var persister storage.Persister
	var err error
	for i := 0; i < core.MaxRetriesToCreateDB; i++ {
		persister, err = persisterFactory.Create(persisterPath)
		if err == nil {
			return persister, nil
		}
		log.Warn("Create Persister failed", "path", persisterPath, "error", err)
		time.Sleep(core.SleepTimeBetweenCreateDBRetries)
	}
	return nil, err
}

func (o *openStorageUnits) getMostUpToDateDirectory(
	pathWithoutShard string,
	shardIdsStr []string,
	persisterFactory *PersisterFactory,
) (string, error) {
	var mostRecentShard string
	highestRoundInStoredShards := int64(0)

	for _, shardIdStr := range shardIdsStr {
		persisterPath := filepath.Join(
			pathWithoutShard,
			fmt.Sprintf("%s_%s", o.defaultShardString, shardIdStr),
			o.generalConfig.BootstrapStorage.DB.FilePath,
		)

		bootstrapData, errGet := o.loadDataForShard(persisterFactory, persisterPath)
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

func (o *openStorageUnits) loadDataForShard(
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
