package storageBootstrap

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
)

type sovereignShardBootstrapFactory struct {
}

// NewSovereignShardBootstrapFactory creates a new instance of shardBootstrapFactory for run type sovereign
func NewSovereignShardBootstrapFactory() *sovereignShardBootstrapFactory {
	return &sovereignShardBootstrapFactory{}
}

// CreateBootstrapper creates a new instance of shardBootstrapFactory for run type sovereign
func (sbf *sovereignShardBootstrapFactory) CreateBootstrapper(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
	bootstrapper, err := sync.NewShardBootstrap(argsBaseBootstrapper)
	if err != nil {
		return nil, err
	}

	return sync.NewSovereignChainShardBootstrap(bootstrapper)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbf *sovereignShardBootstrapFactory) IsInterfaceNil() bool {
	return sbf == nil
}
