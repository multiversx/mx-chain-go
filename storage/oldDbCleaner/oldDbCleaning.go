package oldDbCleaner

import (
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
)

var log = logger.GetOrCreate("storage/oldDbCleaner")

// Args holds the arguments for deletion of the old databases
type Args struct {
	DirectoryReader      storage.DirectoryReaderHandler
	FileRemover          func(file string) error
	WorkingDir           string
	ChainID              string
	PruningStorageConfig config.StoragePruningConfig
	EpochStartTrigger    process.EpochStartTriggerHandler
}

// CleanOldStorageForShuffleOut handles the deletion of old databases in regards to the pruning storer configuration.
// DirectoryReader and FileRemover arguments can be omitted and default I/O handlers will be used.
func CleanOldStorageForShuffleOut(args Args) error {
	if !args.PruningStorageConfig.CleanOldEpochsData {
		return nil
	}

	if check.IfNil(args.EpochStartTrigger) {
		return storage.ErrNilEpochStartTrigger
	}

	if check.IfNil(args.DirectoryReader) {
		args.DirectoryReader = storageFactory.NewDirectoryReader()
	}
	if args.FileRemover == nil {
		args.FileRemover = func(file string) error {
			return os.RemoveAll(file)
		}
	}
	dbPath := path.Join(args.WorkingDir, core.DefaultDBPath, args.ChainID)
	epochDirectories, err := args.DirectoryReader.ListDirectoriesAsString(dbPath)
	if err != nil {
		return err
	}

	if len(epochDirectories) == 0 {
		return nil
	}

	sortedEpochDirectories, sortedEpochs, found := getSortedEpochDirectories(epochDirectories)
	if !found {
		return nil
	}

	epochToRemoveTo := args.EpochStartTrigger.Epoch() - uint32(args.PruningStorageConfig.NumEpochsToKeep)

	for idx, epoch := range sortedEpochs {
		if epoch >= epochToRemoveTo {
			break
		}

		fullDirectoryPath := path.Join(dbPath, sortedEpochDirectories[idx])
		err := args.FileRemover(fullDirectoryPath)
		if err != nil {
			log.Warn("cannot remove old DB", "error", err)
		}
	}

	return nil
}

func getSortedEpochDirectories(directories []string) ([]string, []uint32, bool) {
	epochToDirMap := make(map[uint32]string)
	epochs := make([]uint32, 0)
	for _, dir := range directories {
		epoch, ok := extractEpochFromDirName(dir)
		if !ok {
			continue
		}

		epochToDirMap[epoch] = dir
		epochs = append(epochs, epoch)
	}

	sort.SliceStable(epochs, func(i, j int) bool {
		return epochs[i] < epochs[j]
	})

	sortedDirectories := make([]string, 0, len(epochs))
	for _, epoch := range epochs {
		sortedDirectories = append(sortedDirectories, epochToDirMap[epoch])
	}

	notOkResults := len(sortedDirectories) != len(epochs) || len(sortedDirectories) == 0 || len(epochs) == 0
	if notOkResults {
		return nil, nil, false
	}

	return sortedDirectories, epochs, true
}

func extractEpochFromDirName(dirName string) (uint32, bool) {
	splitStr := strings.Split(dirName, "_")
	if len(splitStr) != 2 {
		return 0, false
	}

	epoch, err := strconv.ParseInt(splitStr[1], 10, 64)
	if err != nil {
		return 0, false
	}

	return uint32(epoch), true
}
