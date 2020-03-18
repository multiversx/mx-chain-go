package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// MetaBlockInterceptorStub -
type MetaBlockInterceptorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
	GetMetaBlockCalled           func(hash []byte, target int) (*block.MetaBlock, error)
}

// SetIsDataForCurrentShardVerifier -
func (m *MetaBlockInterceptorStub) SetIsDataForCurrentShardVerifier(_ process.InterceptedDataVerifier) error {
	return nil
}

// ProcessReceivedMessage -
func (m *MetaBlockInterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	if m.ProcessReceivedMessageCalled != nil {
		return m.ProcessReceivedMessageCalled(message, broadcastHandler)
	}

	return nil
}

// GetMetaBlock -
func (m *MetaBlockInterceptorStub) GetMetaBlock(hash []byte, target int) (*block.MetaBlock, error) {
	if m.GetMetaBlockCalled != nil {
		return m.GetMetaBlockCalled(hash, target)
	}

	return &block.MetaBlock{}, nil
}

// IsInterfaceNil -
func (m *MetaBlockInterceptorStub) IsInterfaceNil() bool {
	return m == nil
}
