package nodesconfigprovider

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type simpleNodesConfigProvider struct {
	originalNodesConfig *sharding.NodesSetup
}

// NewSimpleNodesConfigProvider returns a new instance of simpleNodesConfigProvider
func NewSimpleNodesConfigProvider(originalNodesConfig *sharding.NodesSetup) *simpleNodesConfigProvider {
	return &simpleNodesConfigProvider{
		originalNodesConfig: originalNodesConfig,
	}
}

// GetNodesConfigForMetaBlock will return the original nodes setup
func (sncp *simpleNodesConfigProvider) GetNodesConfigForMetaBlock(_ *block.MetaBlock) (*sharding.NodesSetup, error) {
	return sncp.originalNodesConfig, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sncp *simpleNodesConfigProvider) IsInterfaceNil() bool {
	return sncp == nil
}
