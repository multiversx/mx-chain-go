package factory

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.LatestStorageDataProviderHandler = (*latestDataProvider)(nil)

// ArgsLatestDataProvider holds the arguments needed for creating a latestDataProvider object
type ArgsLatestDataProvider struct {
	GeneralConfig         config.Config
	BootstrapDataProvider BootstrapDataProviderHandler
	DirectoryReader       storage.DirectoryReaderHandler
	ParentDir             string
	DefaultEpochString    string
	DefaultShardString    string
}

type iteratedShardData struct {
	bootstrapData   *bootstrapStorage.BootstrapData
	epochStartRound uint64
	shardIDStr      string
	successful      bool
}

type latestDataProvider struct {
	generalConfig         config.Config
	bootstrapDataProvider BootstrapDataProviderHandler
	directoryReader       storage.DirectoryReaderHandler
	parentDir             string
	defaultEpochString    string
	defaultShardString    string
}

// NewLatestDataProvider returns a new instance of latestDataProvider
func NewLatestDataProvider(args ArgsLatestDataProvider) (*latestDataProvider, error) {
	return &latestDataProvider{
		generalConfig:         args.GeneralConfig,
		parentDir:             args.ParentDir,
		directoryReader:       args.DirectoryReader,
		defaultShardString:    args.DefaultShardString,
		defaultEpochString:    args.DefaultEpochString,
		bootstrapDataProvider: args.BootstrapDataProvider,
	}, nil
}

// Get will return a struct containing the latest data in storage
func (ldp *latestDataProvider) Get() (storage.LatestDataFromStorage, error) {
	parentDir, lastEpoch, err := ldp.GetParentDirAndLastEpoch()
	if err != nil {
		return storage.LatestDataFromStorage{}, err
	}

	return ldp.getLastEpochAndRoundFromStorage(parentDir, lastEpoch)
}

// GetParentDirAndLastEpoch returns the parent directory and last epoch
func (ldp *latestDataProvider) GetParentDirAndLastEpoch() (string, uint32, error) {
	directoriesNames, err := ldp.directoryReader.ListDirectoriesAsString(ldp.parentDir)
	if err != nil {
		return "", 0, err
	}

	epochDirs := make([]string, 0, len(directoriesNames))
	for _, dirName := range directoriesNames {
		isEpochDir := strings.HasPrefix(dirName, ldp.defaultEpochString)
		if !isEpochDir {
			continue
		}

		epochDirs = append(epochDirs, dirName)
	}

	lastEpoch, err := ldp.GetLastEpochFromDirNames(epochDirs)
	if err != nil {
		return "", 0, err
	}

	return ldp.parentDir, lastEpoch, nil
}

func (ldp *latestDataProvider) getLastEpochAndRoundFromStorage(parentDir string, lastEpoch uint32) (storage.LatestDataFromStorage, error) {
	persisterFactory := NewPersisterFactory(ldp.generalConfig.BootstrapStorage.DB)
	pathWithoutShard := filepath.Join(
		parentDir,
		fmt.Sprintf("%s_%d", ldp.defaultEpochString, lastEpoch),
	)
	shardIdsStr, err := ldp.GetShardsFromDirectory(pathWithoutShard)
	if err != nil {
		return storage.LatestDataFromStorage{}, err
	}

	var mostRecentBootstrapData *bootstrapStorage.BootstrapData
	var mostRecentShard string
	highestRoundInStoredShards := int64(0)
	epochStartRound := uint64(0)

	for _, shardIdStr := range shardIdsStr {
		persisterPath := filepath.Join(
			pathWithoutShard,
			fmt.Sprintf("%s_%s", ldp.defaultShardString, shardIdStr),
			ldp.generalConfig.BootstrapStorage.DB.FilePath,
		)

		shardData := ldp.loadDataForShard(highestRoundInStoredShards, shardIdStr, persisterFactory, persisterPath)
		if shardData.successful {
			epochStartRound = shardData.epochStartRound
			highestRoundInStoredShards = shardData.bootstrapData.LastRound
			mostRecentBootstrapData = shardData.bootstrapData
			mostRecentShard = shardIdStr
		}
	}

	if mostRecentBootstrapData == nil {
		return storage.LatestDataFromStorage{}, storage.ErrBootstrapDataNotFoundInStorage
	}
	shardIDAsUint32, err := convertShardIDToUint32(mostRecentShard)
	if err != nil {
		return storage.LatestDataFromStorage{}, err
	}

	lastestData := storage.LatestDataFromStorage{
		Epoch:           mostRecentBootstrapData.LastHeader.Epoch,
		ShardID:         shardIDAsUint32,
		LastRound:       mostRecentBootstrapData.LastRound,
		EpochStartRound: epochStartRound,
	}

	return lastestData, nil
}

func (ldp *latestDataProvider) loadDataForShard(currentHighestRound int64, shardIdStr string, persisterFactory storage.PersisterFactory, persisterPath string) *iteratedShardData {
	bootstrapData, storer, errGet := ldp.bootstrapDataProvider.LoadForPath(persisterFactory, persisterPath)
	defer func() {
		if storer != nil {
			err := storer.Close()
			if err != nil {
				log.Debug("latestDataProvider: closing storer", "path", persisterPath, "error", err)
			}
		}
	}()
	if errGet != nil {
		return &iteratedShardData{}
	}

	if bootstrapData.LastRound > currentHighestRound {
		shardID := uint32(0)
		var err error
		shardID, err = convertShardIDToUint32(shardIdStr)
		if err != nil {
			return &iteratedShardData{}
		}
		epochStartRound, err := ldp.loadEpochStartRound(shardID, bootstrapData.EpochStartTriggerConfigKey, storer)
		if err != nil {
			return &iteratedShardData{}
		}

		return &iteratedShardData{
			bootstrapData:   bootstrapData,
			shardIDStr:      shardIdStr,
			epochStartRound: epochStartRound,
			successful:      true,
		}
	}

	return &iteratedShardData{}
}

// loadEpochStartRound will return the epoch start round from the bootstrap unit
func (ldp *latestDataProvider) loadEpochStartRound(
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

// GetLastEpochFromDirNames returns the last epoch found in storage directory
func (ldp *latestDataProvider) GetLastEpochFromDirNames(epochDirs []string) (uint32, error) {
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

// GetShardsFromDirectory will return names of shards as string from a provided directory
func (ldp *latestDataProvider) GetShardsFromDirectory(path string) ([]string, error) {
	shardIDs := make([]string, 0)
	directoriesNames, err := ldp.directoryReader.ListDirectoriesAsString(path)
	if err != nil {
		return nil, err
	}

	shardDirs := make([]string, 0, len(directoriesNames))
	for _, dirName := range directoriesNames {
		isShardDir := strings.HasPrefix(dirName, ldp.defaultShardString)
		if !isShardDir {
			continue
		}

		shardDirs = append(shardDirs, dirName)
	}

	for _, fileName := range shardDirs {
		stringToSplitBy := ldp.defaultShardString + "_"
		splitSlice := strings.Split(fileName, stringToSplitBy)
		if len(splitSlice) < 2 {
			continue
		}

		shardIDs = append(shardIDs, splitSlice[1])
	}

	return shardIDs, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ldp *latestDataProvider) IsInterfaceNil() bool {
	return ldp == nil
}
