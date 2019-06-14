package mock

import (
	"github.com/ElrondNetwork/elrond-vm-common"
	"math/big"
)

type ArgumentParserMock struct {
	ParseDataCalled                   func(data []byte) error
	GetArgumentsCalled                func() ([]*big.Int, error)
	GetCodeCalled                     func() ([]byte, error)
	GetFunctionCalled                 func() (string, error)
	GetSeparatorCalled                func() string
	CreateDataFromStorageUpdateCalled func(storageUpdates []*vmcommon.StorageUpdate) []byte
	GetStorageUpdatesCalled           func(data []byte) ([]*vmcommon.StorageUpdate, error)
}

func (ap *ArgumentParserMock) ParseData(data []byte) error {
	if ap.ParseDataCalled == nil {
		return nil
	}
	return ap.ParseDataCalled(data)
}

func (ap *ArgumentParserMock) GetArguments() ([]*big.Int, error) {
	if ap.GetArgumentsCalled == nil {
		return make([]*big.Int, 0), nil
	}
	return ap.GetArgumentsCalled()
}

func (ap *ArgumentParserMock) GetCode() ([]byte, error) {
	if ap.GetCodeCalled == nil {
		return []byte(""), nil
	}
	return ap.GetCodeCalled()
}

func (ap *ArgumentParserMock) GetFunction() (string, error) {
	if ap.GetFunctionCalled == nil {
		return "", nil
	}
	return ap.GetFunctionCalled()
}

func (ap *ArgumentParserMock) GetSeparator() string {
	if ap.GetSeparatorCalled == nil {
		return "@"
	}
	return ap.GetSeparatorCalled()
}

func (ap *ArgumentParserMock) CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) []byte {
	if ap.CreateDataFromStorageUpdateCalled == nil {
		return nil
	}
	return ap.CreateDataFromStorageUpdateCalled(storageUpdates)
}

func (ap *ArgumentParserMock) GetStorageUpdates(data []byte) ([]*vmcommon.StorageUpdate, error) {
	if ap.GetStorageUpdatesCalled == nil {
		return nil, nil
	}
	return ap.GetStorageUpdatesCalled(data)
}
