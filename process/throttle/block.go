package throttle

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.DefaultLogger()

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
	defer bst.mutThrottler.RUnlock()
	return bst.maxItems
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
	for index := range bst.statistics {
		if bst.statistics[index].round == round {
			bst.statistics[index].succeed = true
		}
	}
	bst.mutThrottler.Unlock()
}

// ComputeMaxItems computes the max items which could be added in one block, taking into consideration the previous
// results
func (bst *blockSizeThrottle) ComputeMaxItems() {
	log.Info("maximum number of items which could be added in one block is %d", bst.maxItems)
}
