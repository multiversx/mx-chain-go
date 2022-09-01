package sender

import "time"

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

type timerHandler interface {
	CreateNewTimer(duration time.Duration)
	ExecutionReadyChannel() <-chan time.Time
	Close()
}
