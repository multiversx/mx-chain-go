package trieNodeData

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type leafNodeData struct {
	*baseNodeData
}

// NewLeafNodeData creates a new leaf node data
func NewLeafNodeData(key common.KeyBuilder, data []byte) (*leafNodeData, error) {
	if check.IfNil(key) {
		return nil, ErrNilKeyBuilder
	}

	return &leafNodeData{
		baseNodeData: &baseNodeData{
			keyBuilder: key,
			data:       data,
		},
	}, nil
}

// IsLeaf returns true
func (lnd *leafNodeData) IsLeaf() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (lnd *leafNodeData) IsInterfaceNil() bool {
	return lnd == nil
}
