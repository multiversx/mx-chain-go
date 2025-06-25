package txcache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// TODO use a map instead of slice for st.blocks
type selectionTracker struct {
	mutTracker     sync.RWMutex
	latestNonce    uint64
	latestRootHash []byte
	blocks         []*trackedBlock
	txCache        *TxCache
}

// NewSelectionTracker creates a new selectionTracker
func NewSelectionTracker(txCache *TxCache) (*selectionTracker, error) {
	if check.IfNil(txCache) {
		return nil, errNilTxCache
	}
	return &selectionTracker{
		mutTracker: sync.RWMutex{},
		blocks:     make([]*trackedBlock, 0),
		txCache:    txCache,
	}, nil
}

// OnProposedBlock notifies when a block is proposed and updates the state of the selectionTracker
func (st *selectionTracker) OnProposedBlock(blockHash []byte, blockBody *block.Body, handler data.HeaderHandler) error {
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

	txs, err := st.getTransactionsFromBlock(blockBody)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error getting transactions from block", "err", err)
		return err
	}

	tBlock := newTrackedBlock(nonce, blockHash, rootHash, prevHash)
	tBlock.compileBreadcrumbs(txs)
	st.blocks = append(st.blocks, tBlock)

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

func (st *selectionTracker) getTransactionsFromBlock(blockBody *block.Body) ([]*WrappedTransaction, error) {
	miniBlocks := blockBody.GetMiniBlocks()
	txs := make([]*WrappedTransaction, 0)
	for _, miniBlock := range miniBlocks {
		txHashes := miniBlock.GetTxHashes()
		for _, txHash := range txHashes {
			tx, ok := st.txCache.GetByTxHash(txHash)
			if !ok {
				return nil, errNotFoundTx
			}

			txs = append(txs, tx)
		}
	}

	return txs, nil
}
