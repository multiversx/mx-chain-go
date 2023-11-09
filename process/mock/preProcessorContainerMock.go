package mock

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
)

// PreProcessorContainerMock -
type PreProcessorContainerMock struct {
	GetCalled  func(key block.Type) (process.PreProcessor, error)
	KeysCalled func() []block.Type
}

// Get -
func (ppcm *PreProcessorContainerMock) Get(key block.Type) (process.PreProcessor, error) {
	if ppcm.GetCalled == nil {
		return &PreProcessorMock{}, nil
	}
	return ppcm.GetCalled(key)
}

// Add -
func (ppcm *PreProcessorContainerMock) Add(_ block.Type, _ process.PreProcessor) error {
	panic("implement me")
}

// AddMultiple -
func (ppcm *PreProcessorContainerMock) AddMultiple(_ []block.Type, _ []process.PreProcessor) error {
	panic("implement me")
}

// Replace -
func (ppcm *PreProcessorContainerMock) Replace(_ block.Type, _ process.PreProcessor) error {
	panic("implement me")
}

// Remove -
func (ppcm *PreProcessorContainerMock) Remove(_ block.Type) {
	panic("implement me")
}

// Len -
func (ppcm *PreProcessorContainerMock) Len() int {
	panic("implement me")
}

// Keys -
func (ppcm *PreProcessorContainerMock) Keys() []block.Type {
	if ppcm.KeysCalled == nil {
		return nil
	}

	return ppcm.KeysCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *PreProcessorContainerMock) IsInterfaceNil() bool {
	return ppcm == nil
}
