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

// GetDataValue retrieves data stored in a SC account
func (nar *NodeApiResolver) GetDataValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return nar.scDataGetter.Get([]byte(address), funcName, argsBuff...)
}
