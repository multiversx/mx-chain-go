package mock

// NodesConfigProviderStub -
type NodesConfigProviderStub struct {
	MinNumberOfNodesCalled func() uint32
}

// MinNumberOfNodes -
func (n *NodesConfigProviderStub) MinNumberOfNodes() uint32 {
	if n.MinNumberOfNodesCalled != nil {
		return n.MinNumberOfNodesCalled()
	}
	return 10
}

// IsInterfaceNil -
func (n *NodesConfigProviderStub) IsInterfaceNil() bool {
	return n == nil
}
