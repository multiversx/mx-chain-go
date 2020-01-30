package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type BlockNotifierHandlerMock struct {
	CallHandlersCalled    func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	RegisterHandlerCalled func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
}

func (bnhm *BlockNotifierHandlerMock) CallHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	if bnhm.CallHandlersCalled != nil {
		bnhm.CallHandlersCalled(shardID, headers, headersHashes)
	}
}

func (bnhm *BlockNotifierHandlerMock) RegisterHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bnhm.RegisterHandlerCalled != nil {
		bnhm.RegisterHandlerCalled(handler)
	}
}

func (bnhm *BlockNotifierHandlerMock) IsInterfaceNil() bool {
	return bnhm == nil
}
