package factory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// ArgsNewOpenStorageUnits defines the arguments in order to open a set of storage units from disk
type ArgsNewOpenStorageUnits struct {
	GeneralConfig      config.Config
	Marshalizer        marshal.Marshalizer
	WorkingDir         string
	ChainID            string
	DefaultDBPath      string
	DefaultEpochString string
	DefaultShardString string
}

type openStorageUnits struct {
	generalConfig      config.Config
	marshalizer        marshal.Marshalizer
	workingDir         string
	chainID            string
	defaultDBPath      string
	defaultEpochString string
	defaultShardString string
}

// NewStorageUnitOpenHandler creates an openStorageUnits component
// TODO refactor this and unit tests
func NewStorageUnitOpenHandler(args ArgsNewOpenStorageUnits) (*openStorageUnits, error) {
	o := &openStorageUnits{
		generalConfig:      args.GeneralConfig,
		marshalizer:        args.Marshalizer,
		workingDir:         args.WorkingDir,
		chainID:            args.ChainID,
		defaultDBPath:      args.DefaultDBPath,
		defaultEpochString: args.DefaultEpochString,
		defaultShardString: args.DefaultShardString,
	}

	return o, nil
}

// OpenStorageUnits will open bootstrap storage unit
func (o *openStorageUnits) OpenStorageUnits() (storage.Storer, error) {
	parentDir, lastEpoch, err := getParentDirAndLastEpoch(
		o.workingDir,
		o.chainID,
		o.defaultDBPath,
		o.defaultEpochString)
	if err != nil {
		return nil, err
	}

	// TODO: refactor this - as it works with bootstrap storage unit only
	persisterFactory := NewPersisterFactory(o.generalConfig.BootstrapStorage.DB)
	pathWithoutShard := filepath.Join(
		parentDir,
		fmt.Sprintf("%s_%d", o.defaultEpochString, lastEpoch),
	)
	shardIdsStr, err := getShardsFromDirectory(pathWithoutShard, o.defaultShardString)
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

	cacher, errCache := lrucache.NewCache(10)
	if errCache != nil {
		return nil, errCache
	}

	storer, errStorageUnit := storageUnit.NewStorageUnit(cacher, persister)
	if errStorageUnit != nil {
		return nil, errStorageUnit
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
		time.Sleep(core.SleepTimeBetweenCreateDBRetries * time.Second)
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

		bootstrapData, storer, errGet := getBootstrapDataForPersisterPath(persisterFactory, persisterPath, o.marshalizer)
		if errGet != nil {
			continue
		}

		errClose := storer.Close()
		log.LogIfError(errClose)

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

// TODO refactor this and test it

// LatestDataFromStorage represents the DTO structure to return from storage
type LatestDataFromStorage struct {
	Epoch           uint32
	ShardID         uint32
	LastRound       int64
	EpochStartRound uint64
}

// FindLatestDataFromStorage finds the last data (such as last epoch, shard ID or round) by searching over the
// storage folders and opening older databases
func FindLatestDataFromStorage(
	generalConfig config.Config,
	marshalizer marshal.Marshalizer,
	workingDir string,
	chainID string,
	defaultDBPath string,
	defaultEpochString string,
	defaultShardString string,
) (LatestDataFromStorage, error) {
	parentDir, lastEpoch, err := getParentDirAndLastEpoch(workingDir, chainID, defaultDBPath, defaultEpochString)
	if err != nil {
		return LatestDataFromStorage{}, err
	}

	return getLastEpochAndRoundFromStorage(generalConfig, marshalizer, parentDir, defaultEpochString, defaultShardString, lastEpoch)
}

func getParentDirAndLastEpoch(
	workingDir string,
	chainID string,
	defaultDBPath string,
	defaultEpochString string,
) (string, uint32, error) {
	parentDir := filepath.Join(
		workingDir,
		defaultDBPath,
		chainID)

	f, err := os.Open(parentDir)
	if err != nil {
		return "", 0, err
	}

	files, err := f.Readdir(allFiles)
	_ = f.Close()

	if err != nil {
		return "", 0, err
	}

	epochDirs := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		isEpochDir := strings.HasPrefix(file.Name(), defaultEpochString)
		if !isEpochDir {
			continue
		}

		epochDirs = append(epochDirs, file.Name())
	}

	lastEpoch, err := getLastEpochFromDirNames(epochDirs)
	if err != nil {
		return "", 0, err
	}

	return parentDir, lastEpoch, nil
}

func getLastEpochFromDirNames(epochDirs []string) (uint32, error) {
	if len(epochDirs) == 0 {
		return 0, nil
	}

	re := regexp.MustCompile("[0-9]+")
	epochsInDirName := make([]uint32, 0, len(epochDirs))

	for _, dirname := range epochDirs {
		epochStr := re.FindString(dirname)
		epoch, err := strconv.ParseInt(epochStr, 10, 64)
		if err != nil {
			return 0, err
		}

		epochsInDirName = append(epochsInDirName, uint32(epoch))
	}

	sort.Slice(epochsInDirName, func(i, j int) bool {
		return epochsInDirName[i] > epochsInDirName[j]
	})

	return epochsInDirName[0], nil
}

func getLastEpochAndRoundFromStorage(
	config config.Config,
	marshalizer marshal.Marshalizer,
	parentDir string,
	defaultEpochString string,
	defaultShardString string,
	epoch uint32,
) (LatestDataFromStorage, error) {
	persisterFactory := NewPersisterFactory(config.BootstrapStorage.DB)
	pathWithoutShard := filepath.Join(
		parentDir,
		fmt.Sprintf("%s_%d", defaultEpochString, epoch),
	)
	shardIdsStr, err := getShardsFromDirectory(pathWithoutShard, defaultShardString)
	if err != nil {
		return LatestDataFromStorage{}, err
	}

	var mostRecentBootstrapData *bootstrapStorage.BootstrapData
	var mostRecentShard string
	highestRoundInStoredShards := int64(0)
	epochStartRound := uint64(0)

	var storer storage.Storer
	var errGet error
	var bootstrapData *bootstrapStorage.BootstrapData
	for _, shardIdStr := range shardIdsStr {
		if !check.IfNil(storer) {
			errClose := storer.Close()
			log.LogIfError(errClose)
		}

		persisterPath := filepath.Join(
			pathWithoutShard,
			fmt.Sprintf("%s_%s", defaultShardString, shardIdStr),
			config.BootstrapStorage.DB.FilePath,
		)

		bootstrapData, storer, errGet = getBootstrapDataForPersisterPath(persisterFactory, persisterPath, marshalizer)
		if errGet != nil {
			continue
		}

		if bootstrapData.LastRound > highestRoundInStoredShards {
			shardID, err := convertShardIDToUint32(shardIdStr)
			if err != nil {
				log.Debug("could not convert shard id to uint32", "err", err)
				continue
			}
			epochStartRound, err = loadEpochStartRound(shardID, bootstrapData.EpochStartTriggerConfigKey, storer)
			if err != nil {
				log.Debug("could not load epoch start round", "err", err)
				continue
			}

			highestRoundInStoredShards = bootstrapData.LastRound
			mostRecentBootstrapData = bootstrapData
			mostRecentShard = shardIdStr
		}
	}

	if !check.IfNil(storer) {
		errClose := storer.Close()
		log.LogIfError(errClose)
	}

	if mostRecentBootstrapData == nil {
		return LatestDataFromStorage{}, storage.ErrBootstrapDataNotFoundInStorage
	}
	shardIDAsUint32, err := convertShardIDToUint32(mostRecentShard)
	if err != nil {
		return LatestDataFromStorage{}, err
	}

	lastestData := LatestDataFromStorage{
		Epoch:           mostRecentBootstrapData.LastHeader.Epoch,
		ShardID:         shardIDAsUint32,
		LastRound:       mostRecentBootstrapData.LastRound,
		EpochStartRound: epochStartRound,
	}

	return lastestData, nil
}

func loadEpochStartRound(
	shardID uint32,
	key []byte,
	storer storage.Storer,
) (uint64, error) {
	trigInternalKey := append([]byte(core.TriggerRegistryKeyPrefix), key...)
	data, err := storer.Get(trigInternalKey)
	if err != nil {
		return 0, err
	}

	if shardID == core.MetachainShardId {
		state := &metachain.TriggerRegistry{}
		err = json.Unmarshal(data, state)
		if err != nil {
			return 0, err
		}

		return state.CurrEpochStartRound, nil
	}

	state := &shardchain.TriggerRegistry{}
	err = json.Unmarshal(data, state)
	if err != nil {
		return 0, err
	}

	return state.EpochStartRound, nil
}

func getBootstrapDataForPersisterPath(
	persisterFactory *PersisterFactory,
	persisterPath string,
	marshalizer marshal.Marshalizer,
) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
	persister, err := persisterFactory.Create(persisterPath)
	if err != nil {
		return nil, nil, err
	}

	cacher, err := lrucache.NewCache(10)
	if err != nil {
		errClose := persister.Close()
		log.LogIfError(errClose)
		return nil, nil, err
	}

	var storer storage.Storer
	defer func() {
		if err != nil {
			errClose := storer.Close()
			log.LogIfError(errClose)
		}
	}()

	storer, err = storageUnit.NewStorageUnit(cacher, persister)
	if err != nil {
		return nil, nil, err
	}

	bootStorer, err := bootstrapStorage.NewBootstrapStorer(marshalizer, storer)
	if err != nil {
		return nil, nil, err
	}

	highestRound := bootStorer.GetHighestRound()
	bootstrapData, err := bootStorer.Get(highestRound)
	if err != nil {
		return nil, nil, err
	}

	return &bootstrapData, storer, nil
}

func getShardsFromDirectory(path string, defaultShardString string) ([]string, error) {
	shardIDs := make([]string, 0)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdir(allFiles)
	_ = f.Close()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := file.Name()
		stringToSplitBy := defaultShardString + "_"
		splitSlice := strings.Split(fileName, stringToSplitBy)
		if len(splitSlice) < 2 {
			continue
		}

		shardIDs = append(shardIDs, splitSlice[1])
	}

	return shardIDs, nil
}
