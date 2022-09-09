package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
)

// MessageHandlerStub -
type MessageHandlerStub struct {
	CreateHeartbeatFromP2PMessageCalled func(message core.MessageP2P) (*data.Heartbeat, error)
}

// IsInterfaceNil -
func (mhs *MessageHandlerStub) IsInterfaceNil() bool {
	return false
}

// CreateHeartbeatFromP2PMessage -
func (mhs *MessageHandlerStub) CreateHeartbeatFromP2PMessage(message core.MessageP2P) (*data.Heartbeat, error) {
	if mhs.CreateHeartbeatFromP2PMessageCalled != nil {
		return mhs.CreateHeartbeatFromP2PMessageCalled(message)
	}

	return &data.Heartbeat{}, nil
}
