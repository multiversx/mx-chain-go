package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
)

// HeartbeatMessageHandlerStub -
type HeartbeatMessageHandlerStub struct {
	CreateHeartbeatFromP2PMessageCalled func(message core.MessageP2P) (*data.Heartbeat, error)
}

// CreateHeartbeatFromP2PMessage -
func (hbmh *HeartbeatMessageHandlerStub) CreateHeartbeatFromP2PMessage(message core.MessageP2P) (*data.Heartbeat, error) {
	if hbmh.CreateHeartbeatFromP2PMessageCalled != nil {
		return hbmh.CreateHeartbeatFromP2PMessageCalled(message)
	}

	return nil, nil
}

// IsInterfaceNil -
func (hbmh *HeartbeatMessageHandlerStub) IsInterfaceNil() bool {
	return hbmh == nil
}
