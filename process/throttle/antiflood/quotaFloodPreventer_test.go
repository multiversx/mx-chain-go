package antiflood

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

	qfp, err := NewQuotaFloodPreventer(nil, minMessages, minTotalSize)

	assert.True(t, check.IfNil(qfp))
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewQuotaFloodPreventer_LowerMinMessagesShouldErr(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(&mock.CacherStub{}, minMessages-1, minTotalSize)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_LowerMinSizeShouldErr(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(&mock.CacherStub{}, minMessages, minTotalSize-1)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_ShouldWork(t *testing.T) {
	t.Parallel()

	qfp, err := NewQuotaFloodPreventer(&mock.CacherStub{}, minMessages, minTotalSize)

	assert.False(t, check.IfNil(qfp))
	assert.Nil(t, err)
}

//------- Increment

func TestNewQuotaFloodPreventer_IncrementIdentifierNotPresentPutQuotaAndReturnTrue(t *testing.T) {
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
				if q.numMessages == 1 && q.totalSize == size {
					putWasCalled = true
				}

				return
			},
		},
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Increment("identifier", size)

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_IncrementNotQuotaSavedInCacheShouldPutQuotaAndReturnTrue(t *testing.T) {
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
				if q.numMessages == 1 && q.totalSize == size {
					putWasCalled = true
				}

				return
			},
		},
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Increment("identifier", size)

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_IncrementUnderMaxValuesShouldIncrementAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	existingSize := uint64(minTotalSize * 5)
	existingMessages := uint32(minMessages * 2)
	existingQuota := &quota{
		numMessages: existingMessages,
		totalSize:   existingSize,
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
				if q.numMessages == existingMessages+1 && q.totalSize == existingSize+size {
					putWasCalled = true
				}

				return
			},
		},
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Increment("identifier", size)

	assert.True(t, ok)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_IncrementOverMaxNumMessagesShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	existingMessages := uint32(minMessages + 11)
	existingSize := uint64(minTotalSize * 3)
	existingQuota := &quota{
		numMessages: existingMessages,
		totalSize:   existingSize,
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
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Increment("identifier", minTotalSize)

	assert.False(t, ok)
}

func TestNewQuotaFloodPreventer_IncrementOverMaxSizeShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	existingMessages := uint32(minMessages)
	existingSize := uint64(minTotalSize * 11)
	existingQuota := &quota{
		numMessages: existingMessages,
		totalSize:   existingSize,
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
		minMessages*4,
		minTotalSize*10,
	)

	ok := qfp.Increment("identifier", minTotalSize)

	assert.False(t, ok)
}

func TestCountersMap_IncrementShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	qfp, _ := NewQuotaFloodPreventer(
		mock.NewCacherMock(),
		minMessages,
		minTotalSize)
	numIterations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			ok := qfp.Increment(fmt.Sprintf("%d", idx), minTotalSize)
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
		},
		minTotalSize,
		minMessages,
	)

	qfp.Reset()

	assert.True(t, clearCalled)
}

func TestCountersMap_IncrementAndResetShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	qfp, _ := NewQuotaFloodPreventer(
		mock.NewCacherMock(),
		minMessages,
		minTotalSize,
	)
	numIterations := 1000
	wg := sync.WaitGroup{}
	wg.Add(numIterations + numIterations/10)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			ok := qfp.Increment(fmt.Sprintf("%d", idx), minTotalSize)
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
