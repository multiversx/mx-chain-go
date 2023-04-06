package status

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
)

// HeartbeatMonitor defines the operations that a monitor should implement
type HeartbeatMonitor interface {
	GetHeartbeats() []data.PubKeyHeartbeat
	IsInterfaceNil() bool
}

// HeartbeatSenderInfoProvider is able to provide correct information about the current sender
type HeartbeatSenderInfoProvider interface {
	GetCurrentNodeType() (string, core.P2PPeerSubType, error)
	IsInterfaceNil() bool
}
