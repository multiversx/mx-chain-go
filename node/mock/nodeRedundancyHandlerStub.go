package mock

// NodeRedundancyHandlerStub -
type NodeRedundancyHandlerStub struct {
	IsRedundancyNodeCalled func() bool
}

// IsRedundancyNode -
func (nrhs *NodeRedundancyHandlerStub) IsRedundancyNode() bool {
	if nrhs.IsRedundancyNodeCalled != nil {
		return nrhs.IsRedundancyNodeCalled()
	}
	return false
}

// IsInterfaceNil -
func (nrhs *NodeRedundancyHandlerStub) IsInterfaceNil() bool {
	return nrhs == nil
}
