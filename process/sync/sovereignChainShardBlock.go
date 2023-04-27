package sync

import "github.com/multiversx/mx-chain-go/process"

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
	scsb.handleScheduledRollBackToHeaderFunc = scsb.sovereignChainHandleScheduledRollBackToHeader
	scsb.getRootHashFromBlockFunc = scsb.sovereignChainGetRootHashFromBlock
	scsb.doProcessReceivedHeaderJobFunc = scsb.sovereignChainDoProcessReceivedHeaderJob

	return scsb, nil
}
