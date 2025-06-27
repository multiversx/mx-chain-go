package txcache

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

// TODO use a map instead of slice for st.blocks
type selectionTracker struct {
	mutTracker     sync.RWMutex
	latestNonce    uint64
	latestRootHash []byte
	blocks         []*trackedBlock
}

// NewSelectionTracker creates a new selectionTracker
func NewSelectionTracker() (*selectionTracker, error) {
	return &selectionTracker{
		mutTracker: sync.RWMutex{},
		blocks:     make([]*trackedBlock, 0),
	}, nil
}

// OnProposedBlock notifies when a block is proposed and updates the state of the selectionTracker
func (st *selectionTracker) OnProposedBlock(blockHash []byte, blockBody data.BodyHandler, handler data.HeaderHandler) error {
	if len(blockHash) == 0 {
		return errNilBlockHash
	}
	if check.IfNil(blockBody) {
		return errNilBlockBody
	}
	if check.IfNil(handler) {
		return errNilHeaderHandler
	}

	nonce := handler.GetNonce()
	rootHash := handler.GetRootHash()
	prevHash := handler.GetPrevHash()

	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	log.Debug("selectionTracker.OnProposedBlock",
		"blockHash", blockHash,
		"nonce", nonce,
		"rootHash", rootHash,
		"prevHash", prevHash)

	st.blocks = append(st.blocks, newTrackedBlock(nonce, blockHash, rootHash, prevHash))
	return nil
}

// OnExecutedBlock notifies when a block is executed and updates the state of the selectionTracker
func (st *selectionTracker) OnExecutedBlock(handler data.HeaderHandler) error {
	if check.IfNil(handler) {
		return errNilHeaderHandler
	}

	nonce := handler.GetNonce()
	rootHash := handler.GetRootHash()
	prevHash := handler.GetPrevHash()

	tempTrackedBlock := newTrackedBlock(nonce, nil, rootHash, prevHash)
	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	st.removeFromTrackedBlocksNoLock(tempTrackedBlock)
	st.updateLatestRootHashNoLock(nonce, rootHash)

	return nil
}

func (st *selectionTracker) removeFromTrackedBlocksNoLock(searchedBlock *trackedBlock) {
	remainingBlocks := make([]*trackedBlock, 0)
	for _, block := range st.blocks {
		if !searchedBlock.sameNonce(block) {
			remainingBlocks = append(remainingBlocks, block)
		}
	}

	log.Debug("selectionTracker.removeFromTrackedBlocksNoLock",
		"searched block nonce", searchedBlock.nonce,
		"searched block hash", searchedBlock.hash,
		"searched block rootHash", searchedBlock.rootHash,
		"searched block prevHash", searchedBlock.prevHash,
		"removed blocks", len(st.blocks)-len(remainingBlocks),
	)

	st.blocks = remainingBlocks
}

func (st *selectionTracker) updateLatestRootHashNoLock(receivedNonce uint64, receivedRootHash []byte) {
	log.Debug("selectionTracker.updateLatestRootHashNoLock",
		"received root hash", receivedRootHash,
		"received nonce", receivedNonce)

	if st.latestRootHash == nil {
		st.latestRootHash = receivedRootHash
		st.latestNonce = receivedNonce
		return
	}

	if receivedNonce > st.latestNonce {
		st.latestRootHash = receivedRootHash
		st.latestNonce = receivedNonce
	}
}

func (st *selectionTracker) deriveVirtualSelectionSession(session SelectionSession,
	latestExecutedBlockHash []byte, currentBlockNonce uint64) (*virtualSelectionSession, error) {
	rootHash, err := session.GetRootHash()
	if err != nil {
		log.Debug("selectionTracker.deriveVirtualSelectionSession",
			"err", err)
		return nil, err
	}

	log.Debug("selectionTracker.deriveVirtualSelectionSession", "rootHash", rootHash)

	_ = st.getChainOfTrackedBlocks(latestExecutedBlockHash, currentBlockNonce)

	return &virtualSelectionSession{}, nil
}

func (st *selectionTracker) getChainOfTrackedBlocks(latestExecutedBlockHash []byte, beforeNonce uint64) []*trackedBlock {
	chain := make([]*trackedBlock, 0)
	nextBlock := st.findNextBlock(latestExecutedBlockHash)

	for nextBlock != nil && nextBlock.nonce < beforeNonce {
		chain = append(chain, nextBlock)
		blockHash := nextBlock.hash
		nextBlock = st.findNextBlock(blockHash)
	}

	return chain
}

// TODO solve the case of forks
func (st *selectionTracker) findNextBlock(previousHash []byte) *trackedBlock {
	for _, block := range st.blocks {
		if bytes.Equal(previousHash, block.prevHash) {
			return block
		}
	}

	return nil
}
