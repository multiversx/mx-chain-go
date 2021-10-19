package factory

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
)

// TrieStorageCreator is used for creating and managing trie storages
type TrieStorageCreator interface {
	GetStorageForShard(shardId string, trieType string) (common.DBWriteCacher, error)
	GetSnapshotsConfig(shardId string, trieType string) config.DBConfig
	Close() error
	IsInterfaceNil() bool
}
