package mock

type ApiResolverStub struct {
	GetDataValueHandler func(address string, funcName string, argsBuff ...[]byte) ([]byte, error)
}

func (ars *ApiResolverStub) GetDataValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return ars.GetDataValueHandler(address, funcName, argsBuff...)
}
