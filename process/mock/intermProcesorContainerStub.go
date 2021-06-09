package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

// IntermProcessorContainerStub -
type IntermProcessorContainerStub struct {
	GetCalled  func(key block.Type) (process.IntermediateTransactionHandler, error)
	KeysCalled func() []block.Type
}

// Get -
func (ipcm *IntermProcessorContainerStub) Get(key block.Type) (process.IntermediateTransactionHandler, error) {
	if ipcm.GetCalled == nil {
		return &IntermediateTransactionHandlerStub{}, nil
	}
	return ipcm.GetCalled(key)
}

// Add -
func (ipcm *IntermProcessorContainerStub) Add(_ block.Type, _ process.IntermediateTransactionHandler) error {
	panic("implement me")
}

// AddMultiple -
func (ipcm *IntermProcessorContainerStub) AddMultiple(_ []block.Type, _ []process.IntermediateTransactionHandler) error {
	panic("implement me")
}

// Replace -
func (ipcm *IntermProcessorContainerStub) Replace(_ block.Type, _ process.IntermediateTransactionHandler) error {
	panic("implement me")
}

// Remove -
func (ipcm *IntermProcessorContainerStub) Remove(_ block.Type) {
	panic("implement me")
}

// Len -
func (ipcm *IntermProcessorContainerStub) Len() int {
	panic("implement me")
}

// Keys -
func (ipcm *IntermProcessorContainerStub) Keys() []block.Type {
	if ipcm.KeysCalled == nil {
		return nil
	}

	return ipcm.KeysCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ipcm *IntermProcessorContainerStub) IsInterfaceNil() bool {
	return ipcm == nil
}
