package trieNodeData

import "github.com/multiversx/mx-chain-go/common"

type leafNodeData struct {
	*baseNodeData
}

func NewLeafNodeData(key common.KeyBuilder, data []byte) *leafNodeData {
	return &leafNodeData{
		baseNodeData: &baseNodeData{
			keyBuilder: key,
			data:       data,
		},
	}
}

func (lnd *leafNodeData) IsLeaf() bool {
	return true
}
