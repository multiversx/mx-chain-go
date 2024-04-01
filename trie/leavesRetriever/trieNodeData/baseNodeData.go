package trieNodeData

import "github.com/multiversx/mx-chain-go/common"

type baseNodeData struct {
	keyBuilder common.KeyBuilder
	data       []byte
}

func (bnd *baseNodeData) GetData() []byte {
	return bnd.data
}

func (bnd *baseNodeData) GetKeyBuilder() common.KeyBuilder {
	return bnd.keyBuilder
}

func (bnd *baseNodeData) Size() uint64 {
	return uint64(len(bnd.data)) + uint64(bnd.keyBuilder.Size())
}
