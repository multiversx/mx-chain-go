package nodeTypeProviderMock

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// NodeTypeProviderStub -
type NodeTypeProviderStub struct {
	SetTypeCalled func(nodeType core.NodeType)
	GetTypeCalled func() core.NodeType
}

// SetType -
func (n *NodeTypeProviderStub) SetType(nodeType core.NodeType) {
	if n.SetTypeCalled != nil {
		n.SetTypeCalled(nodeType)
	}
}

// GetType -
func (n *NodeTypeProviderStub) GetType() core.NodeType {
	if n.GetTypeCalled != nil {
		return n.GetTypeCalled()
	}

	return core.NodeTypeObserver
}

// IsInterfaceNil -
func (n *NodeTypeProviderStub) IsInterfaceNil() bool {
	return n == nil
}
