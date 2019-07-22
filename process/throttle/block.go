package throttle

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.DefaultLogger()

const jumpAbovePercent = 90
const jumpBelowPercent = 90
const jumpAboveFactor = 0.5
const jumpBelowFactor = 0.5

type blockInfo struct {
	succeed  bool
	round    uint64
	items    uint32
	maxItems uint32
}

// blockSizeThrottle implements BlockSizeThrottler interface which throttle block size
type blockSizeThrottle struct {
	statistics   []*blockInfo
	maxItems     uint32
	mutThrottler sync.RWMutex
}

// NewBlockSizeThrottle creates a new blockSizeThrottle object
func NewBlockSizeThrottle() (*blockSizeThrottle, error) {
	bst := blockSizeThrottle{}
	bst.statistics = make([]*blockInfo, 0)
	bst.maxItems = process.MaxItemsInBlock
	return &bst, nil
}

// MaxItemsToAdd gets the maximum number of items which could be added in one block, taking into consideration the
// previous results
func (bst *blockSizeThrottle) MaxItemsToAdd() uint32 {
	bst.mutThrottler.RLock()
	maxItems := bst.maxItems
	bst.mutThrottler.RUnlock()

	return maxItems
}

// Add adds the new state for last block which has been sent and updates the max items which could be added
// in one block
func (bst *blockSizeThrottle) Add(round uint64, items uint32) {
	bst.mutThrottler.Lock()
	bst.statistics = append(
		bst.statistics,
		&blockInfo{round: round, items: items, maxItems: bst.maxItems},
	)
	bst.mutThrottler.Unlock()
}

// Succeed sets the state of the last block which has been sent at the given round
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

// ComputeMaxItems computes the max items which could be added in one block, taking into consideration the previous
// results
func (bst *blockSizeThrottle) ComputeMaxItems() {
	//TODO: This is the first basic implementation, which will adapt the max items which could be added in one block,
	//based on the last recent history. It will always choose the next value of max items, as a middle distance between
	//the last succeeded and the last not succeeded actions, or vice-versa, depending of the last action state.
	//This algorithm is good when the network speed/latency is changing during the time, and the node will adapt
	//based on the last recent history and not on some minimum/maximum values recorded in its whole history.

	bst.mutThrottler.Lock()
	defer func() {
		log.Info(fmt.Sprintf("max number of items which could be added in one block is %d\n", bst.maxItems))
		bst.mutThrottler.Unlock()
	}()

	if len(bst.statistics) == 0 {
		return
	}

	lastActionSucceed := bst.statistics[len(bst.statistics)-1].succeed
	lastActionMaxItems := bst.statistics[len(bst.statistics)-1].maxItems

	if lastActionSucceed {
		bst.maxItems = bst.getMaxItemsWhenSucceed(lastActionMaxItems)
	} else {
		bst.maxItems = bst.getMaxItemsWhenNotSucceed(lastActionMaxItems)
	}
}

func (bst *blockSizeThrottle) getMaxItemsWhenSucceed(lastActionMaxItems uint32) uint32 {
	if lastActionMaxItems == process.MaxItemsInBlock {
		return lastActionMaxItems
	}

	noOfMaxItemsUsedWithoutSucceed := bst.getCloserAboveMaxItemsUsedWithoutSucceed(lastActionMaxItems)
	if lastActionMaxItems*100/noOfMaxItemsUsedWithoutSucceed > jumpAbovePercent {
		return noOfMaxItemsUsedWithoutSucceed
	}

	increasedNoOfItems := core.Max(1, uint32(float32(noOfMaxItemsUsedWithoutSucceed-lastActionMaxItems)*jumpAboveFactor))
	return lastActionMaxItems + increasedNoOfItems
}

func (bst *blockSizeThrottle) getCloserAboveMaxItemsUsedWithoutSucceed(currentMaxItems uint32) uint32 {
	for i := len(bst.statistics) - 1; i >= 0; i-- {
		if !bst.statistics[i].succeed && bst.statistics[i].maxItems > currentMaxItems {
			return bst.statistics[i].maxItems
		}
	}

	return process.MaxItemsInBlock
}

func (bst *blockSizeThrottle) getMaxItemsWhenNotSucceed(lastActionMaxItems uint32) uint32 {
	if lastActionMaxItems == 0 {
		return lastActionMaxItems
	}

	noOfMaxItemsUsedWithSucceed := bst.getCloserBelowMaxItemsUsedWithSucceed(lastActionMaxItems)
	if noOfMaxItemsUsedWithSucceed*100/lastActionMaxItems > jumpBelowPercent {
		return noOfMaxItemsUsedWithSucceed
	}

	decreasedNoOfItems := core.Max(1, uint32(float32(lastActionMaxItems-noOfMaxItemsUsedWithSucceed)*jumpBelowFactor))
	return lastActionMaxItems - decreasedNoOfItems
}

func (bst *blockSizeThrottle) getCloserBelowMaxItemsUsedWithSucceed(currentMaxItems uint32) uint32 {
	for i := len(bst.statistics) - 1; i >= 0; i-- {
		if bst.statistics[i].succeed && bst.statistics[i].maxItems < currentMaxItems {
			return bst.statistics[i].maxItems
		}
	}

	return 0
}
