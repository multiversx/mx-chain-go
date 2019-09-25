package external

// NodeApiResolver can resolve API requests
type NodeApiResolver struct {
	scDataGetter         ScDataGetter
	statusMetricsHandler StatusMetricsHandler
}

// NewNodeApiResolver creates a new NodeApiResolver instance
func NewNodeApiResolver(scDataGetter ScDataGetter, statusMetricsHandler StatusMetricsHandler) (*NodeApiResolver, error) {
	if scDataGetter == nil || scDataGetter.IsInterfaceNil() {
		return nil, ErrNilScDataGetter
	}
	if statusMetricsHandler == nil || statusMetricsHandler.IsInterfaceNil() {
		return nil, ErrNilStatusMetrics
	}

	return &NodeApiResolver{
		scDataGetter:         scDataGetter,
		statusMetricsHandler: statusMetricsHandler,
	}, nil
}

// GetVmValue retrieves data stored in a SC account through a VM
func (nar *NodeApiResolver) GetVmValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return nar.scDataGetter.Get([]byte(address), funcName, argsBuff...)
}

// StatusMetrics returns an implementation of the StatusMetricsHandler interface
func (nar *NodeApiResolver) StatusMetrics() StatusMetricsHandler {
	return nar.statusMetricsHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *NodeApiResolver) IsInterfaceNil() bool {
	if nar == nil {
		return true
	}
	return false
}
