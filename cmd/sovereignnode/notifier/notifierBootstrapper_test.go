package notifier

import (
	"context"
	"errors"
	"os"
	"reflect"
	"runtime"
	"syscall"
	"testing"
	"time"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	processMocks "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	notifierProcess "github.com/multiversx/mx-chain-sovereign-notifier-go/process"
	"github.com/multiversx/mx-chain-sovereign-notifier-go/testscommon"
	"github.com/stretchr/testify/require"
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

	registerCalledCt := 0
	args.SovereignNotifier = &testscommon.SovereignNotifierStub{
		RegisterHandlerCalled: func(handler notifierProcess.IncomingHeaderSubscriber) error {
			require.Equal(t, args.IncomingHeaderHandler, handler)
			registerCalledCt++
			return nil
		},
	}

	getHighestNonceCalledCt := 0
	args.ForkDetector = &mock.ForkDetectorStub{
		GetHighestFinalBlockNonceCalled: func() uint64 {
			defer func() {
				getHighestNonceCalledCt++
			}()

			return uint64(getHighestNonceCalledCt)
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
	require.Zero(t, registerCalledCt)
	require.Zero(t, getHighestNonceCalledCt)

	nb.receivedSyncState(false)
	time.Sleep(time.Millisecond * 50)
	require.Zero(t, registerCalledCt)
	require.Zero(t, registerCalledCt)

	nb.receivedSyncState(true)
	time.Sleep(time.Millisecond * 50)
	require.Zero(t, registerCalledCt)
	require.Equal(t, 1, getHighestNonceCalledCt)

	nb.receivedSyncState(true)
	time.Sleep(time.Millisecond * 50)
	require.Equal(t, 1, registerCalledCt)
	require.Equal(t, 2, getHighestNonceCalledCt)

	for i := 3; i < 10; i++ {
		nb.receivedSyncState(true)
		time.Sleep(time.Millisecond * 50)
		require.Equal(t, 1, registerCalledCt)
		require.Equal(t, i, getHighestNonceCalledCt)
	}
}

func TestNotifierBootstrapper_StartWithRegisterFailing(t *testing.T) {
	t.Parallel()

	sigStopNodeMock := make(chan os.Signal, 1)

	args := createArgs()
	args.SigStopNode = sigStopNodeMock
	args.RoundDuration = 10

	registerCalledCt := 0
	args.SovereignNotifier = &testscommon.SovereignNotifierStub{
		RegisterHandlerCalled: func(handler notifierProcess.IncomingHeaderSubscriber) error {
			require.Equal(t, args.IncomingHeaderHandler, handler)

			defer func() {
				registerCalledCt++
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

	nb.Start()

	defer func() {
		err := nb.Close()
		require.Nil(t, err)
	}()

	time.Sleep(time.Millisecond * 200)
	require.Zero(t, registerCalledCt)

	nb.receivedSyncState(true)
	time.Sleep(time.Millisecond * 50)
	require.Equal(t, 1, registerCalledCt)

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
		require.Equal(t, 1, registerCalledCt)
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
