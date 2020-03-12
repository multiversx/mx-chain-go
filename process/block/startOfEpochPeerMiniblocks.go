package block

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type startOfEpochPeerMiniblocks struct {
	allPeerMiniblocks     map[string]*block.MiniBlock
	headerHash            []byte
	metaHeader            data.HeaderHandler
	chRcvAllMiniblocks    chan struct{}
	mutMiniBlocksForBlock sync.Mutex
}

func (sepm *startOfEpochPeerMiniblocks) init(metaBlock *block.MetaBlock, headerHash []byte) {
	sepm.mutMiniBlocksForBlock.Lock()
	sepm.metaHeader = metaBlock
	sepm.allPeerMiniblocks = make(map[string]*block.MiniBlock)
	sepm.headerHash = headerHash
	sepm.chRcvAllMiniblocks = make(chan struct{})
	sepm.mutMiniBlocksForBlock.Unlock()
}

func (sepm *startOfEpochPeerMiniblocks) setPeerMiniblocks(blocks map[string]*block.MiniBlock) {
	sepm.mutMiniBlocksForBlock.Lock()
	sepm.allPeerMiniblocks = blocks
	sepm.mutMiniBlocksForBlock.Unlock()
}
