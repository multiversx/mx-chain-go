package mock

type ScDataGetterStub struct {
	GetCalled func(scAddress []byte, funcName string, args ...[]byte) ([]byte, error)
}

func (scds *ScDataGetterStub) Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error) {
	return scds.GetCalled(scAddress, funcName, args...)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scds *ScDataGetterStub) IsInterfaceNil() bool {
	if scds == nil {
		return true
	}
	return false
}
