package disabled

// NodesCoordinator  implements the NodesCoordinator interface, it does nothing as it is disabled
type NodesCoordinator struct {
}

// GetNumTotalEligible -
func (n *NodesCoordinator) GetNumTotalEligible() uint64 {
	return 0
}

// IsInterfaceNil -
func (n *NodesCoordinator) IsInterfaceNil() bool {
	return n == nil
}
