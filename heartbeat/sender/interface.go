package sender

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

type senderHandler interface {
	ExecutionReadyChannel() <-chan time.Time
	Execute()
	Close()
	IsInterfaceNil() bool
}

type peerAuthenticationSenderHandler interface {
	senderHandler
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
