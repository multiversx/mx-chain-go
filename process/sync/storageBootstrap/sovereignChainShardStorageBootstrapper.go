package storageBootstrap

import "github.com/ElrondNetwork/elrond-go/process"

type sovereignChainShardStorageBootstrapper struct {
	*shardStorageBootstrapper
}

// NewSovereignChainShardStorageBootstrapper creates a new instance of sovereignChainShardStorageBootstrapper
func NewSovereignChainShardStorageBootstrapper(shardStorageBootstrapper *shardStorageBootstrapper) (*sovereignChainShardStorageBootstrapper, error) {
	if shardStorageBootstrapper == nil {
		return nil, process.ErrNilShardStorageBootstrapper
	}

	scssb := &sovereignChainShardStorageBootstrapper{
		shardStorageBootstrapper,
	}

	scssb.getScheduledRootHashMethod = scssb.sovereignChainGetScheduledRootHash
	scssb.setScheduledInfoMethod = scssb.sovereignChainSetScheduledInfo

	return scssb, nil
}
