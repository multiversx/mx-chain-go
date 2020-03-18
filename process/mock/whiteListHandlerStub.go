package mock

import "github.com/ElrondNetwork/elrond-go/process"

type WhiteListHandlerStub struct {
	RemoveCalled            func(keys [][]byte)
	AddCalled               func(keys [][]byte)
	IsForCurrentShardCalled func(interceptedData process.InterceptedData) bool
}

func (w *WhiteListHandlerStub) IsForCurrentShard(interceptedData process.InterceptedData) bool {
	if w.IsForCurrentShardCalled != nil {
		return w.IsForCurrentShardCalled(interceptedData)
	}
	return true
}

func (w *WhiteListHandlerStub) Remove(keys [][]byte) {
	if w.RemoveCalled != nil {
		w.RemoveCalled(keys)
	}
}

func (w *WhiteListHandlerStub) Add(keys [][]byte) {
	if w.AddCalled != nil {
		w.AddCalled(keys)
	}
}

func (w *WhiteListHandlerStub) IsInterfaceNil() bool {
	return w == nil
}
