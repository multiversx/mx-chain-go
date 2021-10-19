package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
)

// UserAccountTrie represents the use account identifier
const UserAccountTrie = "userAccount"

// PeerAccountTrie represents the peer account identifier
const PeerAccountTrie = "peerAccount"

// TrieFactoryArgs holds the arguments for creating a trie factory
type TrieFactoryArgs struct {
	TrieStorageCreator       TrieStorageCreator
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	TrieStorageManagerConfig config.TrieStorageManagerConfig
}
