package sync

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type sideChainShardForkDetector struct {
	*shardForkDetector
}

// NewSideChainShardForkDetector creates an object for detecting the shard forks
func NewSideChainShardForkDetector(shardForkDetector *shardForkDetector) (*sideChainShardForkDetector, error) {
	if shardForkDetector == nil {
		return nil, process.ErrNilForkDetector
	}

	return &sideChainShardForkDetector{
		shardForkDetector,
	}, nil
}

func (scsfd *sideChainShardForkDetector) doJobOnBHProcessed(
	header data.HeaderHandler,
	headerHash []byte,
	_ []data.HeaderHandler,
	_ [][]byte,
) {
	scsfd.setFinalCheckpoint(scsfd.lastCheckpoint())
	scsfd.addCheckpoint(&checkpointInfo{nonce: header.GetNonce(), round: header.GetRound(), hash: headerHash})
	scsfd.removePastOrInvalidRecords()
}
