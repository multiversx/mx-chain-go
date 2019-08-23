package external

// NodeApiResolver can resolve API requests
type NodeApiResolver struct {
	scDataGetter ScDataGetter
}

// NewNodeApiResolver creates a new NodeApiResolver instance
func NewNodeApiResolver(scDataGetter ScDataGetter) (*NodeApiResolver, error) {
	if scDataGetter == nil {
		return nil, ErrNilScDataGetter
	}

	return &NodeApiResolver{
		scDataGetter: scDataGetter,
	}, nil
}

// GetVmValue retrieves data stored in a SC account through a VM
func (nar *NodeApiResolver) GetVmValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return nar.scDataGetter.Get([]byte(address), funcName, argsBuff...)
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *NodeApiResolver) IsInterfaceNil() bool {
	if nar == nil {
		return true
	}
	return false
}
