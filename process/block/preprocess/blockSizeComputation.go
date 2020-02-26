package preprocess

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
)

type blockSizeComputation struct {
	numMiniBlocks int
	numTxs        int
	mut           sync.RWMutex
}

func NewBlockSizeComputation() *blockSizeComputation {
	return &blockSizeComputation{}
}

func (bsc *blockSizeComputation) Init() {
	bsc.mut.Lock()
	bsc.numMiniBlocks = 0
	bsc.numTxs = 0
	bsc.mut.Unlock()
}

func (bsc *blockSizeComputation) AddNumMiniBlocks(numMiniBlocks int) {
	bsc.mut.Lock()
	bsc.numMiniBlocks += numMiniBlocks
	bsc.mut.Unlock()
}

func (bsc *blockSizeComputation) AddNumTxs(numTxs int) {
	bsc.mut.Lock()
	bsc.numTxs += numTxs
	bsc.mut.Unlock()
}

func (bsc *blockSizeComputation) IsMaxBlockSizeReached(numNewMiniBlocks int, numNewTxs int) bool {
	//TODO: Here should be implemented a real calculation for block size
	bsc.mut.RLock()
	isMaxBlockSizeReached := bsc.numMiniBlocks+numNewMiniBlocks > core.MaxMiniBlocksInBlock ||
		bsc.numTxs+numNewTxs > process.MaxItemsInBlock
	bsc.mut.RUnlock()

	return isMaxBlockSizeReached
}
