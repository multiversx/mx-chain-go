package mock

// NodesCoordinatorStub -
type NodesCoordinatorStub struct {
	GetNumTotalEligibleCalled func() uint64
}

// GetNumTotalEligible -
func (n *NodesCoordinatorStub) GetNumTotalEligible() uint64 {
	if n.GetNumTotalEligibleCalled != nil {
		return n.GetNumTotalEligibleCalled()
	}
	return 1000
}

// IsInterfaceNil -
func (n *NodesCoordinatorStub) IsInterfaceNil() bool {
	return n == nil
}
