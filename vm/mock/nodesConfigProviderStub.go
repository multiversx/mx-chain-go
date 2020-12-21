package mock

// NodesConfigProviderStub -
type NodesConfigProviderStub struct {
	MinNumberOfNodesCalled               func() uint32
	MinNumberOfNodesWithHysteresisCalled func() uint32
}

// MinNumberOfNodes -
func (n *NodesConfigProviderStub) MinNumberOfNodes() uint32 {
	if n.MinNumberOfNodesCalled != nil {
		return n.MinNumberOfNodesCalled()
	}
	return 10
}

// MinNumberOfNodesWithHysteresis -
func (n *NodesConfigProviderStub) MinNumberOfNodesWithHysteresis() uint32 {
	if n.MinNumberOfNodesWithHysteresisCalled != nil {
		return n.MinNumberOfNodesWithHysteresisCalled()
	}
	return n.MinNumberOfNodes()
}

// IsInterfaceNil -
func (n *NodesConfigProviderStub) IsInterfaceNil() bool {
	return n == nil
}
