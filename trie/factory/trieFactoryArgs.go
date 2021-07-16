package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// UserAccountTrie represents the use account identifier
const UserAccountTrie = "userAccount"

// PeerAccountTrie represents the peer account identifier
const PeerAccountTrie = "peerAccount"

// TrieFactoryArgs holds arguments for creating a trie factory
type TrieFactoryArgs struct {
	SnapshotDbCfg            config.DBConfig
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	PathManager              storage.PathManagerHandler
	TrieStorageManagerConfig config.TrieStorageManagerConfig
}
