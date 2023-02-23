package node

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
)

// sovereignNodeRunner holds the sovereign node runner configuration and controls running of a node
type sovereignNodeRunner struct {
	*nodeRunner
}

// NewSovereignNodeRunner creates a sovereignNodeRunner instance
func NewSovereignNodeRunner(nodeRunner *nodeRunner) (*sovereignNodeRunner, error) {
	if nodeRunner == nil {
		return nil, ErrNilNodeRunner
	}

	snr := &sovereignNodeRunner{
		nodeRunner: nodeRunner,
	}

	snr.getConsensusModelFunc = snr.getConsensusModel
	snr.getChainRunTypeFunc = snr.getChainRunType

	return snr, nil
}

func (nr *sovereignNodeRunner) getConsensusModel() consensus.ConsensusModel {
	return consensus.ConsensusModelV2
}

func (snr *sovereignNodeRunner) getChainRunType() common.ChainRunType {
	return common.ChainRunTypeSovereign
}
