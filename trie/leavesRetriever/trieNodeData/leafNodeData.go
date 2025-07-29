package trieNodeData

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type leafNodeData struct {
	*baseNodeData
	version core.TrieNodeVersion
}

// NewLeafNodeData creates a new leaf node data
func NewLeafNodeData(key common.KeyBuilder, data []byte, version core.TrieNodeVersion) (*leafNodeData, error) {
	if check.IfNil(key) {
		return nil, ErrNilKeyBuilder
	}

	return &leafNodeData{
		baseNodeData: &baseNodeData{
			keyBuilder: key,
			data:       data,
		},
		version: version,
	}, nil
}

// IsLeaf returns true
func (lnd *leafNodeData) IsLeaf() bool {
	return true
}

// GetVersion returns the version of the leaf
func (lnd *leafNodeData) GetVersion() core.TrieNodeVersion {
	return lnd.version
}

// IsInterfaceNil returns true if there is no value under the interface
func (lnd *leafNodeData) IsInterfaceNil() bool {
	return lnd == nil
}
