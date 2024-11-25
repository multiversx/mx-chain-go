package notifier

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	notifierProcess "github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	processMocks "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
)

func createArgs() ArgsNotifierBootstrapper {
	return ArgsNotifierBootstrapper{
		IncomingHeaderHandler: &sovereign.IncomingHeaderSubscriberStub{},
		SovereignNotifier:     &testscommon.SovereignNotifierStub{},
		ForkDetector:          &mock.ForkDetectorStub{},
		Bootstrapper:          &processMocks.BootstrapperStub{},
		RoundDuration:         100,
	}
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func TestNewNotifierBootstrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil incoming header processor", func(t *testing.T) {
		args := createArgs()
		args.IncomingHeaderHandler = nil
		nb, err := NewNotifierBootstrapper(args)
		require.Nil(t, nb)
		require.Equal(t, errorsMx.ErrNilIncomingHeaderSubscriber, err)
	})
	t.Run("nil sovereign notifier", func(t *testing.T) {
		args := createArgs()
		args.SovereignNotifier = nil
		nb, err := NewNotifierBootstrapper(args)
		require.Nil(t, nb)
		require.Equal(t, errNilSovereignNotifier, err)
	})
	t.Run("nil fork detector", func(t *testing.T) {
		args := createArgs()
		args.ForkDetector = nil
		nb, err := NewNotifierBootstrapper(args)
		require.Nil(t, nb)
		require.Equal(t, errorsMx.ErrNilForkDetector, err)
	})
	t.Run("nil bootstrapper", func(t *testing.T) {
		args := createArgs()
		args.Bootstrapper = nil
		nb, err := NewNotifierBootstrapper(args)
		require.Nil(t, nb)
		require.Equal(t, process.ErrNilBootstrapper, err)
	})
	t.Run("zero value round duration", func(t *testing.T) {
		args := createArgs()
		args.RoundDuration = 0
		nb, err := NewNotifierBootstrapper(args)
		require.Nil(t, nb)
		require.Equal(t, errorsMx.ErrInvalidRoundDuration, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createArgs()
		nb, err := NewNotifierBootstrapper(args)
		require.Nil(t, err)
		require.False(t, nb.IsInterfaceNil())
	})
}

func TestNotifierBootstrapper_Start(t *testing.T) {
	t.Parallel()

	args := createArgs()

	wasRegisteredToStateSync := false
	args.Bootstrapper = &processMocks.BootstrapperStub{
		AddSyncStateListenerCalled: func(f func(bool)) {
			require.Contains(t, getFunctionName(f), "(*notifierBootstrapper).receivedSyncState")
			wasRegisteredToStateSync = true
		},
	}

	registerCalledCt := atomic.Int64{}
	args.SovereignNotifier = &testscommon.SovereignNotifierStub{
		RegisterHandlerCalled: func(handler notifierProcess.IncomingHeaderSubscriber) error {
			require.Equal(t, args.IncomingHeaderHandler, handler)
			registerCalledCt.Add(1)
			return nil
		},
	}

	getHighestNonceCalledCt := atomic.Int64{}
	args.ForkDetector = &mock.ForkDetectorStub{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			defer func() {
				getHighestNonceCalledCt.Add(1)
			}()

			return uint64(getHighestNonceCalledCt.Load())
		},
	}

	nb, _ := NewNotifierBootstrapper(args)
	require.True(t, wasRegisteredToStateSync)

	nb.Start()

	defer func() {
		err := nb.Close()
		require.Nil(t, err)
	}()

	time.Sleep(time.Millisecond * 50)
	require.Zero(t, registerCalledCt.Load())
	require.Zero(t, getHighestNonceCalledCt.Load())

	nb.receivedSyncState(false)
	time.Sleep(time.Millisecond * 50)
	require.Zero(t, registerCalledCt.Load())
	require.Zero(t, getHighestNonceCalledCt.Load())

	nb.receivedSyncState(true)
	time.Sleep(time.Millisecond * 50)
	require.Zero(t, registerCalledCt.Load())
	require.Equal(t, int64(1), getHighestNonceCalledCt.Load())

	nb.receivedSyncState(true)
	time.Sleep(time.Millisecond * 50)
	require.Zero(t, registerCalledCt.Load())
	require.Equal(t, int64(2), getHighestNonceCalledCt.Load())

	for i := int64(0); i < 10; i++ {
		nb.receivedSyncState(false)
		time.Sleep(time.Millisecond * 50)
		require.Zero(t, registerCalledCt.Load())
		require.Equal(t, int64(2), getHighestNonceCalledCt.Load())
	}
	for i := int64(1); i < roundsThreshold; i++ {
		nb.receivedSyncState(true)
		time.Sleep(time.Millisecond * 50)
		require.Zero(t, registerCalledCt.Load())
		require.Equal(t, i+2, getHighestNonceCalledCt.Load())
	}

	for i := roundsThreshold; i < roundsThreshold+10; i++ {
		nb.receivedSyncState(true)
		time.Sleep(time.Millisecond * 50)
		require.Equal(t, int64(1), registerCalledCt.Load())
		require.Equal(t, int64(i+2), getHighestNonceCalledCt.Load())
	}
}

func TestNotifierBootstrapper_Start_ConcurrencyTest(t *testing.T) {
	t.Parallel()

	args := createArgs()

	getHighestNonceCalledCt := atomic.Int64{}
	args.ForkDetector = &mock.ForkDetectorStub{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			defer func() {
				getHighestNonceCalledCt.Add(1)
			}()

			return uint64(getHighestNonceCalledCt.Load())
		},
	}

	nb, _ := NewNotifierBootstrapper(args)
	nb.Start()

	defer func() {
		err := nb.Close()
		require.Nil(t, err)
	}()

	numGoRoutines := 1000
	wg := sync.WaitGroup{}
	wg.Add(numGoRoutines)

	for i := 0; i < numGoRoutines; i++ {
		go func(idx int) {
			isSynced := false
			if rand.Int31n(100) < 51 {
				isSynced = true
			}

			nb.receivedSyncState(isSynced)
			wg.Done()
		}(i)
	}

	wg.Wait()
	require.NotZero(t, getHighestNonceCalledCt.Load())
	require.True(t, getHighestNonceCalledCt.Load() < int64(numGoRoutines))
}

func TestNotifierBootstrapper_StartWithRegisterFailing(t *testing.T) {
	t.Parallel()

	sigStopNodeMock := make(chan os.Signal, 1)

	args := createArgs()
	args.SigStopNode = sigStopNodeMock
	args.RoundDuration = 10

	registerCalledCt := atomic.Int64{}
	args.SovereignNotifier = &testscommon.SovereignNotifierStub{
		RegisterHandlerCalled: func(handler notifierProcess.IncomingHeaderSubscriber) error {
			require.Equal(t, args.IncomingHeaderHandler, handler)

			defer func() {
				registerCalledCt.Add(1)
			}()

			return errors.New("local error")
		},
	}

	args.ForkDetector = &mock.ForkDetectorStub{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			return 1
		},
	}

	nb, _ := NewNotifierBootstrapper(args)
	nb.syncedRoundsChan <- roundsThreshold - 1

	nb.Start()

	defer func() {
		err := nb.Close()
		require.Nil(t, err)
	}()

	time.Sleep(time.Millisecond * 200)
	require.Zero(t, registerCalledCt.Load())

	nb.receivedSyncState(true)
	time.Sleep(time.Millisecond * 50)
	require.Equal(t, int64(1), registerCalledCt.Load())

	select {
	case sig := <-sigStopNodeMock:
		require.Equal(t, syscall.SIGTERM, sig)
	case <-time.After(time.Millisecond * 100): // Timeout to avoid hanging
		t.Error("expected SIGTERM signal on sigStopNodeMock, but none received")
	}

	// Once registration fails, the waiting is done, no other register is called
	for i := 0; i < 10; i++ {
		nb.receivedSyncState(true)
		time.Sleep(time.Millisecond * 50)
		require.Equal(t, int64(1), registerCalledCt.Load())
	}
}

func TestCheckNodeState_CtxDone(t *testing.T) {
	t.Parallel()

	args := createArgs()
	nb, _ := NewNotifierBootstrapper(args)

	ctx, cancel := context.WithCancel(context.Background())
	doneChan := make(chan struct{})

	go func() {
		nb.checkNodeState(ctx)
		close(doneChan)
	}()

	cancel()

	select {
	case <-doneChan:
	case <-time.After(time.Second):
		require.Fail(t, "checkNodeState did not exit on ctx.Done() as expected")
	}
}
