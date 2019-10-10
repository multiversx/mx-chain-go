package mock

import (
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type MessageHandlerStub struct {
	CreateHeartbeatFromP2pMessageCalled func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error)
}

func (mhs *MessageHandlerStub) IsInterfaceNil() bool {
	return false
}

func (mhs *MessageHandlerStub) CreateHeartbeatFromP2pMessage(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
	return mhs.CreateHeartbeatFromP2pMessageCalled(message)
}
