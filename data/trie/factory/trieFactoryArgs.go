package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const UserAccountTrie = "userAccount"
const PeerAccountTrie = "peerAccount"

// TrieFactoryArgs holds arguments for creating a trie factory
type TrieFactoryArgs struct {
	Cfg                    config.StorageConfig
	EvictionWaitingListCfg config.EvictionWaitingListConfig
	SnapshotDbCfg          config.DBConfig
	Marshalizer            marshal.Marshalizer
	Hasher                 hashing.Hasher
	PathManager            storage.PathManagerHandler
	ShardId                string
	PruningEnabled         bool
}
