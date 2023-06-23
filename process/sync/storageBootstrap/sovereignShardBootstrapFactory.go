package storageBootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
)

type sovereignShardBootstrapFactory struct {
	shardBootstrapFactory ShardBootstrapFactoryHandler
}

// NewSovereignShardBootstrapFactory creates a new instance of shardBootstrapFactory for run type sovereign
func NewSovereignShardBootstrapFactory(sbf ShardBootstrapFactoryHandler) (*sovereignShardBootstrapFactory, error) {
	if check.IfNil(sbf) {
		return nil, errors.ErrNilShardBootstrapFactory
	}
	return &sovereignShardBootstrapFactory{
		shardBootstrapFactory: sbf,
	}, nil
}

// CreateShardBootstrapFactory creates a new instance of shardBootstrapFactory for run type sovereign
func (sbf *sovereignShardBootstrapFactory) CreateShardBootstrapFactory(argsBaseBootstrapper sync.ArgShardBootstrapper) (process.Bootstrapper, error) {
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
