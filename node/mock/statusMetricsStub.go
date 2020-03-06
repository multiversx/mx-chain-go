package mock

// StatusMetricsStub -
type StatusMetricsStub struct {
	StatusMetricsMapCalled func() (map[string]interface{}, error)
	IsInterfaceNilCalled   func() bool
}

// StatusMetricsMap -
func (nds *StatusMetricsStub) StatusMetricsMap() (map[string]interface{}, error) {
	return nds.StatusMetricsMapCalled()
}

// IsInterfaceNil -
func (nds *StatusMetricsStub) IsInterfaceNil() bool {
	return nds == nil
}
