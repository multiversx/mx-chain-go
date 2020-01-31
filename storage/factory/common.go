package factory

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// GetCacherFromConfig will return the cache config needed for storage unit from a config came from the toml file
func GetCacherFromConfig(cfg config.CacheConfig) storageUnit.CacheConfig {
	return storageUnit.CacheConfig{
		Size:   cfg.Size,
		Type:   storageUnit.CacheType(cfg.Type),
		Shards: cfg.Shards,
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

// FindLastEpochFromStorage finds the last epoch by searching over the storage folders
// TODO: something similar should be done for determining the correct shardId
// when booting from storage with an epoch > 0 or add ShardId in boot storer
func FindLastEpochFromStorage(
	workingDir string,
	chainID string,
	defaultDBPath string,
	defaultEpochString string,
) (uint32, error) {
	parentDir := filepath.Join(
		workingDir,
		defaultDBPath,
		chainID)

	f, err := os.Open(parentDir)
	if err != nil {
		return 0, err
	}

	files, err := f.Readdir(-1)
	_ = f.Close()

	if err != nil {
		return 0, err
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

	if len(epochDirs) == 0 {
		return 0, nil
	}

	sort.Slice(epochDirs, func(i, j int) bool {
		return epochDirs[i] > epochDirs[j]
	})

	re := regexp.MustCompile("[0-9]+")
	epochStr := re.FindString(epochDirs[0])
	epoch, err := strconv.ParseInt(epochStr, 10, 64)

	return uint32(epoch), err
}
