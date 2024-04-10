package trieNodeData

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
)

type baseNodeData struct {
	keyBuilder common.KeyBuilder
	data       []byte
}

// GetData returns the bytes stored in the data field
func (bnd *baseNodeData) GetData() []byte {
	return bnd.data
}

// GetKeyBuilder returns the keyBuilder
func (bnd *baseNodeData) GetKeyBuilder() common.KeyBuilder {
	return bnd.keyBuilder
}

// Size returns the size of the data field combined with the keyBuilder size
func (bnd *baseNodeData) Size() uint64 {
	keyBuilderSize := uint(0)
	if !check.IfNil(bnd.keyBuilder) {
		keyBuilderSize = bnd.keyBuilder.Size()
	}

	return uint64(len(bnd.data)) + uint64(keyBuilderSize)
}
