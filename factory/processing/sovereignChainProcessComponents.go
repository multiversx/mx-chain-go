package processing

import (
	errErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/sync"
	"github.com/ElrondNetwork/elrond-go/process/track"
)

type sovereignChainProcessComponentsFactory struct {
	*processComponentsFactory
}

// NewSovereignChainProcessComponentsFactory will return a new instance of sovereignChainProcessComponentsFactory
func NewSovereignChainProcessComponentsFactory(processComponentsFactory *processComponentsFactory) (*sovereignChainProcessComponentsFactory, error) {
	if processComponentsFactory == nil {
		return nil, errErd.ErrNilProcessComponentsFactory
	}

	scpcf := &sovereignChainProcessComponentsFactory{
		processComponentsFactory,
	}

	scpcf.createShardBlockTrackerMethod = scpcf.createShardBlockTracker
	scpcf.createShardForkDetectorMethod = scpcf.createShardForkDetector

	return scpcf, nil
}

func (scpcf *sovereignChainProcessComponentsFactory) createShardBlockTracker(argBaseTracker track.ArgBaseTracker) (process.BlockTracker, error) {
	arguments := track.ArgShardTracker{
		ArgBaseTracker: argBaseTracker,
	}

	shardBlockTrack, err := track.NewShardBlockTrack(arguments)
	if err != nil {
		return nil, err
	}

	return track.NewSovereignChainShardBlockTrack(shardBlockTrack)
}

func (scpcf *sovereignChainProcessComponentsFactory) createShardForkDetector(headerBlackList process.TimeCacher, blockTracker process.BlockTracker) (process.ForkDetector, error) {
	shardForkDetector, err := sync.NewShardForkDetector(scpcf.coreData.RoundHandler(), headerBlackList, blockTracker, scpcf.coreData.GenesisNodesSetup().GetStartTime())
	if err != nil {
		return nil, err
	}

	return sync.NewSovereignChainShardForkDetector(shardForkDetector)
}
