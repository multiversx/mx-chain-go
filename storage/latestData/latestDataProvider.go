package latestData

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart/metachain"
	"github.com/multiversx/mx-chain-go/epochStart/shardchain"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("storage/latestData")
var _ storage.LatestStorageDataProviderHandler = (*latestDataProvider)(nil)

// ArgsLatestDataProvider holds the arguments needed for creating a latestDataProvider object
type ArgsLatestDataProvider struct {
	GeneralConfig         config.Config
	BootstrapDataProvider factory.BootstrapDataProviderHandler
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
	bootstrapDataProvider factory.BootstrapDataProviderHandler
	directoryReader       storage.DirectoryReaderHandler
	parentDir             string
	defaultEpochString    string
	defaultShardString    string
}

// NewLatestDataProvider returns a new instance of latestDataProvider
func NewLatestDataProvider(args ArgsLatestDataProvider) (*latestDataProvider, error) {
	if check.IfNil(args.DirectoryReader) {
		return nil, storage.ErrNilDirectoryReader
	}
	if check.IfNil(args.BootstrapDataProvider) {
		return nil, storage.ErrNilBootstrapDataProvider
	}

	return &latestDataProvider{
		generalConfig:         args.GeneralConfig,
		parentDir:             args.ParentDir,
		directoryReader:       args.DirectoryReader,
		defaultShardString:    args.DefaultShardString,
		defaultEpochString:    args.DefaultEpochString,
		bootstrapDataProvider: args.BootstrapDataProvider,
	}, nil
}

// Get will return a struct containing the latest usable data in storage
func (ldp *latestDataProvider) Get() (storage.LatestDataFromStorage, error) {
	lastData, _, _, err := ldp.getLastData()
	return lastData, err
}

// GetParentDirectory returns the parent directory
func (ldp *latestDataProvider) GetParentDirectory() string {
	return ldp.parentDir
}

// GetParentDirAndLastEpoch returns the parent directory and last usable epoch for the node
func (ldp *latestDataProvider) GetParentDirAndLastEpoch() (string, uint32, error) {
	_, parentDir, lastEpoch, err := ldp.getLastData()
	return parentDir, lastEpoch, err
}

func (ldp *latestDataProvider) getLastData() (storage.LatestDataFromStorage, string, uint32, error) {
	epochDirs, err := ldp.getEpochDirs()
	if err != nil {
		return storage.LatestDataFromStorage{}, "", 0, err
	}

	for index := range epochDirs {
		parentDir, lastEpoch, errGetDir := ldp.getParentDirAndLastEpochWithIndex(index)
		if errGetDir != nil {
			err = errGetDir
			continue
		}

		dataFromStorage, errGetEpoch := ldp.getLastEpochAndRoundFromStorage(parentDir, lastEpoch)
		if errGetEpoch != nil {
			err = errGetEpoch
			continue
		}

		return dataFromStorage, parentDir, lastEpoch, nil
	}

	return storage.LatestDataFromStorage{}, "", 0, err
}

func (ldp *latestDataProvider) getEpochDirs() ([]string, error) {
	directoriesNames, err := ldp.directoryReader.ListDirectoriesAsString(ldp.parentDir)
	if err != nil {
		return nil, err
	}

	epochDirs := make([]string, 0, len(directoriesNames))
	for _, dirName := range directoriesNames {
		isEpochDir := strings.HasPrefix(dirName, ldp.defaultEpochString)
		if !isEpochDir {
			continue
		}

		epochDirs = append(epochDirs, dirName)
	}
	return epochDirs, nil
}

func (ldp *latestDataProvider) getLastEpochAndRoundFromStorage(parentDir string, lastEpoch uint32) (storage.LatestDataFromStorage, error) {
	persisterFactory := factory.NewPersisterFactory(ldp.generalConfig.BootstrapStorage.DB)
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
	shardIDAsUint32, err := core.ConvertShardIDToUint32(mostRecentShard)
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
		shardID, err = core.ConvertShardIDToUint32(shardIdStr)
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
	trigInternalKey := append([]byte(common.TriggerRegistryKeyPrefix), key...)
	trigData, err := storer.Get(trigInternalKey)
	if err != nil {
		return 0, err
	}

	var state *block.MetaTriggerRegistry
	marshaller := &marshal.GogoProtoMarshalizer{}
	if shardID == core.MetachainShardId {
		state, err = metachain.UnmarshalTrigger(marshaller, trigData)
		if err != nil {
			return 0, err
		}

		return state.CurrEpochStartRound, nil
	}

	var trigHandler data.TriggerRegistryHandler
	trigHandler, err = shardchain.UnmarshalTrigger(marshaller, trigData)
	if err != nil {
		return 0, err
	}

	return trigHandler.GetEpochStartRound(), nil
}

// GetLastEpochFromDirNames returns the last epoch found in storage directory
func (ldp *latestDataProvider) GetLastEpochFromDirNames(epochDirs []string, index int) (uint32, error) {
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

	return epochsInDirName[index], nil
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

func (ldp *latestDataProvider) getParentDirAndLastEpochWithIndex(index int) (string, uint32, error) {
	epochDirs, err := ldp.getEpochDirs()
	if err != nil {
		return "", 0, err
	}

	lastEpoch, err := ldp.GetLastEpochFromDirNames(epochDirs, index)
	if err != nil {
		return "", 0, err
	}

	return ldp.parentDir, lastEpoch, nil
}
