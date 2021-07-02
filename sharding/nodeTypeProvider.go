package sharding

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

type nodeTypeProvider struct {
	sync.RWMutex
	nodeType core.NodeType
}

// NewNodeTypeProvider returns a new instance of nodeTypeProvider
func NewNodeTypeProvider(initialType core.NodeType) *nodeTypeProvider {
	return &nodeTypeProvider{
		nodeType: initialType,
	}
}

// SetType will update the internal node type
func (ntp *nodeTypeProvider) SetType(nodeType core.NodeType) {
	ntp.Lock()
	ntp.nodeType = nodeType
	ntp.Unlock()
}

// GetType returns the node type
func (ntp *nodeTypeProvider) GetType() core.NodeType {
	ntp.RLock()
	defer ntp.RUnlock()

	return ntp.nodeType
}

// IsInterfaceNil returns true is there is no value under the interface
func (ntp *nodeTypeProvider) IsInterfaceNil() bool {
	return ntp == nil
}
