package mock

import (
	"github.com/ElrondNetwork/elrond-vm-common"
)

type ArgumentParserMock struct {
	ParseDataCalled                   func(data string) error
	GetArgumentsCalled                func() ([][]byte, error)
	GetCodeCalled                     func() ([]byte, error)
	GetFunctionCalled                 func() (string, error)
	GetSeparatorCalled                func() string
	CreateDataFromStorageUpdateCalled func(storageUpdates []*vmcommon.StorageUpdate) string
	GetStorageUpdatesCalled           func(data string) ([]*vmcommon.StorageUpdate, error)
}

func (ap *ArgumentParserMock) ParseData(data string) error {
	if ap.ParseDataCalled == nil {
		return nil
	}
	return ap.ParseDataCalled(data)
}

func (ap *ArgumentParserMock) GetArguments() ([][]byte, error) {
	if ap.GetArgumentsCalled == nil {
		return make([][]byte, 0), nil
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

func (ap *ArgumentParserMock) CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) string {
	if ap.CreateDataFromStorageUpdateCalled == nil {
		return ""
	}
	return ap.CreateDataFromStorageUpdateCalled(storageUpdates)
}

func (ap *ArgumentParserMock) GetStorageUpdates(data string) ([]*vmcommon.StorageUpdate, error) {
	if ap.GetStorageUpdatesCalled == nil {
		return nil, nil
	}
	return ap.GetStorageUpdatesCalled(data)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ap *ArgumentParserMock) IsInterfaceNil() bool {
	if ap == nil {
		return true
	}
	return false
}
