package mock

import "github.com/ElrondNetwork/elrond-go/process"

type WhiteListHandlerMock struct {
	RemoveCalled            func(keys [][]byte)
	AddCalled               func(keys [][]byte)
	IsForCurrentShardCalled func(interceptedData process.InterceptedData) bool
}

func (w *WhiteListHandlerMock) IsForCurrentShard(interceptedData process.InterceptedData) bool {
	if w.IsForCurrentShardCalled != nil {
		return w.IsForCurrentShardCalled(interceptedData)
	}
	return true
}

func (w *WhiteListHandlerMock) Remove(keys [][]byte) {
	if w.RemoveCalled != nil {
		w.RemoveCalled(keys)
	}
}

func (w *WhiteListHandlerMock) Add(keys [][]byte) {
	if w.AddCalled != nil {
		w.AddCalled(keys)
	}
}

func (w *WhiteListHandlerMock) IsInterfaceNil() bool {
	return w == nil
}
