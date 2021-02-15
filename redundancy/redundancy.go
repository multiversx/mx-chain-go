package redundancy

type nodeRedundancy struct {
	redundancyLevel uint64
}

// NewNodeRedundancy creates a node redundancy object which implements NodeRedundancyHandler interface
func NewNodeRedundancy(redundancyLevel uint64) (*nodeRedundancy, error) {
	nr := &nodeRedundancy{
		redundancyLevel: redundancyLevel,
	}
	return nr, nil
}

// IsRedundancyNode returns true if the current instance is used as a redundancy node
func (nr *nodeRedundancy) IsRedundancyNode() bool {
	return nr.redundancyLevel > 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (nr *nodeRedundancy) IsInterfaceNil() bool {
	return nr == nil
}
