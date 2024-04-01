package trieNodeData

import "github.com/multiversx/mx-chain-go/common"

type intermediaryNodeData struct {
	*baseNodeData
}

func NewIntermediaryNodeData(key common.KeyBuilder, data []byte) *intermediaryNodeData {
	return &intermediaryNodeData{
		baseNodeData: &baseNodeData{
			keyBuilder: key,
			data:       data,
		},
	}
}

func (ind *intermediaryNodeData) IsLeaf() bool {
	return false
}
