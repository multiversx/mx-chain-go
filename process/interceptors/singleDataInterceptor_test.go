package interceptors_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func createMockInterceptorStub(checkCalledNum *int32, processCalledNum *int32) process.InterceptorProcessor {
	return &mock.InterceptorProcessorStub{
		CheckValidForProcessingCalled: func(data process.InterceptedData) error {
			if checkCalledNum != nil {
				atomic.AddInt32(checkCalledNum, 1)
			}

			return nil
		},
		ProcessInteceptedDataCalled: func(data process.InterceptedData) error {
			if processCalledNum != nil {
				atomic.AddInt32(processCalledNum, 1)
			}

			return nil
		},
	}
}

func createMockThrottler(throttlerStartNum *int32, throttlerEndNum *int32) process.InterceptorThrottler {
	return &mock.InterceptorThrottlerStub{
		CanProcessMessageCalled: func() bool {
			return true
		},
		StartMessageProcessingCalled: func() {
			if throttlerStartNum != nil {
				atomic.AddInt32(throttlerStartNum, 1)
			}
		},
		EndMessageProcessingCalled: func() {
			if throttlerEndNum != nil {
				atomic.AddInt32(throttlerEndNum, 1)
			}
		},
	}
}

func TestNewSingleDataInterceptor_NilInterceptedDataFactoryShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		nil,
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilInterceptedDataFactory, err)
}

func TestNewSingleDataInterceptor_NilInterceptedDataProcessorShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		&mock.InterceptedDataFactoryStub{},
		nil,
		&mock.InterceptorThrottlerStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilInterceptedDataProcessor, err)
}

func TestNewSingleDataInterceptor_NilInterceptorThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		nil,
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilInterceptorThrottler, err)
}

func TestNewSingleDataInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		nil,
	)

	assert.Nil(t, sdi)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewSingleDataInterceptor(t *testing.T) {
	t.Parallel()

	sdi, err := interceptors.NewSingleDataInterceptor(
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.NotNil(t, sdi)
	assert.Nil(t, err)
}

//------- ProcessReceivedMessage

func TestSingleDataInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	sdi, _ := interceptors.NewSingleDataInterceptor(
		&mock.InterceptedDataFactoryStub{},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{},
		mock.NewOneShardCoordinatorMock(),
	)

	err := sdi.ProcessReceivedMessage(nil)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestSingleDataInterceptor_ProcessReceivedMessageFactoryCreationErrorShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	sdi, _ := interceptors.NewSingleDataInterceptor(
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return nil, errExpected
			},
		},
		&mock.InterceptorProcessorStub{},
		&mock.InterceptorThrottlerStub{
			CanProcessMessageCalled: func() bool {
				return true
			},
			StartMessageProcessingCalled: func() {},
		},
		mock.NewOneShardCoordinatorMock(),
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := sdi.ProcessReceivedMessage(msg)

	assert.Equal(t, errExpected, err)
}

func TestSingleDataInterceptor_ProcessReceivedMessageIsNotForCurrentShardShouldNotCallProcess(t *testing.T) {
	t.Parallel()

	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttlerStartNum := int32(0)
	throttlerEndNum := int32(0)
	interceptedData := &mock.InterceptedDataStub{
		CheckValidCalled: func() error {
			return nil
		},
		IsAddressedToOtherShardCalled: func(shardCoordinator sharding.Coordinator) bool {
			return true
		},
	}

	sdi, _ := interceptors.NewSingleDataInterceptor(
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		createMockThrottler(&throttlerStartNum, &throttlerEndNum),
		mock.NewOneShardCoordinatorMock(),
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := sdi.ProcessReceivedMessage(msg)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(0), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(0), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerStartNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerEndNum))
}

func TestSingleDataInterceptor_ProcessReceivedMessageShouldWork(t *testing.T) {
	t.Parallel()

	checkCalledNum := int32(0)
	processCalledNum := int32(0)
	throttlerStartNum := int32(0)
	throttlerEndNum := int32(0)
	interceptedData := &mock.InterceptedDataStub{
		CheckValidCalled: func() error {
			return nil
		},
		IsAddressedToOtherShardCalled: func(shardCoordinator sharding.Coordinator) bool {
			return false
		},
	}

	sdi, _ := interceptors.NewSingleDataInterceptor(
		&mock.InterceptedDataFactoryStub{
			CreateCalled: func(buff []byte) (data process.InterceptedData, e error) {
				return interceptedData, nil
			},
		},
		createMockInterceptorStub(&checkCalledNum, &processCalledNum),
		createMockThrottler(&throttlerStartNum, &throttlerEndNum),
		mock.NewOneShardCoordinatorMock(),
	)

	msg := &mock.P2PMessageMock{
		DataField: []byte("data to be processed"),
	}
	err := sdi.ProcessReceivedMessage(msg)

	time.Sleep(time.Second)

	assert.Nil(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&checkCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&processCalledNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerStartNum))
	assert.Equal(t, int32(1), atomic.LoadInt32(&throttlerEndNum))
}
