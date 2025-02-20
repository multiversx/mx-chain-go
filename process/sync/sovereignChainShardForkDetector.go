package sync

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignChainShardForkDetector struct {
	*shardForkDetector
}

// NewSovereignChainShardForkDetector creates an object for detecting the shard forks
func NewSovereignChainShardForkDetector(shardForkDetector *shardForkDetector) (*sovereignChainShardForkDetector, error) {
	if shardForkDetector == nil {
		return nil, process.ErrNilForkDetector
	}

	scsfd := &sovereignChainShardForkDetector{
		shardForkDetector,
	}

	scsfd.doJobOnBHProcessedFunc = scsfd.doJobOnBHProcessed

	return scsfd, nil
}

func (scsfd *sovereignChainShardForkDetector) doJobOnBHProcessed(
	header data.HeaderHandler,
	headerHash []byte,
	_ []data.HeaderHandler,
	_ [][]byte,
) {
	scsfd.setFinalCheckpoint(scsfd.lastCheckpoint())
	scsfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound(), hash: headerHash})
	scsfd.removePastOrInvalidRecords()
}
