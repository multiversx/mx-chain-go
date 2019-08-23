package mock

type ApiResolverStub struct {
	GetVmValueHandler func(address string, funcName string, argsBuff ...[]byte) ([]byte, error)
}

func (ars *ApiResolverStub) GetVmValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return ars.GetVmValueHandler(address, funcName, argsBuff...)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	if ars == nil {
		return true
	}
	return false
}
