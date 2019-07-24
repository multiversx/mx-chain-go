package external

// ScDataGetter defines how data should be get from a SC account
type ScDataGetter interface {
	Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error)
}
