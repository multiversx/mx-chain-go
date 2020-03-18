package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// MiniBlockInterceptorStub -
type MiniBlockInterceptorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
	GetMiniBlockCalled           func(hash []byte, target int) (*block.MiniBlock, error)
}

// SetIsDataForCurrentShardVerifier -
func (m *MiniBlockInterceptorStub) SetIsDataForCurrentShardVerifier(_ process.InterceptedDataVerifier) error {
	return nil
}

// ProcessReceivedMessage -
func (m *MiniBlockInterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	if m.ProcessReceivedMessageCalled != nil {
		return m.ProcessReceivedMessageCalled(message, broadcastHandler)
	}

	return nil
}

// GetMiniBlock -
func (m *MiniBlockInterceptorStub) GetMiniBlock(hash []byte, target int) (*block.MiniBlock, error) {
	if m.GetMiniBlockCalled != nil {
		return m.GetMiniBlockCalled(hash, target)
	}

	return &block.MiniBlock{}, nil
}

// IsInterfaceNil -
func (m *MiniBlockInterceptorStub) IsInterfaceNil() bool {
	return m == nil
}
