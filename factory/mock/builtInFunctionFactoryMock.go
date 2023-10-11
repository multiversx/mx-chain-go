package mock

import (
	"github.com/multiversx/mx-chain-go/testscommon"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// BuiltInFunctionFactory will handle built-in functions and components
type BuiltInFunctionFactory interface {
	ESDTGlobalSettingsHandler() vmcommon.ESDTGlobalSettingsHandler
	NFTStorageHandler() vmcommon.SimpleESDTNFTStorageHandler
	BuiltInFunctionContainer() vmcommon.BuiltInFunctionContainer
	SetPayableHandler(handler vmcommon.PayableHandler) error
	CreateBuiltInFunctionContainer() error
	IsInterfaceNil() bool
}

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
	if b.ESDTGlobalSettingsHandlerCalled == nil {
		return &testscommon.ESDTGlobalSettingsHandlerStub{}
	}
	return b.ESDTGlobalSettingsHandlerCalled()
}

// NFTStorageHandler -
func (b *BuiltInFunctionFactoryMock) NFTStorageHandler() vmcommon.SimpleESDTNFTStorageHandler {
	if b.NFTStorageHandlerCalled == nil {
		return &testscommon.SimpleNFTStorageHandlerStub{}
	}
	return b.NFTStorageHandlerCalled()
}

// BuiltInFunctionContainer -
func (b *BuiltInFunctionFactoryMock) BuiltInFunctionContainer() vmcommon.BuiltInFunctionContainer {
	if b.BuiltInFunctionContainerCalled == nil {
		return &BuiltInFunctionContainerStub{}
	}
	return b.BuiltInFunctionContainerCalled()
}

// SetPayableHandler -
func (b *BuiltInFunctionFactoryMock) SetPayableHandler(handler vmcommon.PayableHandler) error {
	if b.SetPayableHandlerCalled == nil {
		return nil
	}
	return b.SetPayableHandlerCalled(handler)
}

// CreateBuiltInFunctionContainer -
func (b *BuiltInFunctionFactoryMock) CreateBuiltInFunctionContainer() error {
	if b.CreateBuiltInFunctionContainerCalled == nil {
		return nil
	}
	return b.CreateBuiltInFunctionContainerCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *BuiltInFunctionFactoryMock) IsInterfaceNil() bool {
	return b == nil
}
