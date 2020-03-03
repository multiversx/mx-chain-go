package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MetaBlockInterceptorStub -
type MetaBlockInterceptorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
	GetMetaBlockCalled           func(target int, epoch uint32) (*block.MetaBlock, error)
}

// ProcessReceivedMessage -
func (m *MetaBlockInterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	if m.ProcessReceivedMessageCalled != nil {
		return m.ProcessReceivedMessageCalled(message, broadcastHandler)
	}

	return nil
}

// GetMetaBlock -
func (m *MetaBlockInterceptorStub) GetMetaBlock(target int, epoch uint32) (*block.MetaBlock, error) {
	if m.GetMetaBlockCalled != nil {
		return m.GetMetaBlockCalled(target, epoch)
	}

	return &block.MetaBlock{}, nil
}

// IsInterfaceNil -
func (m *MetaBlockInterceptorStub) IsInterfaceNil() bool {
	return m == nil
}
