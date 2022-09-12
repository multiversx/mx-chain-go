package sender

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type senderHandler interface {
	ExecutionReadyChannel() <-chan time.Time
	Execute()
	Close()
	IsInterfaceNil() bool
}

type hardforkHandler interface {
	ShouldTriggerHardfork() <-chan struct{}
	Execute()
	Close()
}

type peerAuthenticationSenderHandler interface {
	senderHandler
	hardforkHandler
}

type heartbeatSenderHandler interface {
	senderHandler
	GetCurrentNodeType() (string, core.P2PPeerSubType, error)
}

type timerHandler interface {
	CreateNewTimer(duration time.Duration)
	ExecutionReadyChannel() <-chan time.Time
	Close()
}
