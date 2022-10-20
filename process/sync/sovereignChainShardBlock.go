package sync

import "github.com/ElrondNetwork/elrond-go/process"

// SovereignChainShardBootstrap implements the bootstrap mechanism
type SovereignChainShardBootstrap struct {
	*ShardBootstrap
}

// NewSovereignChainShardBootstrap creates a new Bootstrap object
func NewSovereignChainShardBootstrap(shardBootstrap *ShardBootstrap) (*SovereignChainShardBootstrap, error) {
	if shardBootstrap == nil {
		return nil, process.ErrNilShardBootstrap
	}

	scsb := &SovereignChainShardBootstrap{
		shardBootstrap,
	}

	scsb.processAndCommitFunc = scsb.sovereignChainProcessAndCommit

	return scsb, nil
}
