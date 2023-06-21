package consensus

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap"
)

type ShardBootstrapFactory struct {
}

func (sbf *ShardBootstrapFactory) CreateShardBootstrapFactory(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
	return sync.NewShardBootstrap(argsBaseBootstrapper)
}

type ShardStorageBootstrapperFactory struct {
}

func (ssbf *ShardStorageBootstrapperFactory) CreateShardStorageBootstrapper(args storageBootstrap.ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error) {
	return storageBootstrap.NewShardStorageBootstrapper(args)
}
