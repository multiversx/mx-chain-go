package storageResolvers

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

const timeout = time.Second * 2

func TestStorageResolver_ImplementedMethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("%v", r))
		}
	}()

	sr := &storageResolver{}
	assert.Nil(t, sr.ProcessReceivedMessage(nil, ""))
	assert.Nil(t, sr.SetResolverDebugHandler(nil))
	sr.SetNumPeersToQuery(0, 0)
	v1, v2 := sr.NumPeersToQuery()
	assert.Equal(t, 0, v1)
	assert.Equal(t, 0, v2)
}

func TestStorageResolver_signalGracefullyClose(t *testing.T) {
	t.Parallel()

	chanClose := make(chan endProcess.ArgEndProcess, 1)
	timeToWait := time.Millisecond * 100
	sr := &storageResolver{
		chanGracefullyClose: chanClose,
		manualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{
			CurrentEpochCalled: func() uint32 {
				return 0
			},
		},
		delayBeforeGracefulClose: timeToWait,
	}

	startTime := time.Now()
	endTime := time.Now()
	sr.signalGracefullyClose()
	endArg := endProcess.ArgEndProcess{}

	select {
	case endArg = <-chanClose:
		endTime = time.Now()
	case <-time.After(timeout):
		assert.Fail(t, "timeout. Should have written on the output chan")
	}

	assert.True(t, timeToWait <= endTime.Sub(startTime))
	assert.Equal(t, endArg.Reason, core.ImportComplete)
}

func TestStorageResolver_signalGracefullyCloseCanNotWriteOnChanShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	chanClose := make(chan endProcess.ArgEndProcess, 0)
	sr := &storageResolver{
		chanGracefullyClose: chanClose,
		manualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{
			CurrentEpochCalled: func() uint32 {
				return 0
			},
		},
	}

	sr.signalGracefullyClose()

	time.Sleep(time.Second)
}

func TestStorageResolver_signalGracefullyCloseDoubleSignalShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic: %v", r))
		}
	}()

	chanClose := make(chan endProcess.ArgEndProcess, 0)
	sr := &storageResolver{
		chanGracefullyClose: chanClose,
		manualEpochStartNotifier: &mock.ManualEpochStartNotifierStub{
			CurrentEpochCalled: func() uint32 {
				return 0
			},
		},
	}

	sr.signalGracefullyClose()
	sr.signalGracefullyClose()

	time.Sleep(time.Second)
}
