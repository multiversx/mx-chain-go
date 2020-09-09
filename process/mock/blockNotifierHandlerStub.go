package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// BlockNotifierHandlerStub -
type BlockNotifierHandlerStub struct {
	CallHandlersCalled             func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	RegisterHandlerCalled          func(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte))
	GetNumRegisteredHandlersCalled func() int
}

// CallHandlers -
func (bnhs *BlockNotifierHandlerStub) CallHandlers(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	if bnhs.CallHandlersCalled != nil {
		bnhs.CallHandlersCalled(shardID, headers, headersHashes)
	}
}

// RegisterHandler -
func (bnhs *BlockNotifierHandlerStub) RegisterHandler(handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)) {
	if bnhs.RegisterHandlerCalled != nil {
		bnhs.RegisterHandlerCalled(handler)
	}
}

// GetNumRegisteredHandlers -
func (bnhs *BlockNotifierHandlerStub) GetNumRegisteredHandlers() int {
	if bnhs.GetNumRegisteredHandlersCalled != nil {
		return bnhs.GetNumRegisteredHandlersCalled()
	}

	return 0
}

// IsInterfaceNil -
func (bnhs *BlockNotifierHandlerStub) IsInterfaceNil() bool {
	return bnhs == nil
}
