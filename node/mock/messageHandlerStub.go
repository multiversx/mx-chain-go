package mock

import (
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessageHandlerStub -
type MessageHandlerStub struct {
	CreateHeartbeatFromP2PMessageCalled func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error)
}

// IsInterfaceNil -
func (mhs *MessageHandlerStub) IsInterfaceNil() bool {
	return false
}

// CreateHeartbeatFromP2pMessage -
func (mhs *MessageHandlerStub) CreateHeartbeatFromP2PMessage(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
	return mhs.CreateHeartbeatFromP2PMessageCalled(message)
}
