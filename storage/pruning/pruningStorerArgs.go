package pruning

import "github.com/ElrondNetwork/elrond-go/storage/storageUnit"

// PruningStorerArgs will hold the arguments needed for PruningStorer
type PruningStorerArgs struct {
	Identifier            string
	FullArchive           bool
	CacheConf             storageUnit.CacheConfig
	DbPath                string
	PersisterFactory      DbFactoryHandler
	BloomFilterConf       storageUnit.BloomConfig
	NumOfEpochsToKeep     uint32
	NumOfActivePersisters uint32
	Notifier              EpochStartNotifier
}
