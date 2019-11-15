package mock

type ScDataGetterMock struct {
	GetCalled func(scAddress []byte, funcName string, args ...[]byte) ([]byte, error)
}

func (scd *ScDataGetterMock) Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error) {
	if scd.GetCalled != nil {
		return scd.GetCalled(scAddress, funcName, args...)
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scd *ScDataGetterMock) IsInterfaceNil() bool {
	if scd == nil {
		return true
	}
	return false
}
