package storageBootstrap

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/sync"
)

// StorageBootstrapper is the main interface for bootstrap from storage execution engine
type storageBootstrapperHandler interface {
	getHeader(hash []byte) (data.HeaderHandler, error)
	getHeaderWithNonce(nonce uint64, shardID uint32) (data.HeaderHandler, []byte, error)
	applyCrossNotarizedHeaders(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) error
	applyNumPendingMiniBlocks(pendingMiniBlocks []bootstrapStorage.PendingMiniBlocksInfo)
	applySelfNotarizedHeaders(selfNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo) ([]data.HeaderHandler, [][]byte, error)
	cleanupNotarizedStorage(hash []byte)
	cleanupNotarizedStorageForHigherNoncesIfExist(crossNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo)
	getRootHash(hash []byte) []byte
	IsInterfaceNil() bool
}

// BootstrapperFromStorageCreator defines the operations supported by a shard storage bootstrapper factory
type BootstrapperFromStorageCreator interface {
	CreateBootstrapperFromStorage(args ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error)
	IsInterfaceNil() bool
}

// BootstrapperCreator defines the operations supported by a shard bootstrap factory
type BootstrapperCreator interface {
	CreateBootstrapper(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error)
	IsInterfaceNil() bool
}
