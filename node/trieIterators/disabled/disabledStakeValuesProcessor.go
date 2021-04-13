package disabled

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/data/api"
)

var errCannotReturnTotalStakedFromShardNode = errors.New("total staked value cannot be returned by a shard node")

type stakeValuesProcessor struct{}

// NewDisabledStakeValuesProcessor returns a disabled implementation to be used on shard nodes
func NewDisabledStakeValuesProcessor() *stakeValuesProcessor {
	return &stakeValuesProcessor{}
}

// GetTotalStakedValue returns the errCannotReturnTotalStakedFromShardNode error
func (svp *stakeValuesProcessor) GetTotalStakedValue() (*api.StakeValues, error) {
	return nil, errCannotReturnTotalStakedFromShardNode
}

// IsInterfaceNil returns true if there is no value under the interface
func (svp *stakeValuesProcessor) IsInterfaceNil() bool {
	return svp == nil
}
