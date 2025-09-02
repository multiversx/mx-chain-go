package txcache

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
)

type selectionTracker struct {
	mutTracker       sync.RWMutex
	latestNonce      uint64
	latestRootHash   []byte
	blocks           map[string]*trackedBlock
	txCache          txCacheForSelectionTracker
	maxTrackedBlocks uint32
}

// NewSelectionTracker creates a new selectionTracker
func NewSelectionTracker(txCache txCacheForSelectionTracker, maxTrackedBlocks uint32) (*selectionTracker, error) {
	if check.IfNil(txCache) {
		return nil, errNilTxCache
	}
	// TODO compare with the maximum allowed offset between proposing a block and actually executing it
	if maxTrackedBlocks == 0 {
		return nil, errInvalidMaxTrackedBlocks
	}
	return &selectionTracker{
		mutTracker:       sync.RWMutex{},
		blocks:           make(map[string]*trackedBlock),
		txCache:          txCache,
		maxTrackedBlocks: maxTrackedBlocks,
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
	err := st.verifyArgs(blockHash, blockBody, blockHeader, accountsProvider)
	if err != nil {
		return err
	}

	nonce := blockHeader.GetNonce()
	rootHash := blockHeader.GetRootHash()
	prevHash := blockHeader.GetPrevHash()

	tBlock := newTrackedBlock(nonce, blockHash, rootHash, prevHash)

	log.Debug("selectionTracker.OnProposedBlock",
		"nonce", nonce,
		"blockHash", blockHash,
		"rootHash", rootHash,
		"prevHash", prevHash,
	)

	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	err = st.checkReceivedBlockNoLock(blockBody, blockHeader)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error checking the received block", "err", err)
		return err
	}

	err = st.validateTrackedBlocks(blockBody, tBlock, accountsProvider, blockchainInfo)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error validating the tracked blocks", "err", err)
		return err
	}

	st.addNewTrackedBlockNoLock(blockHash, tBlock)
	return nil
}

// OnExecutedBlock notifies when a block is executed and updates the state of the selectionTracker
func (st *selectionTracker) OnExecutedBlock(blockHeader data.HeaderHandler) error {
	if check.IfNil(blockHeader) {
		return errNilHeaderHandler
	}

	nonce := blockHeader.GetNonce()
	rootHash := blockHeader.GetRootHash()
	prevHash := blockHeader.GetPrevHash()

	log.Debug("selectionTracker.OnExecutedBlock",
		"nonce", nonce,
		"rootHash", rootHash,
		"prevHash", prevHash,
	)

	tempTrackedBlock := newTrackedBlock(nonce, nil, rootHash, prevHash)

	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	st.removeFromTrackedBlocksNoLock(tempTrackedBlock)
	st.updateLatestRootHashNoLock(nonce, rootHash)

	return nil
}

func (st *selectionTracker) verifyArgs(
	blockHash []byte,
	blockBody *block.Body,
	blockHeader data.HeaderHandler,
	accountsProvider AccountNonceAndBalanceProvider,
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

	return nil
}

// checkReceivedBlockNoLock first checks if MaxTrackedBlocks is reached
// if MaxTrackedBlocks is reached, the received block must either have an empty body or contain new execution results.
func (st *selectionTracker) checkReceivedBlockNoLock(blockBody *block.Body, blockHeader data.HeaderHandler) error {
	if len(st.blocks) < int(st.maxTrackedBlocks) {
		return nil
	}

	hasNewTransactions := len(blockBody.MiniBlocks) != 0
	hasNoNewExecutionResults := len(blockHeader.GetExecutionResultsHandlers()) == 0

	if hasNewTransactions && hasNoNewExecutionResults {
		log.Warn("selectionTracker.checkReceivedBlockNoLock: received bad block while max tracked blocks is reached. "+
			"should receive empty block or a block with new execution results",
			"len(st.blocks)", len(st.blocks),
		)

		return errBadBlockWhileMaxTrackedBlocksReached
	}

	log.Warn("selectionTracker.checkReceivedBlockNoLock: max tracked blocks reached "+
		"but received a tolerated block - an empty block or a block with new execution results",
		"len(st.blocks)", len(st.blocks))

	return nil
}

func (st *selectionTracker) validateTrackedBlocks(
	blockBody *block.Body,
	tBlock *trackedBlock,
	accountsProvider AccountNonceAndBalanceProvider,
	blockchainInfo common.BlockchainInfo,
) error {
	blocksToBeValidated, err := st.getChainOfTrackedBlocks(
		blockchainInfo.GetLatestExecutedBlockHash(),
		tBlock.prevHash,
		tBlock.nonce,
	)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocks: error creating chain of tracked blocks", "err", err)
		return err
	}

	// if we pass the first validation, only then we extract the txs to compile the breadcrumbs
	txs, err := st.getTransactionsInBlock(blockBody)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocks: error getting transactions from block", "err", err)
		return err
	}

	err = tBlock.compileBreadcrumbs(txs)
	if err != nil {
		log.Debug("selectionTracked.validateTrackedBlocks: error compiling breadcrumbs")
		return err
	}

	// add the new block in the returned chain
	blocksToBeValidated = append(blocksToBeValidated, tBlock)

	// make sure that the breadcrumbs of the proposed block are valid
	// i.e. continuous with the other proposed blocks and no balance issues
	err = st.validateBreadcrumbsOfTrackedBlocks(blocksToBeValidated, accountsProvider)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocks: error validating tracked blocks", "err", err)
		return err
	}

	return nil
}

func (st *selectionTracker) validateBreadcrumbsOfTrackedBlocks(chainOfTrackedBlocks []*trackedBlock, accountsProvider AccountNonceAndBalanceProvider) error {
	validator := newBreadcrumbValidator()

	for _, tb := range chainOfTrackedBlocks {
		for address, breadcrumb := range tb.breadcrumbsByAddress {
			initialNonce, initialBalance, _, err := accountsProvider.GetAccountNonceAndBalance([]byte(address))
			if err != nil {
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks",
					"err", err,
					"address", address,
					"tracked block rootHash", tb.rootHash)
				return err
			}

			if !validator.continuousBreadcrumb(address, initialNonce, breadcrumb) {
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks",
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
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks validation failed",
					"err", err,
					"address", address,
					"tracked block rootHash", tb.rootHash)
				return err
			}
		}
	}

	return nil
}

// addNewTrackedBlockNoLock adds a new tracked block into the map of tracked blocks
// replaces an existing block which has the same nonce with the one received
func (st *selectionTracker) addNewTrackedBlockNoLock(blockToBeAddedHash []byte, blockToBeAdded *trackedBlock) {
	// search if in the tracked block we already have one with same nonce
	for bHash, b := range st.blocks {
		if b.sameNonce(blockToBeAdded) {
			// delete that block and break because there should be maximum one tracked block with that nonce
			delete(st.blocks, bHash)

			log.Debug("selectionTracker.addNewTrackedBlockNoLock block with same nonce was deleted, to be replaced",
				"nonce", blockToBeAdded.nonce,
				"hash of replaced block", b.hash,
				"hash of new block", blockToBeAddedHash,
			)

			break
		}
	}

	// add the new block
	st.blocks[string(blockToBeAddedHash)] = blockToBeAdded
}

func (st *selectionTracker) removeFromTrackedBlocksNoLock(searchedBlock *trackedBlock) {
	removedBlocks := 0
	for blockHash, b := range st.blocks {
		if b.sameNonceOrBelow(searchedBlock) {
			delete(st.blocks, blockHash)
			removedBlocks++
		}
	}

	log.Debug("selectionTracker.removeFromTrackedBlocksNoLock",
		"searched block nonce", searchedBlock.nonce,
		"searched block hash", searchedBlock.hash,
		"searched block rootHash", searchedBlock.rootHash,
		"searched block prevHash", searchedBlock.prevHash,
		"removed blocks", removedBlocks,
	)
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

	trackedBlocks, err := st.getChainOfTrackedBlocks(
		blockchainInfo.GetLatestExecutedBlockHash(),
		blockchainInfo.GetLatestCommittedBlockHash(),
		blockchainInfo.GetCurrentNonce(),
	)
	if err != nil {
		log.Debug("selectionTracker.deriveVirtualSelectionSession",
			"err", err)
		return nil, err
	}

	log.Debug("selectionTracker.deriveVirtualSelectionSession",
		"len(trackedBlocks)", len(trackedBlocks))

	provider := newVirtualSessionProvider(session)
	return provider.createVirtualSelectionSession(trackedBlocks)
}

// getChainOfTrackedBlocks finds the chain of tracked blocks, iterating from tail to head,
// following the previous hash of each block, in order to avoid fork scenarios
// the iteration stops when the previous hash of a block is equal to latestExecutedBlockHash
func (st *selectionTracker) getChainOfTrackedBlocks(
	latestExecutedBlockHash []byte,
	previousHashToBeFound []byte,
	nextNonce uint64,
) ([]*trackedBlock, error) {
	chain := make([]*trackedBlock, 0)

	// if the previous hash to be found is equal to the latest executed hash
	// it means that we do not have any block to be found
	// the block found would be the actual executed block, but that one is not tracked anymore
	if bytes.Equal(latestExecutedBlockHash, previousHashToBeFound) {
		return chain, nil
	}

	// search for the block with the hash equal to the previous hash
	// NOTE: we expect a nil value for a key (block hash) which is not in the map of tracked blocks
	previousBlock := st.blocks[string(previousHashToBeFound)]

	for {
		if nextNonce == 0 {
			break
		}

		// if no block was found, it means there is a gap and we have to return an error
		if previousBlock == nil {
			return nil, errPreviousBlockNotFound
		}

		// extra check for a nonce gap
		hasDiscontinuousBlockNonce := previousBlock.nonce != nextNonce-1
		if hasDiscontinuousBlockNonce {
			return nil, errDiscontinuousSequenceOfBlocks
		}

		// if the block passes the validation, add it to the returned chain
		chain = append(chain, previousBlock)

		// move backwards in the chain and check if the head was reached
		previousBlockHash := previousBlock.prevHash
		if bytes.Equal(latestExecutedBlockHash, previousBlockHash) {
			break
		}

		// update also the nonce
		nextNonce -= 1

		// find the previous block
		previousBlock = st.blocks[string(previousBlockHash)]
	}

	// to be able to validate the blocks later, reverse the order of the blocks to have them from head to tail
	return st.reverseOrderOfBlocks(chain), nil
}

func (st *selectionTracker) reverseOrderOfBlocks(chainOfTrackedBlocks []*trackedBlock) []*trackedBlock {
	reversedChainOfTrackedBlocks := make([]*trackedBlock, 0, len(chainOfTrackedBlocks))
	for i := len(chainOfTrackedBlocks) - 1; i >= 0; i-- {
		reversedChainOfTrackedBlocks = append(reversedChainOfTrackedBlocks, chainOfTrackedBlocks[i])
	}

	return reversedChainOfTrackedBlocks
}
