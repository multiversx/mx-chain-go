package pruning

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/clean"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// StorerArgs will hold the arguments needed for PruningStorer
type StorerArgs struct {
	Identifier                string
	ShardCoordinator          storage.ShardCoordinator
	CacheConf                 storageUnit.CacheConfig
	PathManager               storage.PathManagerHandler
	DbPath                    string
	PersisterFactory          DbFactoryHandler
	Notifier                  EpochStartNotifier
	OldDataCleanerProvider    clean.OldDataCleanerProvider
	MaxBatchSize              int
	NumOfEpochsToKeep         uint32
	NumOfActivePersisters     uint32
	StartingEpoch             uint32
	PruningEnabled            bool
	EnabledDbLookupExtensions bool
}

// FullHistoryStorerArgs will hold the arguments needed for full history PruningStorer
type FullHistoryStorerArgs struct {
	*StorerArgs
	NumOfOldActivePersisters uint32
}
