package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

type InterimProcessorContainerMock struct {
	GetCalled  func(key block.Type) (process.IntermediateTransactionHandler, error)
	KeysCalled func() []block.Type
}

func (ipcm *InterimProcessorContainerMock) Get(key block.Type) (process.IntermediateTransactionHandler, error) {
	if ipcm.GetCalled == nil {
		return &IntermediateTransactionHandlerMock{}, nil
	}
	return ipcm.GetCalled(key)
}

func (ppcm *InterimProcessorContainerMock) Add(key block.Type, val process.IntermediateTransactionHandler) error {
	panic("implement me")
}

func (ppcm *InterimProcessorContainerMock) AddMultiple(keys []block.Type, preprocessors []process.IntermediateTransactionHandler) error {
	panic("implement me")
}

func (ppcm *InterimProcessorContainerMock) Replace(key block.Type, val process.IntermediateTransactionHandler) error {
	panic("implement me")
}

func (ppcm *InterimProcessorContainerMock) Remove(key block.Type) {
	panic("implement me")
}

func (ppcm *InterimProcessorContainerMock) Len() int {
	panic("implement me")
}

func (ppcm *InterimProcessorContainerMock) Keys() []block.Type {
	if ppcm.KeysCalled == nil {
		return nil
	}

	return ppcm.KeysCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *InterimProcessorContainerMock) IsInterfaceNil() bool {
	if ppcm == nil {
		return true
	}
	return false
}
