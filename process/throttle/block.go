package throttle

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-logger-go"
)

var _ process.BlockSizeThrottler = (*blockSizeThrottle)(nil)

var log = logger.GetOrCreate("process/throttle")

const (
	jumpAbovePercent        = 75
	jumpBelowPercent        = 75
	jumpAboveFactor         = 0.5
	jumpBelowFactor         = 0.5
	maxNumOfStatistics      = 600
	numOfStatisticsToRemove = 100
)

type blockInfo struct {
	succeed        bool
	round          uint64
	size           uint32
	currentMaxSize uint32
}

// blockSizeThrottle implements BlockSizeThrottler interface which throttle block size
type blockSizeThrottle struct {
	statistics     []*blockInfo
	currentMaxSize uint32
	mutThrottler   sync.RWMutex
	minSize        uint32
	maxSize        uint32
}

// NewBlockSizeThrottle creates a new blockSizeThrottle object
func NewBlockSizeThrottle(minSize uint32, maxSize uint32) (*blockSizeThrottle, error) {
	bst := blockSizeThrottle{
		statistics:     make([]*blockInfo, 0),
		currentMaxSize: maxSize,
		minSize:        minSize,
		maxSize:        maxSize,
	}
	return &bst, nil
}

// GetCurrentMaxSize gets the current max size in bytes which could be used in one block, taking into consideration the previous results
func (bst *blockSizeThrottle) GetCurrentMaxSize() uint32 {
	bst.mutThrottler.RLock()
	currentMaxSize := bst.currentMaxSize
	bst.mutThrottler.RUnlock()

	return currentMaxSize
}

// Add adds the new size for last block which has been sent in the given round
func (bst *blockSizeThrottle) Add(round uint64, size uint32) {
	bst.mutThrottler.Lock()
	bst.statistics = append(
		bst.statistics,
		&blockInfo{round: round, size: size, currentMaxSize: bst.currentMaxSize},
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

// ComputeCurrentMaxSize computes the current max size in bytes which could be used in one block, taking into consideration the previous results
func (bst *blockSizeThrottle) ComputeCurrentMaxSize() {
	//TODO: This is the first basic implementation, which will adapt the max size which could be used in one block,
	//based on the last recent history. It will always choose the next value of current max size, as a middle distance between
	//the last succeeded and the last not succeeded actions, or vice-versa, depending of the last action state.
	//This algorithm is good when the network speed/latency is changing during the time, and the node will adapt
	//based on the last recent history and not on some minimum/maximum values recorded in its whole history.

	bst.mutThrottler.Lock()
	defer func() {
		log.Debug("ComputeCurrentMaxSize",
			"current max size", bst.currentMaxSize,
		)
		bst.mutThrottler.Unlock()
	}()

	if len(bst.statistics) == 0 {
		return
	}

	lastActionSucceed := bst.statistics[len(bst.statistics)-1].succeed
	lastActionMaxSize := bst.statistics[len(bst.statistics)-1].currentMaxSize

	if lastActionSucceed {
		bst.currentMaxSize = bst.getMaxSizeWhenSucceed(lastActionMaxSize)
	} else {
		bst.currentMaxSize = bst.getMaxSizeWhenNotSucceed(lastActionMaxSize)
	}
}

func (bst *blockSizeThrottle) getMaxSizeWhenSucceed(lastActionMaxSize uint32) uint32 {
	if lastActionMaxSize >= bst.maxSize {
		return bst.maxSize
	}

	maxSizeUsedWithoutSucceed := bst.getCloserAboveCurrentMaxSizeUsedWithoutSucceed(lastActionMaxSize)
	if lastActionMaxSize*100/maxSizeUsedWithoutSucceed > jumpAbovePercent {
		return maxSizeUsedWithoutSucceed
	}

	increasedMaxSize := core.MaxUint32(1, uint32(float32(maxSizeUsedWithoutSucceed-lastActionMaxSize)*jumpAboveFactor))
	return lastActionMaxSize + increasedMaxSize
}

func (bst *blockSizeThrottle) getCloserAboveCurrentMaxSizeUsedWithoutSucceed(currentMaxSize uint32) uint32 {
	for i := len(bst.statistics) - 1; i >= 0; i-- {
		if !bst.statistics[i].succeed && bst.statistics[i].currentMaxSize > currentMaxSize {
			return bst.statistics[i].currentMaxSize
		}
	}

	return bst.maxSize
}

func (bst *blockSizeThrottle) getMaxSizeWhenNotSucceed(lastActionMaxSize uint32) uint32 {
	if lastActionMaxSize <= bst.minSize {
		return bst.minSize
	}

	maxSizeUsedWithSucceed := bst.getCloserBelowCurrentMaxSizeUsedWithSucceed(lastActionMaxSize)
	if maxSizeUsedWithSucceed*100/lastActionMaxSize > jumpBelowPercent {
		return maxSizeUsedWithSucceed
	}

	decreasedMaxSize := core.MaxUint32(1, uint32(float32(lastActionMaxSize-maxSizeUsedWithSucceed)*jumpBelowFactor))
	return lastActionMaxSize - decreasedMaxSize
}

func (bst *blockSizeThrottle) getCloserBelowCurrentMaxSizeUsedWithSucceed(currentMaxSize uint32) uint32 {
	for i := len(bst.statistics) - 1; i >= 0; i-- {
		if bst.statistics[i].succeed && bst.statistics[i].currentMaxSize < currentMaxSize {
			return bst.statistics[i].currentMaxSize
		}
	}

	return bst.minSize
}

// IsInterfaceNil returns true if there is no value under the interface
func (bst *blockSizeThrottle) IsInterfaceNil() bool {
	return bst == nil
}
