package external

// ScDataGetter defines how data should be get from a SC account
type ScDataGetter interface {
	Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error)
	IsInterfaceNil() bool
}

// StatusMetricsHandler is the interface that defines what a node details handler/provider should do
type StatusMetricsHandler interface {
	StatusMetricsMap() (map[string]interface{}, error)
	IsInterfaceNil() bool
}
