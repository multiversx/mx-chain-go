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

func (ipcm *InterimProcessorContainerMock) Add(key block.Type, val process.IntermediateTransactionHandler) error {
	panic("implement me")
}

func (ipcm *InterimProcessorContainerMock) AddMultiple(keys []block.Type, preprocessors []process.IntermediateTransactionHandler) error {
	panic("implement me")
}

func (ipcm *InterimProcessorContainerMock) Replace(key block.Type, val process.IntermediateTransactionHandler) error {
	panic("implement me")
}

func (ipcm *InterimProcessorContainerMock) Remove(key block.Type) {
	panic("implement me")
}

func (ipcm *InterimProcessorContainerMock) Len() int {
	panic("implement me")
}

func (ipcm *InterimProcessorContainerMock) Keys() []block.Type {
	if ipcm.KeysCalled == nil {
		return nil
	}

	return ipcm.KeysCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ipcm *InterimProcessorContainerMock) IsInterfaceNil() bool {
	if ipcm == nil {
		return true
	}
	return false
}
