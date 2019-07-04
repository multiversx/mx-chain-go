package mock

type ApiResolverStub struct {
	GetVmValueHandler func(address string, funcName string, argsBuff ...[]byte) ([]byte, error)
}

func (ars *ApiResolverStub) GetVmValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return ars.GetVmValueHandler(address, funcName, argsBuff...)
}
