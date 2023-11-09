package mock

import (
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

// ArgumentParserMock -
type ArgumentParserMock struct {
	ParseCallDataCalled               func(data string) (string, [][]byte, error)
	ParseArgumentsCalled              func(data string) ([][]byte, error)
	ParseDeployDataCalled             func(data string) (*parsers.DeployArgs, error)
	CreateDataFromStorageUpdateCalled func(storageUpdates []*vmcommon.StorageUpdate) string
	GetStorageUpdatesCalled           func(data string) ([]*vmcommon.StorageUpdate, error)
}

// ParseCallData -
func (ap *ArgumentParserMock) ParseCallData(data string) (string, [][]byte, error) {
	if ap.ParseCallDataCalled == nil {
		return "", nil, nil
	}
	return ap.ParseCallDataCalled(data)
}

// ParseArguments -
func (ap *ArgumentParserMock) ParseArguments(data string) ([][]byte, error) {
	if ap.ParseArgumentsCalled == nil {
		return [][]byte{}, nil
	}
	return ap.ParseArgumentsCalled(data)
}

// ParseData -
func (ap *ArgumentParserMock) ParseData(data string) (string, [][]byte, error) {
	if ap.ParseCallDataCalled == nil {
		return "", nil, nil
	}
	return ap.ParseCallDataCalled(data)
}

// ParseDeployData -
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
