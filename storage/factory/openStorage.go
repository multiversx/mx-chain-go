package factory

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// ArgsNewOpenStorageUnits defines the arguments in order to open a set of storage units from disk
type ArgsNewOpenStorageUnits struct {
	GeneralConfig             config.Config
	Marshalizer               marshal.Marshalizer
	BootstrapDataProvider     BootstrapDataProviderHandler
	LatestStorageDataProvider storage.LatestStorageDataProviderHandler
	WorkingDir                string
	ChainID                   string
	DefaultDBPath             string
	DefaultEpochString        string
	DefaultShardString        string
}

type openStorageUnits struct {
	generalConfig             config.Config
	marshalizer               marshal.Marshalizer
	bootstrapDataProvider     BootstrapDataProviderHandler
	latestStorageDataProvider storage.LatestStorageDataProviderHandler
	workingDir                string
	chainID                   string
	defaultDBPath             string
	defaultEpochString        string
	defaultShardString        string
}

// NewStorageUnitOpenHandler creates an openStorageUnits component
func NewStorageUnitOpenHandler(args ArgsNewOpenStorageUnits) (*openStorageUnits, error) {
	o := &openStorageUnits{
		generalConfig:             args.GeneralConfig,
		marshalizer:               args.Marshalizer,
		workingDir:                args.WorkingDir,
		chainID:                   args.ChainID,
		defaultDBPath:             args.DefaultDBPath,
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

	defer func() {
		if err != nil {
			errClose := persister.Close()
			log.LogIfError(errClose)
		}
	}()

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
		log.Warn("Create Persister failed", "path", persisterPath)
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

		bootstrapData, _, errGet := o.bootstrapDataProvider.LoadForPath(persisterFactory, persisterPath)
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

// IsInterfaceNil returns true if there is no value under the interface
func (o *openStorageUnits) IsInterfaceNil() bool {
	return o == nil
}
