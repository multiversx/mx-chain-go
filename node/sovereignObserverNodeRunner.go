package node

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
)

// sovereignObserverNodeRunner holds the sovereign observer node runner configuration and controls running of a node
type sovereignObserverNodeRunner struct {
	*nodeRunner
}

// NewSovereignObserverNodeRunner creates a sovereignObserverNodeRunner instance
func NewSovereignObserverNodeRunner(nodeRunner *nodeRunner) (*sovereignObserverNodeRunner, error) {
	if nodeRunner == nil {
		return nil, ErrNilNodeRunner
	}

	snr := &sovereignObserverNodeRunner{
		nodeRunner: nodeRunner,
	}

	snr.getConsensusModelFunc = snr.getConsensusModel
	snr.getChainRunTypeFunc = snr.getChainRunType

	return snr, nil
}

func (nr *sovereignObserverNodeRunner) getConsensusModel() consensus.ConsensusModel {
	return consensus.ConsensusModelV1
}

func (snr *sovereignObserverNodeRunner) getChainRunType() common.ChainRunType {
	return common.ChainRunTypeSovereignObserver
}
