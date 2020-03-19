package floodPreventers

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewQuotaFloodPreventer

func TestNewQuotaFloodPreventer_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(
		nil,
		[]QuotaStatusHandler{&mock.QuotaStatusHandlerStub{}},
		minMessages,
		minTotalSize,
		minMessages,
		minTotalSize,
	)

	assert.True(t, check.IfNil(qfp))
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewQuotaFloodPreventer_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(
		&mock.CacherStub{},
		[]QuotaStatusHandler{nil},
		minMessages,
		minTotalSize,
		minMessages,
		minTotalSize,
	)

	assert.True(t, check.IfNil(qfp))
	assert.Equal(t, process.ErrNilQuotaStatusHandler, err)
}

func TestNewQuotaFloodPreventer_LowerMinMessagesPerPeerShouldErr(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(
		&mock.CacherStub{},
		[]QuotaStatusHandler{&mock.QuotaStatusHandlerStub{}},
		minMessages-1,
		minTotalSize,
		minMessages,
		minTotalSize,
	)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_LowerMinSizePerPeerShouldErr(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(
		&mock.CacherStub{},
		[]QuotaStatusHandler{&mock.QuotaStatusHandlerStub{}},
		minMessages,
		minTotalSize-1,
		minMessages,
		minTotalSize,
	)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_LowerMinMessagesShouldErr(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(
		&mock.CacherStub{},
		[]QuotaStatusHandler{&mock.QuotaStatusHandlerStub{}},
		minMessages,
		minTotalSize,
		minMessages-1,
		minTotalSize,
	)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_LowerMinSizeShouldErr(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(
		&mock.CacherStub{},
		[]QuotaStatusHandler{&mock.QuotaStatusHandlerStub{}},
		minMessages,
		minTotalSize,
		minMessages,
		minTotalSize-1,
	)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_ShouldWork(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(
		&mock.CacherStub{},
		[]QuotaStatusHandler{&mock.QuotaStatusHandlerStub{}},
		minMessages,
		minTotalSize,
		minMessages,
		minTotalSize,
	)

	assert.False(t, check.IfNil(qfp))
	assert.Nil(t, err)
}

func TestNewQuotaFloodPreventer_NilListShouldWork(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(
		&mock.CacherStub{},
		nil,
		minMessages,
		minTotalSize,
		minMessages,
		minTotalSize,
	)

	assert.False(t, check.IfNil(qfp))
	assert.Nil(t, err)
}

//------- Accumulate

func TestNewQuotaFloodPreventer_AccumulateIdentifierNotPresentPutQuotaAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	size := uint64(minTotalSize * 5)
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				q, isQuota := value.(*quota)
				if !isQuota {
					return
				}
				if q.numReceivedMessages == 1 && q.sizeReceivedMessages == size {
					putWasCalled = true
				}

				return
			},
		},
		nil,
		minMessages*4,
		minTotalSize*10,
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Accumulate("identifier", size)

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_AccumulateNotQuotaSavedInCacheShouldPutQuotaAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	size := uint64(minTotalSize * 5)
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return "bad value", true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				q, isQuota := value.(*quota)
				if !isQuota {
					return
				}
				if q.numReceivedMessages == 1 && q.sizeReceivedMessages == size {
					putWasCalled = true
				}

				return
			},
		},
		nil,
		minMessages*4,
		minTotalSize*10,
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Accumulate("identifier", size)

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_AccumulateUnderMaxValuesShouldIncrementAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	existingSize := uint64(minTotalSize * 5)
	existingMessages := uint32(minMessages * 2)
	existingQuota := &quota{
		numReceivedMessages:  existingMessages,
		sizeReceivedMessages: existingSize,
	}
	size := uint64(minTotalSize * 2)
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return existingQuota, true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				q, isQuota := value.(*quota)
				if !isQuota {
					return
				}
				if q.numReceivedMessages == existingMessages+1 && q.sizeReceivedMessages == existingSize+size {
					putWasCalled = true
				}

				return
			},
		},
		nil,
		minMessages*4,
		minTotalSize*10,
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Accumulate("identifier", size)

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_AccumulateGlobalWithResetShouldWork(t *testing.T) {
	t.Parallel()

	numPutOperations := 0
	addedGlobalQuotaCalled := false
	existingSize := uint64(0)
	existingMessages := uint32(0)
	existingQuota := &quota{
		numReceivedMessages:  existingMessages,
		sizeReceivedMessages: existingSize,
	}
	identifier := "identifier"
	size := uint64(minTotalSize * 2)
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return existingQuota, true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				if string(key) == identifier {
					numPutOperations++
				}

				return
			},
			KeysCalled: func() [][]byte {
				return make([][]byte, 0)
			},
			ClearCalled: func() {},
		},
		[]QuotaStatusHandler{
			&mock.QuotaStatusHandlerStub{
				AddQuotaCalled: func(_ string, _ uint32, _ uint64, _ uint32, _ uint64) {},
				SetGlobalQuotaCalled: func(numReceived uint32, sizeReceived uint64, numProcessed uint32, sizeProcessed uint64) {
					addedGlobalQuotaCalled = true
					assert.Equal(t, uint32(2), numReceived)
					assert.Equal(t, size+size+1, sizeReceived)
					assert.Equal(t, uint32(2), numProcessed)
					assert.Equal(t, size+size+1, sizeProcessed)
				},
				ResetStatisticsCalled: func() {},
			},
		},
		minMessages*4,
		minTotalSize*10,
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.AccumulateGlobal(identifier, size)
	assert.True(t, ok)

	ok = qfp.AccumulateGlobal(identifier, size+1)
	assert.True(t, ok)

	qfp.Reset()

	assert.Equal(t, 2, numPutOperations)
	assert.True(t, addedGlobalQuotaCalled)
}

//------- Accumulate per peer

func TestNewQuotaFloodPreventer_AccumulateOverMaxPeerNumMessagesShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	existingMessages := uint32(minMessages + 11)
	existingSize := uint64(minTotalSize * 3)
	existingQuota := &quota{
		numReceivedMessages:  existingMessages,
		sizeReceivedMessages: existingSize,
	}
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return existingQuota, true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				assert.Fail(t, "should have not called put")

				return false
			},
		},
		nil,
		minMessages*4,
		minTotalSize*10,
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Accumulate("identifier", minTotalSize)

	assert.False(t, ok)
}

func TestNewQuotaFloodPreventer_AccumulateOverMaxPeerSizeShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	existingMessages := uint32(minMessages)
	existingSize := uint64(minTotalSize * 11)
	existingQuota := &quota{
		numReceivedMessages:  existingMessages,
		sizeReceivedMessages: existingSize,
	}
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return existingQuota, true
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				assert.Fail(t, "should have not called put")

				return false
			},
		},
		nil,
		minMessages*4,
		minTotalSize*10,
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Accumulate("identifier", minTotalSize)

	assert.False(t, ok)
}

//------- Accumulate globally

func TestNewQuotaFloodPreventer_AccumulateOverMaxNumMessagesShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	globalMessages := uint32(minMessages + 11)
	globalSize := uint64(minTotalSize * 3)
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				assert.Fail(t, "should have not called put")

				return false
			},
		},
		nil,
		minMessages*4,
		minTotalSize*10,
		minMessages*4,
		minTotalSize*10,
	)
	qfp.SetGlobalQuotaValues(globalMessages, globalSize)

	ok := qfp.Accumulate("identifier", minTotalSize)

	assert.False(t, ok)
}

func TestNewQuotaFloodPreventer_AccumulateOverMaxSizeShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	globalMessages := uint32(minMessages)
	globalSize := uint64(minTotalSize * 11)
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
			PutCalled: func(key []byte, value interface{}) (evicted bool) {
				assert.Fail(t, "should have not called put")

				return false
			},
		},
		nil,
		minMessages*4,
		minTotalSize*10,
		minMessages*4,
		minTotalSize*10,
	)
	qfp.SetGlobalQuotaValues(globalMessages, globalSize)

	ok := qfp.Accumulate("identifier", minTotalSize)

	assert.False(t, ok)
}

func TestCountersMap_AccumulateShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	numIterations := 1000
	qfp, _ := NewQuotaFloodPreventer(
		mock.NewCacherMock(),
		nil,
		minMessages,
		minTotalSize,
		minMessages*uint32(numIterations),
		minTotalSize*uint64(numIterations),
	)
	wg := sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			ok := qfp.Accumulate(fmt.Sprintf("%d", idx), minTotalSize)
			assert.True(t, ok)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

//------- Reset

func TestCountersMap_ResetShouldCallCacherClear(t *testing.T) {
	t.Parallel()

	clearCalled := false
	qfp, _ := NewQuotaFloodPreventer(
		&mock.CacherStub{
			ClearCalled: func() {
				clearCalled = true
			},
			KeysCalled: func() [][]byte {
				return make([][]byte, 0)
			},
		},
		nil,
		minTotalSize,
		minMessages,
		minTotalSize,
		minMessages,
	)

	qfp.Reset()

	assert.True(t, clearCalled)
}

func TestCountersMap_ResetShouldCallQuotaStatus(t *testing.T) {
	t.Parallel()

	cacher := mock.NewCacherMock()
	key1 := []byte("key1")
	quota1 := &quota{
		numReceivedMessages:   1,
		sizeReceivedMessages:  2,
		numProcessedMessages:  3,
		sizeProcessedMessages: 4,
	}
	key2 := []byte("key2")
	quota2 := &quota{
		numReceivedMessages:   5,
		sizeReceivedMessages:  6,
		numProcessedMessages:  7,
		sizeProcessedMessages: 8,
	}

	cacher.HasOrAdd(key1, quota1)
	cacher.HasOrAdd(key2, quota2)

	resetStatisticsCalled := false
	quota1Compared := false
	quota2Compared := false
	qfp, _ := NewQuotaFloodPreventer(
		cacher,
		[]QuotaStatusHandler{
			&mock.QuotaStatusHandlerStub{
				ResetStatisticsCalled: func() {
					resetStatisticsCalled = true
				},
				AddQuotaCalled: func(identifier string, numReceivedMessages uint32, sizeReceivedMessages uint64, numProcessedMessages uint32, sizeProcessedMessages uint64) {
					quotaProvided := quota{
						numReceivedMessages:   numReceivedMessages,
						sizeReceivedMessages:  sizeReceivedMessages,
						numProcessedMessages:  numProcessedMessages,
						sizeProcessedMessages: sizeProcessedMessages,
					}
					quotaToCompare := quota{}

					switch identifier {
					case string(key1):
						quotaToCompare = *quota1
						quota1Compared = true
					case string(key2):
						quotaToCompare = *quota2
						quota2Compared = true
					default:
						assert.Fail(t, fmt.Sprintf("unknown identifier %s", identifier))
					}

					assert.Equal(t, quotaToCompare, quotaProvided)
				},
				SetGlobalQuotaCalled: func(_ uint32, _ uint64, _ uint32, _ uint64) {},
			},
		},
		minTotalSize,
		minMessages,
		minTotalSize,
		minMessages,
	)

	qfp.Reset()

	assert.True(t, resetStatisticsCalled)
	assert.True(t, quota1Compared)
	assert.True(t, quota2Compared)
}

func TestCountersMap_IncrementAndResetShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	numIterations := 1000
	qfp, _ := NewQuotaFloodPreventer(
		mock.NewCacherMock(),
		nil,
		minMessages,
		minTotalSize,
		minTotalSize*uint32(numIterations),
		minMessages*uint64(numIterations),
	)
	wg := sync.WaitGroup{}
	wg.Add(numIterations + numIterations/10)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			ok := qfp.Accumulate(fmt.Sprintf("%d", idx), minTotalSize)
			assert.True(t, ok)
			wg.Done()
		}(i)

		if i%10 == 0 {
			go func() {
				qfp.Reset()
				wg.Done()
			}()
		}
	}

	wg.Wait()
}
