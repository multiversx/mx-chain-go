package peer

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignChainValidatorStatistics struct {
	*validatorStatistics
}

// NewSovereignChainValidatorStatisticsProcessor instantiates a new sovereignChainValidatorStatistics structure
// responsible for keeping account of each validator actions in the consensus process
func NewSovereignChainValidatorStatisticsProcessor(validatorStatistics *validatorStatistics) (*sovereignChainValidatorStatistics, error) {
	if validatorStatistics == nil {
		return nil, process.ErrNilValidatorStatistics
	}

	scvs := &sovereignChainValidatorStatistics{
		validatorStatistics,
	}

	scvs.updateShardDataPeerStateFunc = scvs.updateShardDataPeerState
	return scvs, nil
}

// updateShardDataPeerState should not be implemented for sovereign, since UpdatePeerState is already taking care
// of updating validators/leader block ratings
func (vs *sovereignChainValidatorStatistics) updateShardDataPeerState(
	_ data.CommonHeaderHandler,
	_ map[string]data.CommonHeaderHandler,
) error {
	return nil
}
