package track

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

// BlockTrackerCreator is an interface for creating block trackers
type BlockTrackerCreator interface {
	CreateBlockTracker(argBaseTracker ArgShardTracker) (process.BlockTracker, error)
	IsInterfaceNil() bool
}

type sovereignBlockTrackerFactory struct {
	blockTrackerCreator BlockTrackerCreator
}

// NewSovereignBlockTrackerFactory creates a new shard block tracker factory for sovereign chain
func NewSovereignBlockTrackerFactory(btc BlockTrackerCreator) (*sovereignBlockTrackerFactory, error) {
	if check.IfNil(btc) {
		return nil, process.ErrNilBlockTrackerCreator
	}
	return &sovereignBlockTrackerFactory{
		blockTrackerCreator: btc,
	}, nil
}

// CreateBlockTracker creates a new block tracker for sovereign chain
func (sbtcf *sovereignBlockTrackerFactory) CreateBlockTracker(argBaseTracker ArgShardTracker) (process.BlockTracker, error) {
	blockTracker, err := sbtcf.blockTrackerCreator.CreateBlockTracker(argBaseTracker)
	if err != nil {
		return nil, err
	}
	shardBlockTracker, ok := blockTracker.(*shardBlockTrack)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return NewSovereignChainShardBlockTrack(shardBlockTracker)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbtcf *sovereignBlockTrackerFactory) IsInterfaceNil() bool {
	return sbtcf == nil
}
