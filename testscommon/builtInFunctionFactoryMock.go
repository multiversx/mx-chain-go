package testscommon

import (
	"github.com/multiversx/mx-chain-vm-common-go"
)

// BuiltInFunctionFactoryMock -
type BuiltInFunctionFactoryMock struct {
	ESDTGlobalSettingsHandlerCalled      func() vmcommon.ESDTGlobalSettingsHandler
	NFTStorageHandlerCalled              func() vmcommon.SimpleESDTNFTStorageHandler
	BuiltInFunctionContainerCalled       func() vmcommon.BuiltInFunctionContainer
	SetPayableHandlerCalled              func(handler vmcommon.PayableHandler) error
	CreateBuiltInFunctionContainerCalled func() error
}

// ESDTGlobalSettingsHandler -
func (b *BuiltInFunctionFactoryMock) ESDTGlobalSettingsHandler() vmcommon.ESDTGlobalSettingsHandler {
	if b.ESDTGlobalSettingsHandlerCalled != nil {
		return b.ESDTGlobalSettingsHandlerCalled()
	}
	return &ESDTGlobalSettingsHandlerStub{}
}

// NFTStorageHandler -
func (b *BuiltInFunctionFactoryMock) NFTStorageHandler() vmcommon.SimpleESDTNFTStorageHandler {
	if b.NFTStorageHandlerCalled != nil {
		return b.NFTStorageHandlerCalled()
	}
	return &SimpleNFTStorageHandlerStub{}
}

// BuiltInFunctionContainer -
func (b *BuiltInFunctionFactoryMock) BuiltInFunctionContainer() vmcommon.BuiltInFunctionContainer {
	if b.BuiltInFunctionContainerCalled != nil {
		return b.BuiltInFunctionContainerCalled()
	}
	return &BuiltInFunctionContainerStub{}
}

// SetPayableHandler -
func (b *BuiltInFunctionFactoryMock) SetPayableHandler(handler vmcommon.PayableHandler) error {
	if b.SetPayableHandlerCalled != nil {
		return b.SetPayableHandlerCalled(handler)
	}
	return nil
}

// CreateBuiltInFunctionContainer -
func (b *BuiltInFunctionFactoryMock) CreateBuiltInFunctionContainer() error {
	if b.CreateBuiltInFunctionContainerCalled != nil {
		return b.CreateBuiltInFunctionContainerCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *BuiltInFunctionFactoryMock) IsInterfaceNil() bool {
	return b == nil
}
