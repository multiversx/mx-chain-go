package mock

type StatusMetricsStub struct {
	StatusMetricsMapCalled func() (map[string]interface{}, error)
	IsInterfaceNilCalled   func() bool
}

func (nds *StatusMetricsStub) StatusMetricsMap() (map[string]interface{}, error) {
	return nds.StatusMetricsMapCalled()
}

func (nds *StatusMetricsStub) IsInterfaceNil() bool {
	if nds == nil {
		return true
	}
	return false
}
