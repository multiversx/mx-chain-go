package mock

// NodesConfigProviderStub -
type NodesConfigProviderStub struct {
	MinNumberOfNodesCalled        func() uint32
	MinShardHysteresisNodesCalled func() uint32
	MinMetaHysteresisNodesCalled  func() uint32
}

// MinNumberOfNodes -
func (n *NodesConfigProviderStub) MinNumberOfNodes() uint32 {
	if n.MinNumberOfNodesCalled != nil {
		return n.MinNumberOfNodesCalled()
	}
	return 10
}

// MinShardHysteresisNodes -
func (n *NodesConfigProviderStub) MinShardHysteresisNodes() uint32 {
	if n.MinShardHysteresisNodesCalled != nil {
		return n.MinShardHysteresisNodesCalled()
	}
	return 0
}

// MinMetaHysteresisNodes -
func (n *NodesConfigProviderStub) MinMetaHysteresisNodes() uint32 {
	if n.MinMetaHysteresisNodesCalled != nil {
		return n.MinMetaHysteresisNodesCalled()
	}
	return 0
}

// IsInterfaceNil -
func (n *NodesConfigProviderStub) IsInterfaceNil() bool {
	return n == nil
}
