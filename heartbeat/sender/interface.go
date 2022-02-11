package sender

import "time"

type senderHandler interface {
	ShouldExecute() <-chan time.Time
	Execute()
	Close()
}

type timerHandler interface {
	CreateNewTimer(duration time.Duration)
}
