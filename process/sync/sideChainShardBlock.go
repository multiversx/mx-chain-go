package sync

import "github.com/ElrondNetwork/elrond-go/process"

// SideChainShardBootstrap implements the bootstrap mechanism
type SideChainShardBootstrap struct {
	*ShardBootstrap
}

// NewSideChainShardBootstrap creates a new Bootstrap object
func NewSideChainShardBootstrap(shardBootstrap *ShardBootstrap) (*SideChainShardBootstrap, error) {
	if shardBootstrap == nil {
		return nil, process.ErrNilShardBootstrap
	}

	scsb := &SideChainShardBootstrap{
		shardBootstrap,
	}

	scsb.processAndCommitFunc = scsb.sideChainProcessAndCommit
	return scsb, nil
}
