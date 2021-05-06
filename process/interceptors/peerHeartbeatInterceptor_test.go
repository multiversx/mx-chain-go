package interceptors_test

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgPeerHeartbeatInterceptor() interceptors.ArgPeerHeartbeatInterceptor {
	argSingle := interceptors.ArgSingleDataInterceptor{
		Topic:            "test topic",
		DataFactory:      &mock.InterceptedDataFactoryStub{},
		Processor:        &mock.InterceptorProcessorStub{},
		Throttler:        createMockThrottler(),
		AntifloodHandler: &mock.P2PAntifloodHandlerStub{},
		WhiteListRequest: &mock.WhiteListHandlerStub{},
		CurrentPeerId:    "pid",
	}

	return interceptors.ArgPeerHeartbeatInterceptor{
		ArgSingleDataInterceptor: argSingle,
		Marshalizer:              &mock.MarshalizerMock{},
		ValidatorChecker:         &mock.NodesCoordinatorMock{},
		ProcessingThrottler:      createMockThrottler(),
		HeartbeatProcessor:       &mock.PeerHeartbeatProcessorStub{},
	}
}

func TestNewPeerHeartbeatInterceptor_NilFactoryShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatInterceptor()
	arg.DataFactory = nil
	interceptor, err := interceptors.NewPeerHeartbeatInterceptor(arg)

	assert.True(t, check.IfNil(interceptor))
	assert.Equal(t, process.ErrNilInterceptedDataFactory, err)
}

func TestNewPeerHeartbeatInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatInterceptor()
	arg.Marshalizer = nil
	interceptor, err := interceptors.NewPeerHeartbeatInterceptor(arg)

	assert.True(t, check.IfNil(interceptor))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewPeerHeartbeatInterceptor_NilValidatorCheckerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatInterceptor()
	arg.ValidatorChecker = nil
	interceptor, err := interceptors.NewPeerHeartbeatInterceptor(arg)

	assert.True(t, check.IfNil(interceptor))
	assert.Equal(t, process.ErrNilValidatorChecker, err)
}

func TestNewPeerHeartbeatInterceptor_NilProcessingThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatInterceptor()
	arg.ProcessingThrottler = nil
	interceptor, err := interceptors.NewPeerHeartbeatInterceptor(arg)

	assert.True(t, check.IfNil(interceptor))
	assert.Equal(t, process.ErrNilProcessingThrottler, err)
}

func TestNewPeerHeartbeatInterceptor_NilHeartbeatProcessorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatInterceptor()
	arg.HeartbeatProcessor = nil
	interceptor, err := interceptors.NewPeerHeartbeatInterceptor(arg)

	assert.True(t, check.IfNil(interceptor))
	assert.Equal(t, process.ErrNilHeartbeatProcessor, err)
}

func TestNewPeerHeartbeatInterceptor_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatInterceptor()
	interceptor, err := interceptors.NewPeerHeartbeatInterceptor(arg)

	assert.False(t, check.IfNil(interceptor))
	assert.Nil(t, err)
}

func TestPeerHeartbeatInterceptor_ProcessReceivedMessagePreProcessFailShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatInterceptor()
	arg.Throttler = &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return false
		},
	}
	interceptor, _ := interceptors.NewPeerHeartbeatInterceptor(arg)

	msg := &mock.P2PMessageMock{
		FromField:  []byte("from"),
		DataField:  []byte("data"),
		TopicField: arg.Topic,
		PeerField:  "",
	}

	err := interceptor.ProcessReceivedMessage(msg, "")
	assert.Equal(t, process.ErrSystemBusy, err)
	checkThrottlerNumStartEndCalls(t, arg.Throttler, 0)
}

func TestPeerHeartbeatInterceptor_ProcessReceivedMessageNotAnInterceptedDataShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	arg := createMockArgPeerHeartbeatInterceptor()
	numBlackListedCalled := 0
	arg.DataFactory = &mock.InterceptedDataFactoryStub{
		CreateCalled: func(buff []byte) (process.InterceptedData, error) {
			return nil, expectedErr
		},
	}
	arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
		BlacklistPeerCalled: func(peer core.PeerID, reason string, duration time.Duration) {
			assert.True(t, strings.Contains(reason, "can not create object from received bytes"))
			numBlackListedCalled++
		},
	}
	interceptor, _ := interceptors.NewPeerHeartbeatInterceptor(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("buff1"), []byte("buff2")},
	}
	buffBatch, _ := arg.Marshalizer.Marshal(b)

	msg := &mock.P2PMessageMock{
		FromField:  []byte("from"),
		DataField:  buffBatch,
		TopicField: arg.Topic,
		PeerField:  "",
	}

	err := interceptor.ProcessReceivedMessage(msg, "")
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 2, numBlackListedCalled)
	checkThrottlerNumStartEndCalls(t, arg.Throttler, 1)
}

func TestPeerHeartbeatInterceptor_ProcessReceivedMessageNotAValidInterceptedDataShouldErr(t *testing.T) {
	t.Parallel()

	cause := "intercepted data is not of type process.InterceptedPeerInfo"
	arg := createMockArgPeerHeartbeatInterceptor()
	numBlackListedCalled := 0
	arg.DataFactory = &mock.InterceptedDataFactoryStub{
		CreateCalled: func(buff []byte) (process.InterceptedData, error) {
			return &mock.InterceptedDataStub{}, nil
		},
	}
	arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
		BlacklistPeerCalled: func(peer core.PeerID, reason string, duration time.Duration) {
			assert.True(t, strings.Contains(reason, cause))
			numBlackListedCalled++
		},
	}
	interceptor, _ := interceptors.NewPeerHeartbeatInterceptor(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("buff1"), []byte("buff2")},
	}
	buffBatch, _ := arg.Marshalizer.Marshal(b)

	msg := &mock.P2PMessageMock{
		FromField:  []byte("from"),
		DataField:  buffBatch,
		TopicField: arg.Topic,
		PeerField:  "",
	}

	err := interceptor.ProcessReceivedMessage(msg, "")
	require.NotNil(t, err)
	assert.Equal(t, cause, err.Error())
	assert.Equal(t, 2, numBlackListedCalled)
	checkThrottlerNumStartEndCalls(t, arg.Throttler, 1)
}

func TestPeerHeartbeatInterceptor_ProcessReceivedMessageSystemBusyForObserversShouldNotProcess(t *testing.T) {
	t.Parallel()

	processCalled := uint32(0)
	arg := createMockArgPeerHeartbeatInterceptor()
	arg.DataFactory = &mock.InterceptedDataFactoryStub{
		CreateCalled: func(buff []byte) (process.InterceptedData, error) {
			return &mock.InterceptedPeerHeartbeatStub{}, nil
		},
	}
	arg.HeartbeatProcessor = &mock.PeerHeartbeatProcessorStub{
		ProcessCalled: func(peerHeartbeat process.InterceptedPeerHeartbeat) error {
			atomic.AddUint32(&processCalled, 1)

			return nil
		},
	}
	arg.ValidatorChecker = &mock.NodesCoordinatorMock{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error) {
			return nil, 0, fmt.Errorf("provided peer is an observer")
		},
	}
	arg.ProcessingThrottler = &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return false //no room for observers
		},
	}

	interceptor, _ := interceptors.NewPeerHeartbeatInterceptor(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("buff1"), []byte("buff2")},
	}
	buffBatch, _ := arg.Marshalizer.Marshal(b)

	msg := &mock.P2PMessageMock{
		FromField:  []byte("from"),
		DataField:  buffBatch,
		TopicField: arg.Topic,
		PeerField:  "",
	}

	err := interceptor.ProcessReceivedMessage(msg, "")
	require.Nil(t, err)

	time.Sleep(time.Second) //allow the processing go routine to start (in case a bug appears in the production code)
	assert.Equal(t, uint32(0), atomic.LoadUint32(&processCalled))
	checkThrottlerNumStartEndCalls(t, arg.Throttler, 1)
	checkThrottlerNumStartEndCalls(t, arg.ProcessingThrottler, 0)
}

func TestPeerHeartbeatInterceptor_ProcessReceivedMessageSystemBusyForValidatorsShouldWork(t *testing.T) {
	t.Parallel()

	processCalled := uint32(0)
	arg := createMockArgPeerHeartbeatInterceptor()
	arg.DataFactory = &mock.InterceptedDataFactoryStub{
		CreateCalled: func(buff []byte) (process.InterceptedData, error) {
			return &mock.InterceptedPeerHeartbeatStub{}, nil
		},
	}
	arg.HeartbeatProcessor = &mock.PeerHeartbeatProcessorStub{
		ProcessCalled: func(peerHeartbeat process.InterceptedPeerHeartbeat) error {
			atomic.AddUint32(&processCalled, 1)

			return nil
		},
	}
	arg.ValidatorChecker = &mock.NodesCoordinatorMock{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator sharding.Validator, shardId uint32, err error) {
			return nil, 0, nil //is a validator
		},
	}
	arg.ProcessingThrottler = &mock.InterceptorThrottlerStub{
		CanProcessCalled: func() bool {
			return false //no room for observers
		},
	}

	interceptor, _ := interceptors.NewPeerHeartbeatInterceptor(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("buff1"), []byte("buff2")},
	}
	buffBatch, _ := arg.Marshalizer.Marshal(b)

	msg := &mock.P2PMessageMock{
		FromField:  []byte("from"),
		DataField:  buffBatch,
		TopicField: arg.Topic,
		PeerField:  "",
	}

	err := interceptor.ProcessReceivedMessage(msg, "")
	require.Nil(t, err)

	time.Sleep(time.Second) //allow the processing go routine to start and finish
	assert.Equal(t, uint32(2), atomic.LoadUint32(&processCalled))
	checkThrottlerNumStartEndCalls(t, arg.Throttler, 1)
	checkThrottlerNumStartEndCalls(t, arg.ProcessingThrottler, 2)
}
