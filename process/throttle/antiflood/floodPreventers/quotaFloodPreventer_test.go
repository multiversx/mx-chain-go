package floodPreventers

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"

	"github.com/stretchr/testify/assert"
)

func createDefaultArgument() ArgQuotaFloodPreventer {
	return ArgQuotaFloodPreventer{
		Name:             "test",
		Cacher:           cache.NewCacherStub(),
		StatusHandlers:   []QuotaStatusHandler{&mock.QuotaStatusHandlerStub{}},
		AntifloodConfigs: &testscommon.AntifloodConfigsHandlerStub{},
		ConfigFetcher: func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
			return config.FloodPreventerConfig{
				IntervalInSeconds: 1,
				ReservedPercent:   10,
				PeerMaxInput: config.AntifloodLimitsConfig{
					BaseMessagesPerInterval: minMessages,
					TotalSizePerInterval:    minTotalSize,
					IncreaseFactor: config.IncreaseFactorConfig{
						Threshold: 0,
						Factor:    0,
					},
				},
				BlackList: config.BlackListConfig{
					ThresholdNumMessagesPerInterval: 10,
					ThresholdSizePerInterval:        10,
					NumFloodingRounds:               10,
					PeerBanDurationInSeconds:        10,
				},
			}
		},
	}
}

// ------- NewQuotaFloodPreventer

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

// ------- IncreaseLoad

func TestNewQuotaFloodPreventer_IncreaseLoadIdentifierNotPresentPutQuotaAndReturnTrue(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	size := uint64(minTotalSize * 5)
	arg := createDefaultArgument()
	arg.Cacher = &cache.CacherStub{
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

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: minMessages * 4,
				TotalSizePerInterval:    minTotalSize * 1,
			},
		}
	}

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
	arg.Cacher = &cache.CacherStub{
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

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: minMessages * 4,
				TotalSizePerInterval:    minTotalSize * 1,
			},
		}
	}

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
	arg.Cacher = &cache.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return existingQuota, true
		},
	}

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: minMessages * 10,
				TotalSizePerInterval:    minTotalSize * 10,
			},
		}
	}

	qfp, _ := NewQuotaFloodPreventer(arg)

	err := qfp.IncreaseLoad("identifier", size)

	assert.Nil(t, err)
}

// ------- IncreaseLoad per peer

func TestNewQuotaFloodPreventer_IncreaseLoadOverMaxPeerNumMessagesShouldNotPutAndReturnFalse(t *testing.T) {
	t.Parallel()

	existingMessages := uint32(minMessages + 11)
	existingSize := uint64(minTotalSize * 3)
	existingQuota := &quota{
		numReceivedMessages:  existingMessages,
		sizeReceivedMessages: existingSize,
	}
	arg := createDefaultArgument()
	arg.Cacher = &cache.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return existingQuota, true
		},
		PutCalled: func(key []byte, value interface{}, _ int) (evicted bool) {
			assert.Fail(t, "should have not called put")

			return false
		},
	}

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: minMessages * 4,
				TotalSizePerInterval:    minTotalSize * 10,
			},
		}
	}

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
	arg.Cacher = &cache.CacherStub{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return existingQuota, true
		},
		PutCalled: func(key []byte, value interface{}, _ int) (evicted bool) {
			assert.Fail(t, "should have not called put")

			return false
		},
	}

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: minMessages * 4,
				TotalSizePerInterval:    minTotalSize * 10,
			},
		}
	}
	qfp, _ := NewQuotaFloodPreventer(arg)

	err := qfp.IncreaseLoad("identifier", minTotalSize)

	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestCountersMap_IncreaseLoadShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	numIterations := 1000
	arg := createDefaultArgument()
	arg.Cacher = cache.NewCacherMock()
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

// ------- Reset

func TestCountersMap_ResetShouldCallCacherClear(t *testing.T) {
	t.Parallel()

	clearCalled := false
	arg := createDefaultArgument()
	arg.Cacher = &cache.CacherStub{
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

	cacher := cache.NewCacherMock()
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
	arg.Cacher = cache.NewCacherMock()
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
	arg.Cacher = cache.NewCacherMock()

	percentReserved := float32(17)

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			ReservedPercent: percentReserved,
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: numMessages,
				TotalSizePerInterval:    math.MaxUint64,
				IncreaseFactor:          config.IncreaseFactorConfig{},
			},
		}
	}

	qfp, _ := NewQuotaFloodPreventer(arg)

	identifier := core.PeerID("id")
	for i := uint32(0); i < numMessages-uint32(percentReserved); i++ {
		err := qfp.IncreaseLoad(identifier, 1)
		assert.Nil(t, err, fmt.Sprintf("failed at the %d iteration", i))
	}

	for i := uint32(0); i < uint32(percentReserved)*2; i++ {
		err := qfp.IncreaseLoad(identifier, 1)
		assert.True(t, errors.Is(err, process.ErrSystemBusy),
			fmt.Sprintf("failed at the %d iteration", numMessages-uint32(percentReserved)+i))
	}
}

// ------- ApplyConsensusSize

func TestQuotaFloodPreventer_ApplyConsensusSizeInvalidConsensusSize(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	baseMaxNumMessagesPerPeer := uint32(4682)

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: baseMaxNumMessagesPerPeer,
			},
		}
	}

	qfp, _ := NewQuotaFloodPreventer(arg)

	qfp.ApplyConsensusSize(0)

	assert.Equal(t, baseMaxNumMessagesPerPeer, qfp.computedMaxNumMessagesPerPeer)
}

func TestQuotaFloodPreventer_ApplyConsensusUnderThreshold(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()

	baseMaxNumMessagesPerPeer := uint32(4682)

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: baseMaxNumMessagesPerPeer,
				IncreaseFactor: config.IncreaseFactorConfig{
					Threshold: 100,
				},
			},
		}
	}

	qfp, _ := NewQuotaFloodPreventer(arg)

	qfp.ApplyConsensusSize(99)

	assert.Equal(t, baseMaxNumMessagesPerPeer, qfp.computedMaxNumMessagesPerPeer)
}

func TestQuotaFloodPreventer_ApplyConsensusShouldWork(t *testing.T) {
	t.Parallel()

	arg := createDefaultArgument()
	arg.Cacher = cache.NewCacherMock()

	arg.ConfigFetcher = func(confHandler common.AntifloodConfigsHandler, id string) config.FloodPreventerConfig {
		return config.FloodPreventerConfig{
			IntervalInSeconds: 1,
			ReservedPercent:   10,
			PeerMaxInput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: 2000,
				IncreaseFactor: config.IncreaseFactorConfig{
					Threshold: 1000,
					Factor:    0.25,
				},
			},
		}
	}

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
