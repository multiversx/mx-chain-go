package factory

import (
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

// UserAccountTrie represents the use account identifier
const UserAccountTrie = "userAccount"

// PeerAccountTrie represents the peer account identifier
const PeerAccountTrie = "peerAccount"

// TrieFactoryArgs holds the arguments for creating a trie factory
type TrieFactoryArgs struct {
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	PathManager              storage.PathManagerHandler
	TrieStorageManagerConfig config.TrieStorageManagerConfig
}
