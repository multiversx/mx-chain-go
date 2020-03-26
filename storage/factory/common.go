package factory

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

const allFiles = -1

// GetCacherFromConfig will return the cache config needed for storage unit from a config came from the toml file
func GetCacherFromConfig(cfg config.CacheConfig) storageUnit.CacheConfig {
	return storageUnit.CacheConfig{
		Size:        cfg.Size,
		SizeInBytes: cfg.SizeInBytes,
		Type:        storageUnit.CacheType(cfg.Type),
		Shards:      cfg.Shards,
	}
}

// GetDBFromConfig will return the db config needed for storage unit from a config came from the toml file
func GetDBFromConfig(cfg config.DBConfig) storageUnit.DBConfig {
	return storageUnit.DBConfig{
		Type:              storageUnit.DBType(cfg.Type),
		MaxBatchSize:      cfg.MaxBatchSize,
		BatchDelaySeconds: cfg.BatchDelaySeconds,
		MaxOpenFiles:      cfg.MaxOpenFiles,
	}
}

// GetBloomFromConfig will return the bloom config needed for storage unit from a config came from the toml file
func GetBloomFromConfig(cfg config.BloomFilterConfig) storageUnit.BloomConfig {
	var hashFuncs []storageUnit.HasherType
	if cfg.HashFunc != nil {
		hashFuncs = make([]storageUnit.HasherType, len(cfg.HashFunc))
		idx := 0
		for _, hf := range cfg.HashFunc {
			hashFuncs[idx] = storageUnit.HasherType(hf)
			idx++
		}
	}

	return storageUnit.BloomConfig{
		Size:     cfg.Size,
		HashFunc: hashFuncs,
	}
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
) (uint32, uint32, int64, error) {
	parentDir := filepath.Join(
		workingDir,
		defaultDBPath,
		chainID)

	f, err := os.Open(parentDir)
	if err != nil {
		return 0, 0, 0, err
	}

	files, err := f.Readdir(allFiles)
	_ = f.Close()

	if err != nil {
		return 0, 0, 0, err
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
		return 0, 0, 0, err
	}

	return getLastEpochAndRoundFromStorage(generalConfig, marshalizer, parentDir, defaultEpochString, defaultShardString, lastEpoch)
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
) (uint32, uint32, int64, error) {
	persisterFactory := NewPersisterFactory(config.BootstrapStorage.DB)
	pathWithoutShard := filepath.Join(
		parentDir,
		fmt.Sprintf("%s_%d", defaultEpochString, epoch),
	)
	shardIdsStr, err := getShardsFromDirectory(pathWithoutShard, defaultShardString)
	if err != nil {
		return 0, 0, 0, err
	}

	var mostRecentBootstrapData *bootstrapStorage.BootstrapData
	var mostRecentShard string
	highestRoundInStoredShards := int64(0)

	for _, shardIdStr := range shardIdsStr {
		persisterPath := filepath.Join(
			pathWithoutShard,
			fmt.Sprintf("%s_%s", defaultShardString, shardIdStr),
			config.BootstrapStorage.DB.FilePath,
		)

		bootstrapData, errGet := getBootstrapDataForPersisterPath(persisterFactory, persisterPath, marshalizer)
		if errGet != nil {
			continue
		}

		if bootstrapData.LastRound > highestRoundInStoredShards {
			highestRoundInStoredShards = bootstrapData.LastRound
			mostRecentBootstrapData = bootstrapData
			mostRecentShard = shardIdStr
		}
	}

	if mostRecentBootstrapData == nil {
		return 0, 0, 0, storage.ErrBootstrapDataNotFoundInStorage
	}
	shardIDAsUint32, err := convertShardIDToUint32(mostRecentShard)
	if err != nil {
		return 0, 0, 0, err
	}

	return mostRecentBootstrapData.LastHeader.Epoch, shardIDAsUint32, mostRecentBootstrapData.LastRound, nil
}

func getBootstrapDataForPersisterPath(
	persisterFactory *PersisterFactory,
	persisterPath string,
	marshalizer marshal.Marshalizer,
) (*bootstrapStorage.BootstrapData, error) {
	persister, err := persisterFactory.Create(persisterPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		errClose := persister.Close()
		log.LogIfError(errClose)
	}()

	cacher, err := lrucache.NewCache(10)
	if err != nil {
		return nil, err
	}

	storer, err := storageUnit.NewStorageUnit(cacher, persister)
	if err != nil {
		return nil, err
	}

	bootStorer, err := bootstrapStorage.NewBootstrapStorer(marshalizer, storer)
	if err != nil {
		return nil, err
	}

	highestRound := bootStorer.GetHighestRound()
	bootstrapData, err := bootStorer.Get(highestRound)
	if err != nil {
		return nil, err
	}

	return &bootstrapData, nil
}

func getShardsFromDirectory(path string, defaultShardString string) ([]string, error) {
	shardIDs := make([]string, 0)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	files, err := f.Readdir(allFiles)
	_ = f.Close()

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

func convertShardIDToUint32(shardIDStr string) (uint32, error) {
	if shardIDStr == "metachain" {
		return math.MaxUint32, nil
	}

	shardID, err := strconv.ParseInt(shardIDStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint32(shardID), nil
}
