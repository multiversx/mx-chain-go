package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NodesConfigProviderStub -
type NodesConfigProviderStub struct {
	GetNodesConfigForMetaBlockCalled func(metaBlock *block.MetaBlock) (*sharding.NodesSetup, error)
}

// GetNodesConfigForMetaBlock -
func (n *NodesConfigProviderStub) GetNodesConfigForMetaBlock(metaBlock *block.MetaBlock) (*sharding.NodesSetup, error) {
	if n.GetNodesConfigForMetaBlockCalled != nil {
		return n.GetNodesConfigForMetaBlockCalled(metaBlock)
	}

	return &sharding.NodesSetup{}, nil
}

// IsInterfaceNil -
func (n *NodesConfigProviderStub) IsInterfaceNil() bool {
	return n == nil
}
