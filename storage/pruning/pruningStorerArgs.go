package pruning

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/clean"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
)

// StorerArgs will hold the arguments needed for PruningStorer
type StorerArgs struct {
	Identifier                string
	ShardCoordinator          storage.ShardCoordinator
	CacheConf                 storageunit.CacheConfig
	PathManager               storage.PathManagerHandler
	DbPath                    string
	PersisterFactory          DbFactoryHandler
	Notifier                  EpochStartNotifier
	OldDataCleanerProvider    clean.OldDataCleanerProvider
	CustomDatabaseRemover     storage.CustomDatabaseRemoverHandler
	MaxBatchSize              int
	EpochsData                *EpochArgs
	PruningEnabled            bool
	EnabledDbLookupExtensions bool
	PersistersTracker         PersistersTracker
}

// EpochArgs will hold the arguments needed for persistersTracker
type EpochArgs struct {
	NumOfEpochsToKeep     uint32
	NumOfActivePersisters uint32
	StartingEpoch         uint32
}

// FullHistoryStorerArgs will hold the arguments needed for full history PruningStorer
type FullHistoryStorerArgs struct {
	*StorerArgs
	NumOfOldActivePersisters uint32
}
