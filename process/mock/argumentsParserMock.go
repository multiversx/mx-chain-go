package mock

import (
	"github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type AtArgumentParserMock struct {
	ParseDataCalled                   func(data []byte) error
	GetArgumentsCalled                func() ([]*big.Int, error)
	GetCodeCalled                     func() ([]byte, error)
	GetFunctionCalled                 func() (string, error)
	GetSeparatorCalled                func() string
	CreateDataFromStorageUpdateCalled func(storageUpdates []*vmcommon.StorageUpdate) []byte
	GetStorageUpdatesCalled           func(data []byte) ([]*vmcommon.StorageUpdate, error)
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

func (at *AtArgumentParserMock) GetSeparator() string {
	if at.GetSeparatorCalled == nil {
		return "@"
	}
	return at.GetSeparatorCalled()
}

func (at *AtArgumentParserMock) CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) []byte {
	if at.CreateDataFromStorageUpdateCalled == nil {
		return nil
	}
	return at.CreateDataFromStorageUpdateCalled(storageUpdates)
}

func (at *AtArgumentParserMock) GetStorageUpdates(data []byte) ([]*vmcommon.StorageUpdate, error) {
	if at.GetStorageUpdatesCalled == nil {
		return nil, nil
	}
	return at.GetStorageUpdatesCalled(data)
}
