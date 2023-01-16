package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
)

// InterimProcessorContainerMock -
type InterimProcessorContainerMock struct {
	GetCalled  func(key block.Type) (process.IntermediateTransactionHandler, error)
	KeysCalled func() []block.Type
}

// Get -
func (ipcm *InterimProcessorContainerMock) Get(key block.Type) (process.IntermediateTransactionHandler, error) {
	if ipcm.GetCalled == nil {
		return &IntermediateTransactionHandlerMock{}, nil
	}
	return ipcm.GetCalled(key)
}

// Add -
func (ipcm *InterimProcessorContainerMock) Add(_ block.Type, _ process.IntermediateTransactionHandler) error {
	panic("implement me")
}

// AddMultiple -
func (ipcm *InterimProcessorContainerMock) AddMultiple(_ []block.Type, _ []process.IntermediateTransactionHandler) error {
	panic("implement me")
}

// Replace -
func (ipcm *InterimProcessorContainerMock) Replace(_ block.Type, _ process.IntermediateTransactionHandler) error {
	panic("implement me")
}

// Remove -
func (ipcm *InterimProcessorContainerMock) Remove(_ block.Type) {
	panic("implement me")
}

// Len -
func (ipcm *InterimProcessorContainerMock) Len() int {
	panic("implement me")
}

// Keys -
func (ipcm *InterimProcessorContainerMock) Keys() []block.Type {
	if ipcm.KeysCalled == nil {
		return nil
	}

	return ipcm.KeysCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ipcm *InterimProcessorContainerMock) IsInterfaceNil() bool {
	return ipcm == nil
}
