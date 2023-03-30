package mock

import (
	"strings"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("node/mock")

type applicationRunningTrigger struct {
	chanClose chan struct{}
}

// NewApplicationRunningTrigger -
func NewApplicationRunningTrigger() *applicationRunningTrigger {
	return &applicationRunningTrigger{
		chanClose: make(chan struct{}),
	}
}

// Write -
func (trigger *applicationRunningTrigger) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "application is now running") {
		log.Info("got signal, trying to gracefully close the node")
		close(trigger.chanClose)
	}

	return 0, nil
}

// ChanClose -
func (trigger *applicationRunningTrigger) ChanClose() chan struct{} {
	return trigger.chanClose
}
