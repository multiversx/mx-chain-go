package floodPreventers

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createDefaultArgument() ArgQuotaFloodPreventer {
	return ArgQuotaFloodPreventer{
		Name:                      "test",
		Cacher:                    testscommon.NewCacherStub(),
		StatusHandlers:            []QuotaStatusHandler{&mock.QuotaStatusHandlerStub{}},
		BaseMaxNumMessagesPerPeer: minMessages,
		MaxTotalSizePerPeer:       minTotalSize,
		PercentReserved:           10,
		IncreaseThreshold:         0,
		IncreaseFactor:            0,
	}
}

//------- NewQuotaFloodPreventer

func TestNewQuotaFloodPreventer_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.Cacher = nil
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.True(t, check.IfNil(qfp))
	assert.Equal(t, process.ErrNilCacher, err)
}

func TestNewQuotaFloodPreventer_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.StatusHandlers = []QuotaStatusHandler{nil}
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.True(t, check.IfNil(qfp))
	assert.Equal(t, process.ErrNilQuotaStatusHandler, err)
}

func TestNewQuotaFloodPreventer_LowerMinMessagesPerPeerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.BaseMaxNumMessagesPerPeer = minMessages - 1
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_LowerMinSizePerPeerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.MaxTotalSizePerPeer = minTotalSize - 1
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_HigherPercentReservedShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.PercentReserved = maxPercentReserved + 0.01
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_LowerPercentReservedShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.PercentReserved = minPercentReserved - 0.01
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_NegativeIncreaseFactorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.IncreaseFactor = -0.001
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.True(t, check.IfNil(qfp))
	assert.True(t, errors.Is(err, process.ErrInvalidValue))
}

func TestNewQuotaFloodPreventer_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.False(t, check.IfNil(qfp))
	assert.Nil(t, err)
}

func TestNewQuotaFloodPreventer_NilListShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.StatusHandlers = nil
	qfp, err := NewQuotaFloodPreventer(arg)

	assert.False(t, check.IfNil(qfp))
	assert.Nil(t, err)
}

//------- IncreaseLoad

func TestNewQuotaFloodPreventer_IncreaseLoadIdentifierNotPresentPutQuotaAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	size := uint64(minTotalSize * 5)
	arg := createDefaultArgument()
	arg.Cacher = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		PutCalled: func(key []byte, value interface{}, _ int) (evicted bool) {
			q, isQuota := value.(*quota)
			if !isQuota {
				return
			}
			if q.numReceivedMessages == 1 && q.sizeReceivedMessages == size {
				putWasCalled = true
			}

			return
		},
	}
	arg.BaseMaxNumMessagesPerPeer = minMessages * 4
	arg.MaxTotalSizePerPeer = minTotalSize * 1
	qfp, _ := NewQuotaFloodPreventer(arg)

	err := qfp.IncreaseLoad("identifier", size)

	assert.Nil(t, err)
	assert.True(t, putWasCalled)
}

func TestNewQuotaFloodPreventer_IncreaseLoadNotQuotaSavedInCacheShouldPutQuotaAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	size := uint64(minTotalSize * 5)
	arg := createDefaultArgument()
	arg.Cacher = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return "bad value", true
		},
		PutCalled: func(key []byte, value interface{}, _ int) (evicted bool) {
			q, isQuota := value.(*quota)
			if !isQuota {
				return
			}
			if q.numReceivedMessages == 1 && q.sizeReceivedMessages == size {
				putWasCalled = true
			}

			return
		},
	}
	arg.BaseMaxNumMessagesPerPeer = minMessages * 4
	arg.MaxTotalSizePerPeer = minTotalSize * 1
	qfp, _ := NewQuotaFloodPreventer(arg)

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
	arg := createDefaultArgument()
	arg.Cacher = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return existingQuota, true
		},
	}
	arg.BaseMaxNumMessagesPerPeer = minMessages * 10
	arg.MaxTotalSizePerPeer = minTotalSize * 10
	qfp, _ := NewQuotaFloodPreventer(arg)

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
	arg := createDefaultArgument()
	arg.Cacher = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return existingQuota, true
		},
		PutCalled: func(key []byte, value interface{}, _ int) (evicted bool) {
			assert.Fail(t, "should have not called put")

			return false
		},
	}
	arg.BaseMaxNumMessagesPerPeer = minMessages * 4
	arg.MaxTotalSizePerPeer = minTotalSize * 10
	qfp, _ := NewQuotaFloodPreventer(arg)

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
	arg := createDefaultArgument()
	arg.Cacher = &testscommon.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return existingQuota, true
		},
		PutCalled: func(key []byte, value interface{}, _ int) (evicted bool) {
			assert.Fail(t, "should have not called put")

			return false
		},
	}
	arg.BaseMaxNumMessagesPerPeer = minMessages * 4
	arg.MaxTotalSizePerPeer = minTotalSize * 10
	qfp, _ := NewQuotaFloodPreventer(arg)

	err := qfp.IncreaseLoad("identifier", minTotalSize)

	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestCountersMap_IncreaseLoadShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	numIterations := 1000
	arg := createDefaultArgument()
	arg.Cacher = testscommon.NewCacherMock()
	qfp, _ := NewQuotaFloodPreventer(arg)
	wg := sync.WaitGroup{}
	wg.Add(numIterations)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			err := qfp.IncreaseLoad(core.PeerID(fmt.Sprintf("%d", idx)), minTotalSize)
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
	arg := createDefaultArgument()
	arg.Cacher = &testscommon.CacherStub{
		ClearCalled: func() {
			clearCalled = true
		},
		KeysCalled: func() [][]byte {
			return make([][]byte, 0)
		},
	}
	qfp, _ := NewQuotaFloodPreventer(arg)

	qfp.Reset()

	assert.True(t, clearCalled)
}

func TestCountersMap_ResetShouldCallQuotaStatus(t *testing.T) {
	t.Parallel()

	cacher := testscommon.NewCacherMock()
	key1 := core.PeerID("key1")
	quota1 := &quota{
		numReceivedMessages:   1,
		sizeReceivedMessages:  2,
		numProcessedMessages:  3,
		sizeProcessedMessages: 4,
	}
	key2 := core.PeerID("key2")
	quota2 := &quota{
		numReceivedMessages:   5,
		sizeReceivedMessages:  6,
		numProcessedMessages:  7,
		sizeProcessedMessages: 8,
	}

	cacher.HasOrAdd(key1.Bytes(), quota1, quota1.Size())
	cacher.HasOrAdd(key2.Bytes(), quota2, quota2.Size())

	resetStatisticsCalled := false
	quota1Compared := false
	quota2Compared := false
	arg := createDefaultArgument()
	arg.Cacher = cacher
	arg.StatusHandlers = []QuotaStatusHandler{
		&mock.QuotaStatusHandlerStub{
			ResetStatisticsCalled: func() {
				resetStatisticsCalled = true
			},
			AddQuotaCalled: func(pid core.PeerID, numReceivedMessages uint32, sizeReceivedMessages uint64, numProcessedMessages uint32, sizeProcessedMessages uint64) {
				quotaProvided := quota{
					numReceivedMessages:   numReceivedMessages,
					sizeReceivedMessages:  sizeReceivedMessages,
					numProcessedMessages:  numProcessedMessages,
					sizeProcessedMessages: sizeProcessedMessages,
				}
				quotaToCompare := quota{}

				switch pid {
				case key1:
					quotaToCompare = *quota1
					quota1Compared = true
				case key2:
					quotaToCompare = *quota2
					quota2Compared = true
				default:
					assert.Fail(t, fmt.Sprintf("unknown identifier %s", pid))
				}

				assert.Equal(t, quotaToCompare, quotaProvided)
			},
		},
	}
	qfp, _ := NewQuotaFloodPreventer(arg)

	qfp.Reset()

	assert.True(t, resetStatisticsCalled)
	assert.True(t, quota1Compared)
	assert.True(t, quota2Compared)
}

func TestCountersMap_IncrementAndResetShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	numIterations := 1000
	arg := createDefaultArgument()
	arg.Cacher = testscommon.NewCacherMock()
	qfp, _ := NewQuotaFloodPreventer(arg)
	wg := sync.WaitGroup{}
	wg.Add(numIterations + numIterations/10)
	for i := 0; i < numIterations; i++ {
		go func(idx int) {
			err := qfp.IncreaseLoad(core.PeerID(fmt.Sprintf("%d", idx)), minTotalSize)
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

	numMessages := uint32(100)
	arg := createDefaultArgument()
	arg.Cacher = testscommon.NewCacherMock()
	arg.BaseMaxNumMessagesPerPeer = numMessages
	arg.MaxTotalSizePerPeer = math.MaxUint64
	arg.PercentReserved = float32(17)
	qfp, _ := NewQuotaFloodPreventer(arg)

	identifier := core.PeerID("id")
	for i := uint32(0); i < numMessages-uint32(arg.PercentReserved); i++ {
		err := qfp.IncreaseLoad(identifier, 1)
		assert.Nil(t, err, fmt.Sprintf("failed at the %d iteration", i))
	}

	for i := uint32(0); i < uint32(arg.PercentReserved)*2; i++ {
		err := qfp.IncreaseLoad(identifier, 1)
		assert.True(t, errors.Is(err, process.ErrSystemBusy),
			fmt.Sprintf("failed at the %d iteration", numMessages-uint32(arg.PercentReserved)+i))
	}
}

//------- ApplyConsensusSize

func TestQuotaFloodPreventer_ApplyConsensusSizeInvalidConsensusSize(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.BaseMaxNumMessagesPerPeer = uint32(4682)
	qfp, _ := NewQuotaFloodPreventer(arg)

	qfp.ApplyConsensusSize(0)

	assert.Equal(t, arg.BaseMaxNumMessagesPerPeer, qfp.computedMaxNumMessagesPerPeer)
}

func TestQuotaFloodPreventer_ApplyConsensusUnderThreshold(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.BaseMaxNumMessagesPerPeer = uint32(4682)
	arg.IncreaseThreshold = 100
	qfp, _ := NewQuotaFloodPreventer(arg)

	qfp.ApplyConsensusSize(99)

	assert.Equal(t, arg.BaseMaxNumMessagesPerPeer, qfp.computedMaxNumMessagesPerPeer)
}

func TestQuotaFloodPreventer_ApplyConsensusShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.Cacher = testscommon.NewCacherMock()
	arg.BaseMaxNumMessagesPerPeer = 2000
	arg.IncreaseThreshold = 1000
	arg.IncreaseFactor = 0.25
	qfp, _ := NewQuotaFloodPreventer(arg)
	absoluteExpected := uint32(2250)
	qfp.ApplyConsensusSize(2000)

	assert.Equal(t, absoluteExpected, qfp.computedMaxNumMessagesPerPeer)
	relativeExpected := uint32(2250 * 0.9)
	fmt.Printf("expected relative: %d\n", relativeExpected)
	identifier := core.PeerID("identifier")
	for i := 0; i < int(relativeExpected); i++ {
		err := qfp.IncreaseLoad(identifier, 0)
		assert.Nil(t, err, fmt.Sprintf("on iteration %d", i))
	}

	err := qfp.IncreaseLoad(identifier, 0)
	assert.NotNil(t, err)
}
