package trieNodeData

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type intermediaryNodeData struct {
	*baseNodeData
}

// NewIntermediaryNodeData creates a new intermediary node data
func NewIntermediaryNodeData(key common.KeyBuilder, data []byte) (*intermediaryNodeData, error) {
	if check.IfNil(key) {
		return nil, ErrNilKeyBuilder
	}

	return &intermediaryNodeData{
		baseNodeData: &baseNodeData{
			keyBuilder: key,
			data:       data,
		},
	}, nil
}

// IsLeaf returns false
func (ind *intermediaryNodeData) IsLeaf() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (ind *intermediaryNodeData) IsInterfaceNil() bool {
	return ind == nil
}
