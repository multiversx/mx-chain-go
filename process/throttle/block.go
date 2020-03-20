package throttle

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("process/throttle")

const (
	jumpAbovePercent        = 90
	jumpBelowPercent        = 90
	jumpAboveFactor         = 0.5
	jumpBelowFactor         = 0.5
	maxNumOfStatistics      = 600
	numOfStatisticsToRemove = 100
)

type blockInfo struct {
	succeed bool
	round   uint64
	size    uint32
	maxSize uint32
}

// blockSizeThrottle implements BlockSizeThrottler interface which throttle block size
type blockSizeThrottle struct {
	statistics   []*blockInfo
	maxSize      uint32
	mutThrottler sync.RWMutex
}

// NewBlockSizeThrottle creates a new blockSizeThrottle object
func NewBlockSizeThrottle() (*blockSizeThrottle, error) {
	bst := blockSizeThrottle{}
	bst.statistics = make([]*blockInfo, 0)
	bst.maxSize = core.MaxSizeInBytes
	return &bst, nil
}

// GetMaxSize gets the max size in bytes which could be used in one block, taking into consideration the previous results
func (bst *blockSizeThrottle) GetMaxSize() uint32 {
	bst.mutThrottler.RLock()
	maxSize := bst.maxSize
	bst.mutThrottler.RUnlock()

	return maxSize
}

// Add adds the new size for last block which has been sent in the given round
func (bst *blockSizeThrottle) Add(round uint64, size uint32) {
	if size < core.MinSizeInBytes || size > core.MaxSizeInBytes {
		return
	}
	bst.mutThrottler.Lock()
	bst.statistics = append(
		bst.statistics,
		&blockInfo{round: round, size: size, maxSize: bst.maxSize},
	)

	if len(bst.statistics) > maxNumOfStatistics {
		bst.statistics = bst.statistics[numOfStatisticsToRemove:]
	}

	bst.mutThrottler.Unlock()
}

// Succeed sets the state of the last block which has been sent in the given round
func (bst *blockSizeThrottle) Succeed(round uint64) {
	bst.mutThrottler.Lock()
	for i := len(bst.statistics) - 1; i >= 0; i-- {
		if bst.statistics[i].round == round {
			bst.statistics[i].succeed = true
			break
		}
	}
	bst.mutThrottler.Unlock()
}

// ComputeMaxSize computes the max size in bytes which could be used in one block, taking into consideration the previous results
func (bst *blockSizeThrottle) ComputeMaxSize() {
	//TODO: This is the first basic implementation, which will adapt the max size which could be used in one block,
	//based on the last recent history. It will always choose the next value of max size, as a middle distance between
	//the last succeeded and the last not succeeded actions, or vice-versa, depending of the last action state.
	//This algorithm is good when the network speed/latency is changing during the time, and the node will adapt
	//based on the last recent history and not on some minimum/maximum values recorded in its whole history.

	bst.mutThrottler.Lock()
	defer func() {
		log.Debug("ComputeMaxSize",
			"max size", bst.maxSize,
		)
		bst.mutThrottler.Unlock()
	}()

	if len(bst.statistics) == 0 {
		return
	}

	lastActionSucceed := bst.statistics[len(bst.statistics)-1].succeed
	lastActionMaxSize := bst.statistics[len(bst.statistics)-1].maxSize

	if lastActionSucceed {
		bst.maxSize = bst.getMaxSizeWhenSucceed(lastActionMaxSize)
	} else {
		bst.maxSize = bst.getMaxSizeWhenNotSucceed(lastActionMaxSize)
	}
}

func (bst *blockSizeThrottle) getMaxSizeWhenSucceed(lastActionMaxSize uint32) uint32 {
	if lastActionMaxSize >= core.MaxSizeInBytes {
		return core.MaxSizeInBytes
	}

	maxSizeUsedWithoutSucceed := bst.getCloserAboveMaxSizeUsedWithoutSucceed(lastActionMaxSize)
	if lastActionMaxSize*100/maxSizeUsedWithoutSucceed > jumpAbovePercent {
		return maxSizeUsedWithoutSucceed
	}

	increasedMaxSize := core.MaxUint32(1, uint32(float32(maxSizeUsedWithoutSucceed-lastActionMaxSize)*jumpAboveFactor))
	return lastActionMaxSize + increasedMaxSize
}

func (bst *blockSizeThrottle) getCloserAboveMaxSizeUsedWithoutSucceed(currentMaxSize uint32) uint32 {
	for i := len(bst.statistics) - 1; i >= 0; i-- {
		if !bst.statistics[i].succeed && bst.statistics[i].maxSize > currentMaxSize {
			return bst.statistics[i].maxSize
		}
	}

	return core.MaxSizeInBytes
}

func (bst *blockSizeThrottle) getMaxSizeWhenNotSucceed(lastActionMaxSize uint32) uint32 {
	if lastActionMaxSize <= core.MinSizeInBytes {
		return core.MinSizeInBytes
	}

	maxSizeUsedWithSucceed := bst.getCloserBelowMaxSizeUsedWithSucceed(lastActionMaxSize)
	if maxSizeUsedWithSucceed*100/lastActionMaxSize > jumpBelowPercent {
		return maxSizeUsedWithSucceed
	}

	decreasedMaxSize := core.MaxUint32(1, uint32(float32(lastActionMaxSize-maxSizeUsedWithSucceed)*jumpBelowFactor))
	return lastActionMaxSize - decreasedMaxSize
}

func (bst *blockSizeThrottle) getCloserBelowMaxSizeUsedWithSucceed(currentMaxSize uint32) uint32 {
	for i := len(bst.statistics) - 1; i >= 0; i-- {
		if bst.statistics[i].succeed && bst.statistics[i].maxSize < currentMaxSize {
			return bst.statistics[i].maxSize
		}
	}

	return core.MinSizeInBytes
}

// IsInterfaceNil returns true if there is no value under the interface
func (bst *blockSizeThrottle) IsInterfaceNil() bool {
	return bst == nil
}
