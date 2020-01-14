package pruning

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// StorerArgs will hold the arguments needed for PruningStorer
type StorerArgs struct {
	Identifier            string
	PruningEnabled        bool
	ShardCoordinator      sharding.Coordinator
	StartingEpoch         uint32
	FullArchive           bool
	CacheConf             storageUnit.CacheConfig
	PathManager           storage.PathManagerHandler
	DbPath                string
	PersisterFactory      DbFactoryHandler
	BloomFilterConf       storageUnit.BloomConfig
	NumOfEpochsToKeep     uint32
	NumOfActivePersisters uint32
	Notifier              EpochStartNotifier
}
