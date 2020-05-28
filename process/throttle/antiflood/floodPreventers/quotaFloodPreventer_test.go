package floodPreventers

import (
	"errors"
	"fmt"
	"math"
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
		10,
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
		10,
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
		10,
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
		10,
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
		10,
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
		10,
	)

	assert.False(t, check.IfNil(qfp))
	assert.Nil(t, err)
}

//------- IncreaseLoad

func TestNewQuotaFloodPreventer_IncreaseLoadIdentifierNotPresentPutQuotaAndReturnTrue(t *testing.T) {
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
		10,
	)

	err := qfp.IncreaseLoad("identifier", size)

	assert.Nil(t, err)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_IncreaseLoadNotQuotaSavedInCacheShouldPutQuotaAndReturnTrue(t *testing.T) {
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
		10,
	)

	err := qfp.IncreaseLoad("identifier", size)

	assert.Nil(t, err)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_IncreaseLoadUnderMaxValuesShouldIncrementAndReturnTrue(t *testing.T) {
	t.Parallel()

	existingSize := uint64(minTotalSize * 5)
	existingMessages := uint32(minMessages * 8)
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
		},
		nil,
		minMessages*10,
		minTotalSize*10,
		10,
	)

	err := qfp.IncreaseLoad("identifier", size)

	assert.Nil(t, err)
}

//------- IncreaseLoad per peer

func TestNewQuotaFloodPreventer_IncreaseLoadOverMaxPeerNumMessagesShouldNotPutAndReturnFalse(t *testing.T) {
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
		10,
	)

	err := qfp.IncreaseLoad("identifier", minTotalSize)

	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestNewQuotaFloodPreventer_IncreaseLoadOverMaxPeerSizeShouldNotPutAndReturnFalse(t *testing.T) {
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
		10,
	)

	err := qfp.IncreaseLoad("identifier", minTotalSize)

	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestCountersMap_IncreaseLoadShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	numIterations := 1000
	qfp, _ := NewQuotaFloodPreventer(
		mock.NewCacherMock(),
		nil,
		minMessages,
		minTotalSize,
		10,
	)
	wg := sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			err := qfp.IncreaseLoad(fmt.Sprintf("%d", idx), minTotalSize)
			assert.Nil(t, err)
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
		10,
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
			},
		},
		minTotalSize,
		minMessages,
		10,
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
		10,
	)
	wg := sync.WaitGroup{}
	wg.Add(numIterations + numIterations/10)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			err := qfp.IncreaseLoad(fmt.Sprintf("%d", idx), minTotalSize)
			assert.Nil(t, err)
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

func TestNewQuotaFloodPreventer_IncreaseLoadWithMockCacherShouldWork(t *testing.T) {
	t.Parallel()

	reserved := uint32(17)
	numMessages := uint32(100)
	qfp, _ := NewQuotaFloodPreventer(
		mock.NewCacherMock(),
		nil,
		numMessages,
		math.MaxUint64,
		reserved,
	)

	identifier := "id"
	for i := uint32(0); i < numMessages-reserved; i++ {
		err := qfp.IncreaseLoad(identifier, 1)
		assert.Nil(t, err, fmt.Sprintf("failed at the %d iteration", i))
	}

	for i := uint32(0); i < reserved*2; i++ {
		err := qfp.IncreaseLoad(identifier, 1)
		assert.True(t, errors.Is(err, process.ErrSystemBusy), fmt.Sprintf("failed at the %d iteration", numMessages-reserved+i))
	}
}
