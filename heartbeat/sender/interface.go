package sender

import "time"

type senderHandler interface {
	ShouldExecute() <-chan time.Time
	Execute()
	Close()
	IsInterfaceNil() bool
}

type timerHandler interface {
	CreateNewTimer(duration time.Duration)
	ShouldExecute() <-chan time.Time
	Close()
}
