package storageBootstrap

import (
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignShardStorageBootstrapperFactory struct {
}

// NewSovereignShardStorageBootstrapperFactory creates a new instance of shardStorageBootstrapperFactory for run type sovereign
func NewSovereignShardStorageBootstrapperFactory() *sovereignShardStorageBootstrapperFactory {
	return &sovereignShardStorageBootstrapperFactory{}
}

// CreateBootstrapperFromStorage creates a new instance of shardStorageBootstrapperFactory for run type sovereign
func (ssbf *sovereignShardStorageBootstrapperFactory) CreateBootstrapperFromStorage(args ArgsShardStorageBootstrapper) (process.BootstrapperFromStorage, error) {
	ssb, err := NewShardStorageBootstrapper(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignChainShardStorageBootstrapper(ssb)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ssbf *sovereignShardStorageBootstrapperFactory) IsInterfaceNil() bool {
	return ssbf == nil
}
