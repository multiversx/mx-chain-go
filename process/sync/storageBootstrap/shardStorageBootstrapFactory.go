package storageBootstrap

import (
	"github.com/multiversx/mx-chain-go/process"
)

type shardStorageBootstrapperFactory struct {
}

// NewShardStorageBootstrapperFactory creates a new instance of shardStorageBootstrapperFactory for run type normal
func NewShardStorageBootstrapperFactory() (*shardStorageBootstrapperFactory, error) {
	return &shardStorageBootstrapperFactory{}, nil
}

// CreateBootstrapperFromStorage creates a new instance of shardStorageBootstrapperFactory for run type normal
func (ssbf *shardStorageBootstrapperFactory) CreateBootstrapperFromStorage(args ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error) {
	return NewShardStorageBootstrapper(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ssbf *shardStorageBootstrapperFactory) IsInterfaceNil() bool {
	return ssbf == nil
}
