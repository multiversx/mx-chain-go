package clean

import (
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/directoryhandler"
	logger "github.com/multiversx/mx-chain-logger-go"
	storageErrors "github.com/multiversx/mx-chain-storage-go/common"
)

var log = logger.GetOrCreate("storage/clean")

// ArgsOldDatabaseCleaner holds the arguments needed for creating an oldDatabaseCleaner
type ArgsOldDatabaseCleaner struct {
	DatabasePath           string
	StorageListProvider    StorageListProviderHandler
	EpochStartNotifier     EpochStartNotifier
	OldDataCleanerProvider OldDataCleanerProvider
}

type oldDatabaseCleaner struct {
	sync.RWMutex

	databasePath           string
	storageListProvider    StorageListProviderHandler
	pathRemover            func(file string) error
	directoryReader        storage.DirectoryReaderHandler
	oldDataCleanerProvider OldDataCleanerProvider
	oldestEpochsToKeep     map[uint32]uint32
}

// NewOldDatabaseCleaner returns a new instance of oldDatabaseCleaner
func NewOldDatabaseCleaner(args ArgsOldDatabaseCleaner) (*oldDatabaseCleaner, error) {
	if check.IfNil(args.StorageListProvider) {
		return nil, storage.ErrNilStorageListProvider
	}
	if check.IfNil(args.EpochStartNotifier) {
		return nil, storage.ErrNilEpochStartNotifier
	}
	if check.IfNil(args.OldDataCleanerProvider) {
		return nil, storage.ErrNilOldDataCleanerProvider
	}

	pathRemoverFunc := func(file string) error {
		return os.RemoveAll(file)
	}
	directoryReader := directoryhandler.NewDirectoryReader()

	odc := &oldDatabaseCleaner{
		databasePath:           args.DatabasePath,
		storageListProvider:    args.StorageListProvider,
		pathRemover:            pathRemoverFunc,
		directoryReader:        directoryReader,
		oldDataCleanerProvider: args.OldDataCleanerProvider,
		oldestEpochsToKeep:     make(map[uint32]uint32),
	}

	odc.registerHandler(args.EpochStartNotifier)

	return odc, nil
}

// registerHandler will register a new function to the epoch start notifier
func (odc *oldDatabaseCleaner) registerHandler(handler EpochStartNotifier) {
	subscribeHandler := notifier.NewHandlerForEpochStart(
		odc.epochChangeActionHandler,
		odc.epochChangePrepareHandler,
		common.OldDatabaseCleanOrder)

	handler.RegisterHandler(subscribeHandler)
}

func (odc *oldDatabaseCleaner) epochChangeActionHandler(hdr data.HeaderHandler) {
	err := odc.handleEpochChangeAction(hdr.GetEpoch())
	if err != nil {
		log.Debug("oldDatabaseCleaner: handleEpochChangeAction", "error", err)
	}
}

func (odc *oldDatabaseCleaner) epochChangePrepareHandler(_ data.HeaderHandler) {
}

func (odc *oldDatabaseCleaner) handleEpochChangeAction(epoch uint32) error {
	newOldestEpoch, err := odc.computeOldestEpochToKeep()
	if err != nil {
		return err
	}

	odc.Lock()
	odc.oldestEpochsToKeep[epoch] = newOldestEpoch
	odc.Unlock()

	shouldClean := odc.shouldCleanOldData(epoch, newOldestEpoch)
	log.Debug("old database cleaner", "epoch", epoch, "should clean", shouldClean, "oldest epoch", newOldestEpoch, "inner map", odc.oldestEpochsToKeep)
	if !shouldClean {
		return nil
	}

	err = odc.cleanOldEpochs(epoch)
	if err != nil {
		return err
	}

	return nil
}

func (odc *oldDatabaseCleaner) shouldCleanOldData(currentEpoch uint32, newOldestEpoch uint32) bool {
	odc.RLock()
	defer odc.RUnlock()

	if !odc.oldDataCleanerProvider.ShouldClean() {
		return false
	}

	epochForDeletion := currentEpoch - 1
	previousOldestEpoch, epochExists := odc.oldestEpochsToKeep[epochForDeletion]
	if !epochExists {
		log.Debug("cannot delete old databases as the previous epoch does not exist in configuration",
			"epoch", epochForDeletion,
			"epochs configuration", odc.oldestEpochsToKeep)
		return false
	}
	if previousOldestEpoch > newOldestEpoch {
		log.Debug("skipping cleaning of old databases because new oldest epoch is older than previous",
			"previous older epoch", previousOldestEpoch,
			"actual older epoch", newOldestEpoch)
		return false
	}

	return true
}

func (odc *oldDatabaseCleaner) computeOldestEpochToKeep() (uint32, error) {
	odc.RLock()
	defer odc.RUnlock()

	oldestEpoch := uint32(math.MaxUint32)
	storers := odc.storageListProvider.GetAllStorers()
	for _, storer := range storers {
		localEpoch, err := storer.GetOldestEpoch()
		if err != nil {
			logOldestEpochCompute(err)
			continue
		}

		if localEpoch < oldestEpoch {
			oldestEpoch = localEpoch
		}
	}

	if oldestEpoch == uint32(math.MaxUint32) {
		return 0, storage.ErrCannotComputeStorageOldestEpoch
	}

	return oldestEpoch, nil
}

func logOldestEpochCompute(err error) {
	if err != storageErrors.ErrOldestEpochNotAvailable {
		log.Debug("cannot compute oldest epoch for storer", "error", err)
	}
}

func (odc *oldDatabaseCleaner) cleanOldEpochs(currentEpoch uint32) error {
	odc.Lock()
	defer odc.Unlock()

	epochForDeletion := currentEpoch - 1
	epochToDeleteTo := odc.oldestEpochsToKeep[epochForDeletion]

	epochDirectories, err := odc.directoryReader.ListDirectoriesAsString(odc.databasePath)
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

	for idx, epoch := range sortedEpochs {
		if epoch >= epochToDeleteTo {
			break
		}

		fullDirectoryPath := path.Join(odc.databasePath, sortedEpochDirectories[idx])
		log.Debug("removing old database", "db path", fullDirectoryPath)
		err := odc.pathRemover(fullDirectoryPath)
		if err != nil {
			log.Warn("cannot remove old DB", "path", fullDirectoryPath, "error", err)
		}
	}

	odc.cleanMap(currentEpoch)

	return nil
}

// cleanMap will remove all the entries from the map that aren't for current epoch.
// should be called under mutex protection
func (odc *oldDatabaseCleaner) cleanMap(currentEpoch uint32) {
	epochsToRemove := make([]uint32, 0)
	for epoch := range odc.oldestEpochsToKeep {
		if epoch == currentEpoch {
			continue
		}

		epochsToRemove = append(epochsToRemove, epoch)
	}

	for _, epochToRemove := range epochsToRemove {
		delete(odc.oldestEpochsToKeep, epochToRemove)
	}
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

// IsInterfaceNil returns true if there is no value under interface
func (odc *oldDatabaseCleaner) IsInterfaceNil() bool {
	return odc == nil
}
