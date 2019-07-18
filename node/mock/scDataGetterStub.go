package mock

type ScDataGetterStub struct {
	GetCalled func(scAddress []byte, funcName string, args ...[]byte) ([]byte, error)
}

func (scds *ScDataGetterStub) Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error) {
	return scds.GetCalled(scAddress, funcName, args...)
}
