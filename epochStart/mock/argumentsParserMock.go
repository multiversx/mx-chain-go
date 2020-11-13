package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	vmcommon "github.com/ElrondNetwork/elrond-go/core/vm-common"
)

// ArgumentParserMock -
type ArgumentParserMock struct {
	ParseCallDataCalled               func(data string) (string, [][]byte, error)
	ParseDeployDataCalled             func(data string) (*parsers.DeployArgs, error)
	CreateDataFromStorageUpdateCalled func(storageUpdates []*vmcommon.StorageUpdate) string
	GetStorageUpdatesCalled           func(data string) ([]*vmcommon.StorageUpdate, error)
}

// ParseData -
func (ap *ArgumentParserMock) ParseCallData(data string) (string, [][]byte, error) {
	if ap.ParseCallDataCalled == nil {
		return "", nil, nil
	}
	return ap.ParseCallDataCalled(data)
}

// ParseData -
func (ap *ArgumentParserMock) ParseDeployData(data string) (*parsers.DeployArgs, error) {
	if ap.ParseDeployDataCalled == nil {
		return nil, nil
	}
	return ap.ParseDeployDataCalled(data)
}

// CreateDataFromStorageUpdate -
func (ap *ArgumentParserMock) CreateDataFromStorageUpdate(storageUpdates []*vmcommon.StorageUpdate) string {
	if ap.CreateDataFromStorageUpdateCalled == nil {
		return ""
	}
	return ap.CreateDataFromStorageUpdateCalled(storageUpdates)
}

// GetStorageUpdates -
func (ap *ArgumentParserMock) GetStorageUpdates(data string) ([]*vmcommon.StorageUpdate, error) {
	if ap.GetStorageUpdatesCalled == nil {
		return nil, nil
	}
	return ap.GetStorageUpdatesCalled(data)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ap *ArgumentParserMock) IsInterfaceNil() bool {
	return ap == nil
}
