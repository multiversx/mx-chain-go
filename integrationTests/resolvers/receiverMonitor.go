package resolvers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var defaultReceiverMonitorTimeout = time.Second * 2

type receiverMonitor struct {
	timeout  time.Duration
	tb       testing.TB
	chanDone chan struct{}
}

func newReceiverMonitor(tb testing.TB) *receiverMonitor {
	return &receiverMonitor{
		timeout:  defaultReceiverMonitorTimeout,
		tb:       tb,
		chanDone: make(chan struct{}, 1),
	}
}

func (rm *receiverMonitor) done() {
	select {
	case rm.chanDone <- struct{}{}:
	default:
	}
}

func (rm *receiverMonitor) waitWithTimeout() {
	select {
	case <-rm.chanDone:
		return
	case <-time.After(rm.timeout):
		assert.Fail(rm.tb, "timout waiting for data")
	}
}
