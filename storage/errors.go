package storage

import (
	"errors"
	"strings"

	storageErrors "github.com/multiversx/mx-chain-storage-go/common"
)

// ErrInvalidNumberOfPersisters signals that an invalid number of persisters has been provided
var ErrInvalidNumberOfPersisters = errors.New("invalid number of active persisters")

// ErrInvalidNumberOfOldPersisters signals that an invalid number of old persisters has been provided
var ErrInvalidNumberOfOldPersisters = errors.New("invalid number of old active persisters")

// ErrNilEpochStartNotifier signals that a nil epoch start notifier has been provided
var ErrNilEpochStartNotifier = errors.New("nil epoch start notifier")

// ErrNilPersisterFactory signals that a nil persister factory has been provided
var ErrNilPersisterFactory = errors.New("nil persister factory")

// ErrDestroyingUnit signals that the destroy unit method did not manage to destroy all the persisters in a pruning storer
var ErrDestroyingUnit = errors.New("destroy unit didn't remove all the persisters")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilPathManager signals that a nil path manager has been provided
var ErrNilPathManager = errors.New("nil path manager")

// ErrNilCustomDatabaseRemover signals that a nil custom database remover has been provided
var ErrNilCustomDatabaseRemover = errors.New("custom database remover")

// ErrNilStorageListProvider signals that a nil storage list provided has been provided
var ErrNilStorageListProvider = errors.New("nil storage list provider")

// ErrInvalidNumberOfEpochsToSave signals that an invalid number of epochs to save has been provided
var ErrInvalidNumberOfEpochsToSave = errors.New("invalid number of epochs to save")

// ErrInvalidNumberOfActivePersisters signals that an invalid number of active persisters has been provided
var ErrInvalidNumberOfActivePersisters = errors.New("invalid number of active persisters")

// ErrClosingPersisters signals that not all persisters were closed
var ErrClosingPersisters = errors.New("cannot close all the persisters")

// ErrCacheSizeIsLowerThanBatchSize signals that size of cache is lower than size of batch
var ErrCacheSizeIsLowerThanBatchSize = errors.New("cache size is lower than batch size")

// ErrBootstrapDataNotFoundInStorage signals that no BootstrapData was find in the storage
var ErrBootstrapDataNotFoundInStorage = errors.New("didn't find any bootstrap data in storage")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer")

// ErrWrongTypeAssertion is thrown when a wrong type assertion is spotted
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrCannotComputeStorageOldestEpoch signals an issue when computing the oldest epoch for storage
var ErrCannotComputeStorageOldestEpoch = errors.New("could not compute the oldest epoch for storage")

// ErrNilNodeTypeProvider signals that a nil node type provider has been provided
var ErrNilNodeTypeProvider = errors.New("nil node type provider")

// ErrNilOldDataCleanerProvider signals that a nil old data cleaner provider has been provided
var ErrNilOldDataCleanerProvider = errors.New("nil old data cleaner provider")

// ErrKeyNotFound is raised when a key is not found
var ErrKeyNotFound = storageErrors.ErrKeyNotFound

// ErrInvalidConfig signals an invalid config
var ErrInvalidConfig = storageErrors.ErrInvalidConfig

// ErrCacheSizeInvalid signals that size of cache is less than 1
var ErrCacheSizeInvalid = storageErrors.ErrCacheSizeInvalid

// ErrNotSupportedDBType is raised when an unsupported database type is provided
var ErrNotSupportedDBType = storageErrors.ErrNotSupportedDBType

// ErrNotSupportedCacheType is raised when an unsupported cache type is provided
var ErrNotSupportedCacheType = storageErrors.ErrNotSupportedCacheType

// ErrDBIsClosed is raised when the DB is closed
var ErrDBIsClosed = storageErrors.ErrDBIsClosed

// ErrEpochKeepIsLowerThanNumActive signals that num epochs to keep is lower than num active epochs
var ErrEpochKeepIsLowerThanNumActive = errors.New("num epochs to keep is lower than num active epochs")

// ErrNilPersistersTracker signals that a nil persisters tracker has been provided
var ErrNilPersistersTracker = errors.New("nil persisters tracker provided")

// ErrNilStatsCollector signals that a nil stats collector has been provided
var ErrNilStatsCollector = errors.New("nil stats collector provided")

// ErrNilShardIDProvider signals that a nil shard id provider has been provided
var ErrNilShardIDProvider = errors.New("nil shard id provider")

// ErrNotSupportedShardIDProviderType is raised when an unsupported shard id provider type is provided
var ErrNotSupportedShardIDProviderType = errors.New("invalid shard id provider type has been provided")

// ErrInvalidFilePath signals that an invalid file path has been provided
var ErrInvalidFilePath = errors.New("invalid file path")

// ErrNilDBConfigHandler signals that a nil db config handler has been provided
var ErrNilDBConfigHandler = errors.New("nil db config handler")

// ErrNilManagedPeersHolder signals that a nil managed peers holder has been provided
var ErrNilManagedPeersHolder = errors.New("nil managed peers holder")

// ErrNilLatestStorageDataProvider signals that a nil latest storage data provider has been provided
var ErrNilLatestStorageDataProvider = errors.New("nil latest storage data provider")

// ErrNilBootstrapDataProvider signals that a nil bootstrap data provider has been provided
var ErrNilBootstrapDataProvider = errors.New("nil bootstrap data provider")

// ErrNilDirectoryReader signals that a nil directory reader has been provided
var ErrNilDirectoryReader = errors.New("nil directory reader")

// IsNotFoundInStorageErr returns whether an error is a "not found in storage" error.
// Currently, "item not found" storage errors are untyped (thus not distinguishable from others). E.g. see "pruningStorer.go".
// As a workaround, we test the error message for a match.
func IsNotFoundInStorageErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "not found")
}
