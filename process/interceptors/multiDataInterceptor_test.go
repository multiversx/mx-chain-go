package interceptors_test

import (
	"bytes"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fromConnectedPeerId = core.PeerID("from connected peer Id")
var testTopic = "test topic"

func TestNewMultiDataInterceptor_EmptyTopicShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		"",
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrEmptyTopic, err)
}

func TestNewMultiDataInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		testTopic,
		nil,
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewMultiDataInterceptor_NilInterceptedDataFactoryShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		nil,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilInterceptedDataFactory, err)
}

func TestNewMultiDataInterceptor_NilInterceptedDataProcessorShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		nil,
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilInterceptedDataProcessor, err)
}

func TestNewMultiDataInterceptor_NilInterceptorThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		nil,
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilInterceptorThrottler, err)
}

func TestNewMultiDataInterceptor_NilAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		nil,
		&mock.WhiteListHandlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilAntifloodHandler, err)
}

func TestNewMultiDataInterceptor_NilWhiteListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		nil,
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilWhiteListHandler, err)
}

func TestNewMultiDataInterceptor(t *testing.T) {
	t.Parallel()

	factory := &mock.InterceptedDataFactoryStub{}
	mdi, err := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		factory,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	require.False(t, check.IfNil(mdi))
	require.Nil(t, err)
	assert.Equal(t, testTopic, mdi.Topic())
}

//------- ProcessReceivedMessage

func TestMultiDataInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	err := mdi.ProcessReceivedMessage(nil, fromConnectedPeerId)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestMultiDataInterceptor_ProcessReceivedMessageUnmarshalFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpeced := errors.New("expected error")
	originatorPid := core.PeerID("originator")
	originatorBlackListed := false
	fromConnectedPeerBlackListed := false
	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errExpeced
			},
		},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		createMockThrottler(),
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
	err := mdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	assert.Equal(t, errExpeced, err)
	assert.True(t, originatorBlackListed)
	assert.True(t, fromConnectedPeerBlackListed)
}

func TestMultiDataInterceptor_ProcessReceivedMessageUnmarshalReturnsEmptySliceShouldErr(t *testing.T) {
	t.Parallel()

	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
		},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		createMockThrottler(),
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := mdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	assert.Equal(t, process.ErrNoDataInMessage, err)
}

func TestMultiDataInterceptor_ProcessReceivedCreateFailsShouldErr(t *testing.T) {
	t.Parallel()

	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttler := createMockThrottler()
	errExpected := errors.New("expected err")
	originatorPid := core.PeerID("originator")
	originatorBlackListed := false
	fromConnectedPeerBlackListed := false
	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		marshalizer,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return nil, errExpected
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
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

	dataField, _ := marshalizer.Marshal(&batch.Batch{Data: buffData})
	msg := &mock.P2PMessageMock{
		DataField: dataField,
		PeerField: originatorPid,
	}
	err := mdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	time.Sleep(time.Second)

	assert.Equal(t, errExpected, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(0), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
	assert.True(t, originatorBlackListed)
	assert.True(t, fromConnectedPeerBlackListed)
}

func TestMultiDataInterceptor_ProcessReceivedPartiallyCorrectDataShouldErr(t *testing.T) {
	t.Parallel()

	correctData := []byte("buff1")
	incorrectData := []byte("buff2")
	buffData := [][]byte{incorrectData, correctData}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttler := createMockThrottler()
	errExpected := errors.New("expected err")
	interceptedData := &mock.InterceptedDataStub{
		CheckValidityCalled: func() error {
			return nil
		},
		IsForCurrentShardCalled: func() bool {
			return true
		},
	}
	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		marshalizer,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				if bytes.Equal(buff, incorrectData) {
					return nil, errExpected
				}

				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	dataField, _ := marshalizer.Marshal(&batch.Batch{Data: buffData})
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	time.Sleep(time.Second)

	assert.Equal(t, errExpected, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(0), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
}

func TestMultiDataInterceptor_ProcessReceivedMessageNotValidShouldErrAndNotProcess(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	testProcessReceiveMessageMultiData(t, true, errExpected, 0)
}

func TestMultiDataInterceptor_ProcessReceivedMessageIsAddressedToOtherShardShouldRetNilAndNotProcess(t *testing.T) {
	t.Parallel()

	testProcessReceiveMessageMultiData(t, false, process.ErrInterceptedDataNotForCurrentShard, 0)
}

func TestMultiDataInterceptor_ProcessReceivedMessageOkMessageShouldRetNil(t *testing.T) {
	t.Parallel()

	testProcessReceiveMessageMultiData(t, true, nil, 2)
}

func testProcessReceiveMessageMultiData(t *testing.T, isForCurrentShard bool, expectedErr error, calledNum int) {
	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttler := createMockThrottler()
	interceptedData := &mock.InterceptedDataStub{
		CheckValidityCalled: func() error {
			return expectedErr
		},
		IsForCurrentShardCalled: func() bool {
			return isForCurrentShard
		},
	}
	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		marshalizer,
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

	dataField, _ := marshalizer.Marshal(&batch.Batch{Data: buffData})
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	time.Sleep(time.Second)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, int32(calledNum), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(calledNum), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
}

func TestMultiDataInterceptor_ProcessReceivedMessageWhitelistedShouldRetNil(t *testing.T) {
	t.Parallel()

	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
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
	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		marshalizer,
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

	dataField, _ := marshalizer.Marshal(&batch.Batch{Data: buffData})
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(2), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
}

//------- debug

func TestMultiDataInterceptor_SetInterceptedDebugHandlerNilShouldErr(t *testing.T) {
	t.Parallel()

	factory := &mock.InterceptedDataFactoryStub{}
	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		factory,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	err := mdi.SetInterceptedDebugHandler(nil)

	assert.Equal(t, process.ErrNilDebugger, err)
}

func TestMultiDataInterceptor_SetInterceptedDebugHandlerShouldWork(t *testing.T) {
	t.Parallel()

	factory := &mock.InterceptedDataFactoryStub{}
	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		&mock.MarshalizerMock{},
		factory,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		&mock.P2PAntifloodHandlerStub{},
		&mock.WhiteListHandlerStub{},
	)

	debugger := &mock.InterceptedDebugHandlerStub{}
	err := mdi.SetInterceptedDebugHandler(debugger)

	assert.Nil(t, err)
	assert.True(t, debugger == mdi.InterceptedDebugHandler()) //pointer testing
}

func TestMultiDataInterceptor_ProcessReceivedMessageIsOriginatorNotOkButWhiteListed(t *testing.T) {
	t.Parallel()

	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
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

	whiteListHandler := &mock.WhiteListHandlerStub{
		IsWhiteListedCalled: func(interceptedData process.InterceptedData) bool {
			return true
		},
	}
	errOriginator := process.ErrOnlyValidatorsCanUseThisTopic
	mdi, _ := interceptors.NewMultiDataInterceptor(
		testTopic,
		marshalizer,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
		&mock.P2PAntifloodHandlerStub{
			IsOriginatorEligibleForTopicCalled: func(pid core.PeerID, topic string) error {
				return errOriginator
			},
		},
		whiteListHandler,
	)

	dataField, _ := marshalizer.Marshal(&batch.Batch{Data: buffData})
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, fromConnectedPeerId)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(2), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())

	whiteListHandler.IsWhiteListedCalled = func(interceptedData process.InterceptedData) bool {
		return false
	}
	err = mdi.ProcessReceivedMessage(msg, fromConnectedPeerId)
	time.Sleep(time.Second)

	assert.Equal(t, err, errOriginator)
	assert.Equal(t, int32(2), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(2), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(2), throttler.StartProcessingCount())
	assert.Equal(t, int32(2), throttler.EndProcessingCount())
}

//------- IsInterfaceNil

func TestMultiDataInterceptor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var mdi *interceptors.MultiDataInterceptor

	assert.True(t, check.IfNil(mdi))
}
