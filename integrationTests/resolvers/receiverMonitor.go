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

// NewReceiverMonitor -
func NewReceiverMonitor(tb testing.TB) *receiverMonitor {
	return &receiverMonitor{
		timeout:  defaultReceiverMonitorTimeout,
		tb:       tb,
		chanDone: make(chan struct{}, 1),
	}
}

// Done -
func (rm *receiverMonitor) Done() {
	select {
	case rm.chanDone <- struct{}{}:
	default:
	}
}

// WaitWithTimeout -
func (rm *receiverMonitor) WaitWithTimeout() {
	select {
	case <-rm.chanDone:
		return
	case <-time.After(rm.timeout):
		assert.Fail(rm.tb, "timout waiting for data")
	}
}
