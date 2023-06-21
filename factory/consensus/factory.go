package consensus

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
)

type ShardBootstrapFactory struct {
}

func (sbf *ShardBootstrapFactory) CreateShardBootstrapFactory(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
	return sync.NewShardBootstrap(argsBaseBootstrapper)
}
