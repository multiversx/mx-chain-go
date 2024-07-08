package process

type vmNodesCoordinator struct {
	numEligible uint64
}

// GetNumTotalEligible returns the number of eligible nodes
func (vnc *vmNodesCoordinator) GetNumTotalEligible() uint64 {
	return vnc.numEligible
}

// IsInterfaceNil checks if the underlying pointer is nil
func (vnc *vmNodesCoordinator) IsInterfaceNil() bool {
	return vnc == nil
}
