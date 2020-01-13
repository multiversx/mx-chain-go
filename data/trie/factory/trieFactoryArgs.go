package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type trieFactoryArgs struct {
	cfg                    config.StorageConfig
	evictionWaitingListCfg config.EvictionWaitingListConfig
	snapshotDbCfg          config.DBConfig
	marshalizer            marshal.Marshalizer
	hasher                 hashing.Hasher
	pathManager            storage.PathManagerHandler
	shardId                string
	pruningEnabled         bool
}

// NewTrieFactoryArgs returns a struct that holds all arguments needed for creating a trie factory
func NewTrieFactoryArgs(
	cfg config.StorageConfig,
	evictionWaitingListCfg config.EvictionWaitingListConfig,
	snapshotDbCfg config.DBConfig,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	pathManager storage.PathManagerHandler,
	shardId string,
	pruningEnabled bool,
) *trieFactoryArgs {
	return &trieFactoryArgs{
		cfg:                    cfg,
		evictionWaitingListCfg: evictionWaitingListCfg,
		snapshotDbCfg:          snapshotDbCfg,
		marshalizer:            marshalizer,
		hasher:                 hasher,
		pathManager:            pathManager,
		shardId:                shardId,
		pruningEnabled:         pruningEnabled,
	}
}
