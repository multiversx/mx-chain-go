package mock

import (
	"math/big"
)

type AtArgumentParserMock struct {
	ParseDataCalled    func(data []byte) error
	GetArgumentsCalled func() ([]*big.Int, error)
	GetCodeCalled      func() ([]byte, error)
	GetFunctionCalled  func() (string, error)
}

func (at *AtArgumentParserMock) ParseData(data []byte) error {
	if at.ParseDataCalled == nil {
		return nil
	}
	return at.ParseDataCalled(data)
}

func (at *AtArgumentParserMock) GetArguments() ([]*big.Int, error) {
	if at.GetArgumentsCalled == nil {
		return make([]*big.Int, 0), nil
	}
	return at.GetArgumentsCalled()
}

func (at *AtArgumentParserMock) GetCode() ([]byte, error) {
	if at.GetCodeCalled == nil {
		return []byte(""), nil
	}
	return at.GetCodeCalled()
}

func (at *AtArgumentParserMock) GetFunction() (string, error) {
	if at.GetFunctionCalled == nil {
		return "", nil
	}
	return at.GetFunctionCalled()
}
