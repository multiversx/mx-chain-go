package interceptors_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInterceptorStub(checkCalledNum *int32, processCalledNum *int32) process.InterceptorProcessor {
	return &mock.InterceptorProcessorStub{
		ValidateCalled: func(data process.InterceptedData) error {
			if checkCalledNum != nil {
				atomic.AddInt32(checkCalledNum, 1)
			}

			return nil
		},
		SaveCalled: func(data process.InterceptedData) error {
			if processCalledNum != nil {
				atomic.AddInt32(processCalledNum, 1)
			}

			return nil
		},
	}
}

func createMockThrottler() *mock.InterceptorThrottlerStub {
	return &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
}

func TestNewSingleDataInterceptor_EmptyTopicShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		"",
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrEmptyTopic, err)
}

func TestNewSingleDataInterceptor_NilInterceptedDataFactoryShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		testTopic,
		nil,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilInterceptedDataFactory, err)
}

func TestNewSingleDataInterceptor_NilInterceptedDataProcessorShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		testTopic,
		&mock.InterceptedDataFactoryStub{},
		nil,
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilInterceptedDataProcessor, err)
}

func TestNewSingleDataInterceptor_NilInterceptorThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		testTopic,
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		nil,
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilInterceptorThrottler, err)
}

func TestNewSingleDataInterceptor_NilP2PAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		testTopic,
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		nil,
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilAntifloodHandler, err)
}

func TestNewSingleDataInterceptor_NilWhiteListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		testTopic,
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		nil,
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilWhiteListHandler, err)
}

func TestNewSingleDataInterceptor(t *testing.T) {
	t.Parallel()

	factory := &mock.InterceptedDataFactoryStub{}
	sdi, err := interceptors.NewSingleDataInterceptor(
		testTopic,
		factory,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	require.False(t, check.IfNil(sdi))
	require.Nil(t, err)
	assert.Equal(t, testTopic, sdi.Topic())
}

//------- ProcessReceivedMessage

func TestSingleDataInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	sdi, _ := interceptors.NewSingleDataInterceptor(
		testTopic,
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	err := sdi.ProcessReceivedMessage(nil, fromConnectedPeerId)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestSingleDataInterceptor_ProcessReceivedMessageFactoryCreationErrorShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	originatorPid := core.PeerID("originator")
	originatorBlackListed := false
	fromConnectedPeerBlackListed := false
	sdi, _ := interceptors.NewSingleDataInterceptor(
		testTopic,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return nil, errExpected
			},
		},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{
			CanProcessCalled: func() bool {
				return true
			},
		},
		&mock.P2PAntifloodHandlerStub{
			BlacklistPeerCalled: func(peer core.PeerID, reason string, duration time.Duration) {
				if peer == originatorPid {
					originatorBlackListed = true
				}
				if peer == fromConnectedPeerId {
					fromConnectedPeerBlackListed = true
				}
			},
		},
		&mock.WhiteListHandlerStub{},
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
		PeerField: originatorPid,
	}
	err := sdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	assert.Equal(t, errExpected, err)
	assert.True(t, originatorBlackListed)
	assert.True(t, fromConnectedPeerBlackListed)
}

func TestSingleDataInterceptor_ProcessReceivedMessageIsNotValidShouldNotCallProcess(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	testProcessReceiveMessage(t, true, errExpected, 0)
}

func TestSingleDataInterceptor_ProcessReceivedMessageIsNotForCurrentShardShouldNotCallProcess(t *testing.T) {
	t.Parallel()

	testProcessReceiveMessage(t, false, nil, 0)
}

func TestSingleDataInterceptor_ProcessReceivedMessageShouldWork(t *testing.T) {
	t.Parallel()

	testProcessReceiveMessage(t, true, nil, 1)
}

func testProcessReceiveMessage(t *testing.T, isForCurrentShard bool, validityErr error, calledNum int) {
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttler := createMockThrottler()
	interceptedData := &mock.InterceptedDataStub{
		CheckValidityCalled: func() error {
			return validityErr
		},
		IsForCurrentShardCalled: func() bool {
			return isForCurrentShard
		},
	}

	sdi, _ := interceptors.NewSingleDataInterceptor(
		testTopic,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := sdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	time.Sleep(time.Second)

	assert.Equal(t, validityErr, err)
	assert.Equal(t, int32(calledNum), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(calledNum), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
}

func TestSingleDataInterceptor_ProcessReceivedMessageWhitelistedShouldWork(t *testing.T) {
	t.Parallel()

	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttler := createMockThrottler()
	interceptedData := &mock.InterceptedDataStub{
		CheckValidityCalled: func() error {
			return nil
		},
		IsForCurrentShardCalled: func() bool {
			return false
		},
	}

	sdi, _ := interceptors.NewSingleDataInterceptor(
		testTopic,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{
			IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
				return true
			},
		},
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := sdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
}

//------- debug

func TestSingleDataInterceptor_SetInterceptedDebugHandlerNilShouldErr(t *testing.T) {
	t.Parallel()

	factory := &mock.InterceptedDataFactoryStub{}
	sdi, _ := interceptors.NewSingleDataInterceptor(
		testTopic,
		factory,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	err := sdi.SetInterceptedDebugHandler(nil)

	assert.Equal(t, process.ErrNilDebugger, err)
}

func TestSingleDataInterceptor_SetInterceptedDebugHandlerShouldWork(t *testing.T) {
	t.Parallel()

	factory := &mock.InterceptedDataFactoryStub{}
	sdi, _ := interceptors.NewSingleDataInterceptor(
		testTopic,
		factory,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	debugger := &mock.InterceptedDebugHandlerStub{}
	err := sdi.SetInterceptedDebugHandler(debugger)

	assert.Nil(t, err)
	assert.True(t, debugger == sdi.InterceptedDebugHandler()) //pointer testing
}

//------- IsInterfaceNil

func TestSingleDataInterceptor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var sdi *interceptors.SingleDataInterceptor

	assert.True(t, check.IfNil(sdi))
}
