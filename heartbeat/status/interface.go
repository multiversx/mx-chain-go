package status

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
)

// HeartbeatMonitor defines the operations that a monitor should implement
type HeartbeatMonitor interface {
	GetHeartbeats() []data.PubKeyHeartbeat
	IsInterfaceNil() bool
}

// HeartbeatSenderInfoProvider is able to provide correct information about the current sender
type HeartbeatSenderInfoProvider interface {
	GetSenderInfo() (string, core.P2PPeerSubType, error)
	IsInterfaceNil() bool
}
