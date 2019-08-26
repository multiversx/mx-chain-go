package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

type PreProcessorContainerMock struct {
	GetCalled  func(key block.Type) (process.PreProcessor, error)
	KeysCalled func() []block.Type
}

func (ppcm *PreProcessorContainerMock) Get(key block.Type) (process.PreProcessor, error) {
	if ppcm.GetCalled == nil {
		return &PreProcessorMock{}, nil
	}
	return ppcm.GetCalled(key)
}

func (ppcm *PreProcessorContainerMock) Add(key block.Type, val process.PreProcessor) error {
	panic("implement me")
}

func (ppcm *PreProcessorContainerMock) AddMultiple(keys []block.Type, preprocessors []process.PreProcessor) error {
	panic("implement me")
}

func (ppcm *PreProcessorContainerMock) Replace(key block.Type, val process.PreProcessor) error {
	panic("implement me")
}

func (ppcm *PreProcessorContainerMock) Remove(key block.Type) {
	panic("implement me")
}

func (ppcm *PreProcessorContainerMock) Len() int {
	panic("implement me")
}

func (ppcm *PreProcessorContainerMock) Keys() []block.Type {
	if ppcm.KeysCalled == nil {
		return nil
	}

	return ppcm.KeysCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *PreProcessorContainerMock) IsInterfaceNil() bool {
	if ppcm == nil {
		return true
	}
	return false
}
