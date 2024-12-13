package storageBootstrap

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
)

type shardBootstrapFactory struct {
}

// NewShardBootstrapFactory creates a new instance of shardBootstrapFactory for run type normal
func NewShardBootstrapFactory() (*shardBootstrapFactory, error) {
	return &shardBootstrapFactory{}, nil
}

// CreateBootstrapper creates a new instance of shardBootstrapFactory for run type normal
func (sbf *shardBootstrapFactory) CreateBootstrapper(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
	return sync.NewShardBootstrap(argsBaseBootstrapper)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbf *shardBootstrapFactory) IsInterfaceNil() bool {
	return sbf == nil
}
