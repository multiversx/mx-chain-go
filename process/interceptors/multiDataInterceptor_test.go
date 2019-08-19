package interceptors_test

import (
	"bytes"
	"errors"
	"sync/atomic"
	"testing"
	"time"

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

	assert.NotNil(t, mdi)
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

	err := mdi.ProcessReceivedMessage(nil)

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
		createMockThrottler(nil, nil),
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := mdi.ProcessReceivedMessage(msg)

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
		createMockThrottler(nil, nil),
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := mdi.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrNoTransactionInMessage, err)
}

func TestMultiDataInterceptor_ProcessReceivedCreateFailsShouldNotResend(t *testing.T) {
	t.Parallel()

	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttlerStartNum := int32(0)
	throttlerEndNum := int32(0)
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
		createMockThrottler(&throttlerStartNum, &throttlerEndNum),
	)
	mdi.SetBroadcastCallback(func(buffToSend []byte) {
		atomic.AddInt32(&broadcastNum, 1)
	})

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg)

	time.Sleep(time.Second)

	assert.Equal(t, errExpected, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(0), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerStartNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerEndNum))
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
	throttlerStartNum := int32(0)
	throttlerEndNum := int32(0)
	broadcastNum := int32(0)
	errExpected := errors.New("expected err")
	interceptedData := &mock.InterceptedDataStub{
		CheckValidCalled: func() error {
			return nil
		},
		IsAddressedToOtherShardCalled: func() bool {
			return false
		},
	}
	mdi, _ := interceptors.NewMultiDataInterceptor(
		marshalizer,
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				if bytes.Equal(buff, incorrectData) {
					return nil, errExpected
				} else {
					return interceptedData, nil
				}
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		createMockThrottler(&throttlerStartNum, &throttlerEndNum),
	)
	mdi.SetBroadcastCallback(func(buffToSend []byte) {
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
	})

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg)

	time.Sleep(time.Second)

	assert.Equal(t, errExpected, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerStartNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerEndNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&broadcastNum))
}

func TestMultiDataInterceptor_ProcessReceivedMessageIsAddressedToOtherShardShouldRetNilAndNotProcess(t *testing.T) {
	t.Parallel()

	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttlerStartNum := int32(0)
	throttlerEndNum := int32(0)
	interceptedData := &mock.InterceptedDataStub{
		CheckValidCalled: func() error {
			return nil
		},
		IsAddressedToOtherShardCalled: func() bool {
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
		createMockThrottler(&throttlerStartNum, &throttlerEndNum),
	)

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(0), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerStartNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerEndNum))
}

func TestMultiDataInterceptor_ProcessReceivedMessageOkMessageShouldRetNil(t *testing.T) {
	t.Parallel()

	buffData := [][]byte{[]byte("buff1"), []byte("buff2")}

	marshalizer := &mock.MarshalizerMock{}
	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttlerStartNum := int32(0)
	throttlerEndNum := int32(0)
	interceptedData := &mock.InterceptedDataStub{
		CheckValidCalled: func() error {
			return nil
		},
		IsAddressedToOtherShardCalled: func() bool {
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
		createMockThrottler(&throttlerStartNum, &throttlerEndNum),
	)

	dataField, _ := marshalizer.Marshal(buffData)
	msg := &mock.P2PMessageMock{
		DataField: dataField,
	}
	err := mdi.ProcessReceivedMessage(msg)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(2), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerStartNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerEndNum))
}
