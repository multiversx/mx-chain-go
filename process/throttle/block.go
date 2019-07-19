package throttle

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.DefaultLogger()

type blockInfo struct {
	succeeded bool
	round     uint64
	items     uint32
	maxItems  uint32
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
	defer bst.mutThrottler.RUnlock()
	return bst.maxItems
}

// Update sets the new state for last block which has been sent and updates the max items which could be added
// in one block
func (bst *blockSizeThrottle) Update(succeeded bool, round uint64, items uint32) {
	bst.mutThrottler.Lock()
	bst.statistics = append(
		bst.statistics,
		&blockInfo{succeeded: succeeded, round: round, items: items, maxItems: bst.maxItems},
	)
	bst.computeMaxItems()
	bst.mutThrottler.Unlock()
}

func (bst *blockSizeThrottle) computeMaxItems() {
	log.Info("maximum number of items which could be added in one block is %d", bst.maxItems)
}
