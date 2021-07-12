package closing

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

const timeout = time.Second

func TestNewSafeChanCloser(t *testing.T) {
	t.Parallel()

	closer := NewSafeChanCloser()
	assert.False(t, check.IfNil(closer))
}

func TestSafeChanCloser_CloseShouldWork(t *testing.T) {
	t.Parallel()

	chDone := make(chan struct{}, 1)
	closer := NewSafeChanCloser()
	go func() {
		<-closer.ChanClose()
		chDone <- struct{}{}
	}()

	closer.Close()

	select {
	case <-chDone:
	case <-time.After(timeout):
		assert.Fail(t, "should have not timed out")
	}
}

func TestSafeChanCloser_CloseConcurrentShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer panicHandler(t)

	closer := NewSafeChanCloser()
	for i := 0; i < 10; i++ {
		go func() {
			defer panicHandler(t)
			time.Sleep(time.Millisecond * 100)
			closer.Close()
		}()
	}

	time.Sleep(timeout)
}

func panicHandler(t *testing.T) {
	r := recover()
	if r != nil {
		assert.Fail(t, fmt.Sprintf("should have not paniced: %v", r))
	}
}
