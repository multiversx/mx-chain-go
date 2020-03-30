package mock

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ArgumentParserMock -
type ArgumentParserMock struct {
	ParseDataCalled                   func(data string) error
	GetConstructorArgumentsCalled     func() ([][]byte, error)
	GetFunctionArgumentsCalled        func() ([][]byte, error)
	GetCodeCalled                     func() ([]byte, error)
	GetCodeDecodedCalled              func() ([]byte, error)
	GetCodeMetadataCalled             func() (vmcommon.CodeMetadata, error)
	GetVMTypeCalled                   func() ([]byte, error)
	GetFunctionCalled                 func() (string, error)
	GetSeparatorCalled                func() string
	CreateDataFromStorageUpdateCalled func(storageUpdates []*vmcommon.StorageUpdate) string
	GetStorageUpdatesCalled           func(data string) ([]*vmcommon.StorageUpdate, error)
}

// ParseData -
func (ap *ArgumentParserMock) ParseData(data string) error {
	if ap.ParseDataCalled == nil {
		return nil
	}
	return ap.ParseDataCalled(data)
}

// GetConstructorArguments -
func (ap *ArgumentParserMock) GetConstructorArguments() ([][]byte, error) {
	if ap.GetConstructorArgumentsCalled == nil {
		return make([][]byte, 0), nil
	}
	return ap.GetConstructorArgumentsCalled()
}

// GetFunctionArguments -
func (ap *ArgumentParserMock) GetFunctionArguments() ([][]byte, error) {
	if ap.GetFunctionArgumentsCalled == nil {
		return make([][]byte, 0), nil
	}
	return ap.GetFunctionArgumentsCalled()
}

// GetCode -
func (ap *ArgumentParserMock) GetCode() ([]byte, error) {
	if ap.GetCodeCalled == nil {
		return []byte(""), nil
	}
	return ap.GetCodeCalled()
}

// GetCodeDecoded -
func (ap *ArgumentParserMock) GetCodeDecoded() ([]byte, error) {
	if ap.GetCodeDecodedCalled == nil {
		return []byte(""), nil
	}
	return ap.GetCodeDecodedCalled()
}

// GetCodeMetadata -
func (ap *ArgumentParserMock) GetCodeMetadata() (vmcommon.CodeMetadata, error) {
	if ap.GetCodeMetadataCalled == nil {
		return vmcommon.CodeMetadata{}, nil
	}
	return ap.GetCodeMetadataCalled()
}

// GetVMType -
func (ap *ArgumentParserMock) GetVMType() ([]byte, error) {
	if ap.GetVMTypeCalled == nil {
		return []byte("00"), nil
	}
	return ap.GetVMTypeCalled()
}

// GetFunction -
func (ap *ArgumentParserMock) GetFunction() (string, error) {
	if ap.GetFunctionCalled == nil {
		return "", nil
	}
	return ap.GetFunctionCalled()
}

// GetSeparator -
func (ap *ArgumentParserMock) GetSeparator() string {
	if ap.GetSeparatorCalled == nil {
		return "@"
	}
	return ap.GetSeparatorCalled()
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
