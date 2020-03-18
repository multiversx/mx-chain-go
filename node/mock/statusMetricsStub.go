package mock

// StatusMetricsStub -
type StatusMetricsStub struct {
	StatusMetricsMapWithoutP2PCalled func() map[string]interface{}
	StatusP2pMetricsMapCalled        func() map[string]interface{}
	IsInterfaceNilCalled             func() bool
}

// StatusMetricsMap -
func (nds *StatusMetricsStub) StatusMetricsMapWithoutP2P() map[string]interface{} {
	return nds.StatusMetricsMapWithoutP2PCalled()
}

// StatusP2pMetricsMap -
func (nds *StatusMetricsStub) StatusP2pMetricsMap() map[string]interface{} {
	return nds.StatusP2pMetricsMapCalled()
}

// IsInterfaceNil -
func (nds *StatusMetricsStub) IsInterfaceNil() bool {
	return nds == nil
}
