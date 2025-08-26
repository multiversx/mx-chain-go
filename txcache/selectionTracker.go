package txcache

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
)

// TODO use a map instead of slice for st.blocks
type selectionTracker struct {
	mutTracker     sync.RWMutex
	latestNonce    uint64
	latestRootHash []byte
	blocks         []*trackedBlock
	txCache        txCacheForSelectionTracker
}

// NewSelectionTracker creates a new selectionTracker
func NewSelectionTracker(txCache txCacheForSelectionTracker) (*selectionTracker, error) {
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
func (st *selectionTracker) OnProposedBlock(
	blockHash []byte,
	blockBody *block.Body,
	blockHeader data.HeaderHandler,
	accountsProvider AccountNonceAndBalanceProvider,
	blockchainInfo common.BlockchainInfo,
) error {
	if len(blockHash) == 0 {
		return errNilBlockHash
	}
	if check.IfNil(blockBody) {
		return errNilBlockBody
	}
	if check.IfNil(blockHeader) {
		return errNilHeaderHandler
	}
	if check.IfNil(accountsProvider) {
		return errNilAccountNonceAndBalanceProvider
	}

	nonce := blockHeader.GetNonce()
	rootHash := blockHeader.GetRootHash()
	prevHash := blockHeader.GetPrevHash()

	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	log.Debug("selectionTracker.OnProposedBlock",
		"blockHash", blockHash,
		"nonce", nonce,
		"rootHash", rootHash,
		"prevHash", prevHash)

	txs, err := st.getTransactionsInBlock(blockBody)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error getting transactions from block", "err", err)
		return err
	}

	tBlock, err := newTrackedBlock(nonce, blockHash, rootHash, prevHash, txs)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error creating tracked block", "err", err)
		return err
	}

	st.blocks = append(st.blocks, tBlock)

	blocksToBeValidated := st.getChainOfTrackedBlocks(blockchainInfo.GetLatestExecutedBlockHash(), blockchainInfo.GetCurrentNonce())
	// make sure that the proposed block is valid (continuous with the other proposed blocks and no balance issues)
	return st.validateTrackedBlocks(blocksToBeValidated, accountsProvider)
}

// OnExecutedBlock notifies when a block is executed and updates the state of the selectionTracker
func (st *selectionTracker) OnExecutedBlock(handler data.HeaderHandler) error {
	if check.IfNil(handler) {
		return errNilHeaderHandler
	}

	nonce := handler.GetNonce()
	rootHash := handler.GetRootHash()
	prevHash := handler.GetPrevHash()

	tempTrackedBlock, err := newTrackedBlock(nonce, nil, rootHash, prevHash, nil)
	if err != nil {
		return err
	}
	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	st.removeFromTrackedBlocksNoLock(tempTrackedBlock)
	st.updateLatestRootHashNoLock(nonce, rootHash)

	return nil
}

func (st *selectionTracker) validateTrackedBlocks(chainOfTrackedBlocks []*trackedBlock, accountsProvider AccountNonceAndBalanceProvider) error {
	validator := newBreadcrumbValidator()

	for _, tb := range chainOfTrackedBlocks {
		for address, breadcrumb := range tb.breadcrumbsByAddress {
			initialNonce, initialBalance, _, err := accountsProvider.GetAccountNonceAndBalance([]byte(address))
			if err != nil {
				log.Debug("selectionTracker.validateTrackedBlocks",
					"err", err,
					"address", address,
					"tracked block rootHash", tb.rootHash)
				return err
			}

			if !validator.continuousBreadcrumb(address, initialNonce, breadcrumb) {
				log.Debug("selectionTracker.validateTrackedBlocks",
					"err", errDiscontinuousBreadcrumbs,
					"address", address,
					"tracked block rootHash", tb.rootHash)
				return errDiscontinuousBreadcrumbs
			}

			// TODO re-brainstorm, validate with more integration tests
			// use its balance to accumulate and validate (make sure is < than initialBalance from the session)
			err = validator.validateBalance(address, initialBalance, breadcrumb)
			if err != nil {
				// exit at the first failure
				log.Debug("selectionTracker.validateTrackedBlocks validation failed",
					"err", err,
					"address", address,
					"tracked block rootHash", tb.rootHash)
				return err
			}
		}
	}

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

func (st *selectionTracker) getTransactionsInBlock(blockBody *block.Body) ([]*WrappedTransaction, error) {
	miniBlocks := blockBody.GetMiniBlocks()
	numberOfTxs := st.computeNumberOfTxsInMiniBlocks(miniBlocks)
	txs := make([]*WrappedTransaction, 0, numberOfTxs)

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

func (st *selectionTracker) computeNumberOfTxsInMiniBlocks(miniBlocks []*block.MiniBlock) int {
	numberOfTxs := 0
	for _, miniBlock := range miniBlocks {
		numberOfTxs += len(miniBlock.GetTxHashes())
	}

	return numberOfTxs
}

func (st *selectionTracker) deriveVirtualSelectionSession(
	session SelectionSession,
	blockchainInfo common.BlockchainInfo,
) (*virtualSelectionSession, error) {
	rootHash, err := session.GetRootHash()
	if err != nil {
		log.Debug("selectionTracker.deriveVirtualSelectionSession",
			"err", err)
		return nil, err
	}

	log.Debug("selectionTracker.deriveVirtualSelectionSession", "rootHash", rootHash)

	trackedBlocks := st.getChainOfTrackedBlocks(blockchainInfo.GetLatestExecutedBlockHash(), blockchainInfo.GetCurrentNonce())
	log.Debug("selectionTracker.deriveVirtualSelectionSession",
		"len(trackedBlocks)", len(trackedBlocks))

	provider := newVirtualSessionProvider(session)
	return provider.createVirtualSelectionSession(trackedBlocks)
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
