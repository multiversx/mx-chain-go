package mock

import "errors"

var errNotImplemented = errors.New("not implemented")

type DataGetterStub struct {
	GetCalled func(scAddress []byte, funcName string, args ...[]byte) ([]byte, error)
}

func (dgs *DataGetterStub) Get(scAddress []byte, funcName string, args ...[]byte) ([]byte, error) {
	if dgs.GetCalled == nil {
		return nil, errNotImplemented
	}

	return dgs.GetCalled(scAddress, funcName, args...)
}
