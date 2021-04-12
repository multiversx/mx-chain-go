package disabled

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/data/api"
)

var errCannotReturnDelegatedListFromShardNode = errors.New("total staked value cannot be returned by a shard node")

type delegatedListProcessor struct{}

// NewDisabledDelegatedListProcessor returns a disabled implementation to be used on shard nodes
func NewDisabledDelegatedListProcessor() *delegatedListProcessor {
	return &delegatedListProcessor{}
}

// GetDelegatorsList returns the errCannotReturnDelegatedListFromShardNode error
func (dlp *delegatedListProcessor) GetDelegatorsList() ([]*api.Delegator, error) {
	return nil, errCannotReturnDelegatedListFromShardNode
}

// IsInterfaceNil returns true if there is no value under the interface
func (dlp *delegatedListProcessor) IsInterfaceNil() bool {
	return dlp == nil
}
