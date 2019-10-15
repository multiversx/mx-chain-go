package interceptors_test

import (
	"bytes"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewMultiDataInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		nil,
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewMultiDataInterceptor_NilInterceptedDataFactoryShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		&mock.MarshalizerMock{},
		nil,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilInterceptedDataFactory, err)
}

func TestNewMultiDataInterceptor_NilInterceptedDataProcessorShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		nil,
		&mock.InterceptorThrottlerStub{},
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilInterceptedDataProcessor, err)
}

func TestNewMultiDataInterceptor_NilInterceptorThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		nil,
	)

	assert.Nil(t, mdi)
	assert.Equal(t, process.ErrNilInterceptorThrottler, err)
}

func TestNewMultiDataInterceptor(t *testing.T) {
	t.Parallel()

	mdi, err := interceptors.NewMultiDataInterceptor(
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
	)

	assert.False(t, check.IfNil(mdi))
	assert.Nil(t, err)
}

//------- ProcessReceivedMessage

func TestMultiDataInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	mdi, _ := interceptors.NewMultiDataInterceptor(
		&mock.MarshalizerMock{},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
	)

	err := mdi.ProcessReceivedMessage(nil, nil)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestMultiDataInterceptor_ProcessReceivedMessageUnmarshalFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpeced := errors.New("expected error")
	mdi, _ := interceptors.NewMultiDataInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errExpeced
			},
		},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		createMockThrottler(),
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := mdi.ProcessReceivedMessage(msg, nil)

	assert.Equal(t, errExpeced, err)
}

func TestMultiDataInterceptor_ProcessReceivedMessageUnmarshalReturnsEmptySliceShouldErr(t *testing.T) {
	t.Parallel()

	mdi, _ := interceptors.NewMultiDataInterceptor(
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
		},
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		createMockThrottler(),
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := mdi.ProcessReceivedMessage(msg, nil)

	assert.Equal(t, process.ErrNoDataInMessage, err)
}

func TestMultiDataInterceptor_ProcessReceivedCreateFailsShouldNotResend(t *testing.T) {
	t.Parallel()

	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttler := createMockThrottler()
	broadcastNum := int32(0)
	errExpected := errors.New("expected err")
	mdi, _ := interceptors.NewMultiDataInterceptor(
		marshalizer,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return nil, errExpected
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
	)
	bradcastCallback := func(buffToSend []byte) {
		atomic.AddInt32(&broadcastNum, 1)
	}

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, bradcastCallback)

	time.Sleep(time.Second)

	assert.Equal(t, errExpected, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(0), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
	assert.Equal(t, int32(0), atomic.LoadInt32(&broadcastNum))
}

func TestMultiDataInterceptor_ProcessReceivedPartiallyCorrectDataShouldSendOnlyCorrectPart(t *testing.T) {
	t.Parallel()

	correctData := []byte("buff1")
	incorrectData := []byte("buff2")
	buffData := [][]byte{incorrectData, correctData}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttler := createMockThrottler()
	broadcastNum := int32(0)
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
	)
	bradcastCallback := func(buffToSend []byte) {
		unmarshalledBuffs := make([][]byte, 0)
		err := marshalizer.Unmarshal(&unmarshalledBuffs, buffToSend)
		if err != nil {
			return
		}
		if len(unmarshalledBuffs) == 0 {
			return
		}
		if !bytes.Equal(unmarshalledBuffs[0], correctData) {
			return
		}

		atomic.AddInt32(&broadcastNum, 1)
	}

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, bradcastCallback)

	time.Sleep(time.Second)

	assert.Equal(t, errExpected, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
	assert.Equal(t, int32(1), atomic.LoadInt32(&broadcastNum))
}

func TestMultiDataInterceptor_ProcessReceivedMessageNotValidShouldErrAndNotProcess(t *testing.T) {
	t.Parallel()

	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttler := createMockThrottler()
	errExpected := errors.New("expected err")
	interceptedData := &mock.InterceptedDataStub{
		CheckValidityCalled: func() error {
			return errExpected
		},
		IsForCurrentShardCalled: func() bool {
			return true
		},
	}
	mdi, _ := interceptors.NewMultiDataInterceptor(
		marshalizer,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
	)

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, nil)

	time.Sleep(time.Second)

	assert.Equal(t, errExpected, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(0), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
}

func TestMultiDataInterceptor_ProcessReceivedMessageIsAddressedToOtherShardShouldRetNilAndNotProcess(t *testing.T) {
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
		marshalizer,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
	)

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, nil)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(0), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
}

func TestMultiDataInterceptor_ProcessReceivedMessageOkMessageShouldRetNil(t *testing.T) {
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
			return true
		},
	}
	mdi, _ := interceptors.NewMultiDataInterceptor(
		marshalizer,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		throttler,
	)

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg, nil)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(2), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), throttler.StartProcessingCount())
	assert.Equal(t, int32(1), throttler.EndProcessingCount())
}

//------- IsInterfaceNil

func TestMultiDataInterceptor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var mdi *interceptors.MultiDataInterceptor

	assert.True(t, check.IfNil(mdi))
}
