package txcache

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
	"golang.org/x/exp/slices"
)

// TODO rename this to proposedBlocksTracker
type selectionTracker struct {
	mutTracker                sync.RWMutex
	latestNonce               uint64
	latestRootHash            []byte
	blocks                    map[string]*trackedBlock
	globalBreadcrumbsCompiler *globalAccountBreadcrumbsCompiler
	txCache                   txCacheForSelectionTracker
	maxTrackedBlocks          uint32
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
		mutTracker:                sync.RWMutex{},
		blocks:                    make(map[string]*trackedBlock),
		globalBreadcrumbsCompiler: newGlobalAccountBreadcrumbsCompiler(),
		txCache:                   txCache,
		maxTrackedBlocks:          maxTrackedBlocks,
	}, nil
}

// OnProposedBlock notifies when a block is proposed and updates the state of the selectionTracker.
// blockHash is the hash of the new proposed block.
// blockBody contains the transactions of the new block (required for creating the breadcrumbs and validating the block).
// blockHeader contains the nonce, the rootHash and the previousHash of the new proposed block.
// accountsProvider is a wrapper over the current blockchain state.
// blockchainInfo must contain the information about the last executed block. The other information is not used in this flow.
func (st *selectionTracker) OnProposedBlock(
	blockHash []byte,
	blockBody *block.Body,
	blockHeader data.HeaderHandler,
	accountsProvider common.AccountNonceAndBalanceProvider,
	blockchainInfo common.BlockchainInfo,
) error {
	err := st.verifyArgsOfOnProposedBlock(blockHash, blockBody, blockHeader, accountsProvider)
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

	err = st.validateTrackedBlocksAndCompileBreadcrumbsNoLock(blockBody, tBlock, accountsProvider, blockchainInfo)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error validating the tracked blocks", "err", err)
		return err
	}

	err = st.addNewTrackedBlockNoLock(blockHash, tBlock)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error adding the new tracked block", "err", err)
		return err
	}
	return nil
}

func (st *selectionTracker) verifyArgsOfOnProposedBlock(
	blockHash []byte,
	blockBody *block.Body,
	blockHeader data.HeaderHandler,
	accountsProvider common.AccountNonceAndBalanceProvider,
) error {
	if len(blockHash) == 0 {
		return errNilBlockHash
	}
	if check.IfNil(blockBody) {
		return errNilBlockBody
	}
	if check.IfNil(blockHeader) {
		return errNilBlockHeader
	}
	if check.IfNil(accountsProvider) {
		return errNilAccountNonceAndBalanceProvider
	}

	return nil
}

// checkReceivedBlockNoLock first checks if MaxTrackedBlocks is reached.
// If MaxTrackedBlocks is reached, the received block must either have an empty body or contain new execution results.
func (st *selectionTracker) checkReceivedBlockNoLock(blockBody *block.Body, blockHeader data.HeaderHandler) error {
	if len(st.blocks) < int(st.maxTrackedBlocks) {
		return nil
	}

	hasNewTransactions := len(blockBody.MiniBlocks) != 0
	hasNoNewExecutionResults := len(blockHeader.GetExecutionResultsHandlers()) == 0

	// should receive empty block or a block with new execution results
	if hasNewTransactions && hasNoNewExecutionResults {
		log.Warn("selectionTracker.checkReceivedBlockNoLock: received non-tolerated block while max tracked blocks is reached. "+
			"len(st.blocks)", len(st.blocks),
		)

		return errBadBlockWhileMaxTrackedBlocksReached
	}

	// received an empty block or a block with new execution results
	log.Warn("selectionTracker.checkReceivedBlockNoLock: max tracked blocks reached "+
		"but received a tolerated block",
		"len(st.blocks)", len(st.blocks),
		"nonce", blockHeader.GetNonce())

	return nil
}

// validateTrackedBlocksAndCompileBreadcrumbsNoLock is used when a new block is proposed.
// Firstly, the method finds the chain of tracked blocks.
// Secondly, the method extracts the transaction of the new block, compiles its breadcrumbs and adds the new block to the previous returned chain.
// Then, it validates the entire chain (by nonce and balance of each breadcrumb).
func (st *selectionTracker) validateTrackedBlocksAndCompileBreadcrumbsNoLock(
	blockBody *block.Body,
	blockToTrack *trackedBlock,
	accountsProvider common.AccountNonceAndBalanceProvider,
	blockchainInfo common.BlockchainInfo,
) error {
	blocksToBeValidated, err := st.getChainOfTrackedPendingBlocks(
		blockchainInfo.GetLatestExecutedBlockHash(),
		blockToTrack.prevHash,
		blockToTrack.nonce,
	)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocksAndCompileBreadcrumbsNoLock: error creating chain of tracked blocks", "err", err)
		return err
	}

	// if we pass the first validation, only then we extract the txs to compile the breadcrumbs
	txs, err := getTransactionsInBlock(blockBody, st.txCache)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocksAndCompileBreadcrumbsNoLock: error getting transactions from block", "err", err)
		return err
	}

	err = blockToTrack.compileBreadcrumbs(txs)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocksAndCompileBreadcrumbsNoLock: error compiling breadcrumbs",
			"error", err)
		return err
	}

	// add the new block in the returned chain
	blocksToBeValidated = append(blocksToBeValidated, blockToTrack)

	// make sure that the breadcrumbs of the proposed block are valid
	// i.e. continuous with the other proposed blocks and no balance issues
	err = st.validateBreadcrumbsOfTrackedBlocks(blocksToBeValidated, accountsProvider)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocksAndCompileBreadcrumbsNoLock: error validating tracked blocks", "err", err)
		return err
	}

	return nil
}

// validateBreadcrumbsOfTrackedBlocks validates the breadcrumbs of each tracked block.
// Firstly, it checks for nonce continuity.
// Then, it checks that each account has enough balance.
func (st *selectionTracker) validateBreadcrumbsOfTrackedBlocks(
	chainOfTrackedBlocks []*trackedBlock,
	accountsProvider common.AccountNonceAndBalanceProvider,
) error {
	validator := newBreadcrumbValidator()

	for _, tb := range chainOfTrackedBlocks {
		for address, breadcrumb := range tb.breadcrumbsByAddress {
			initialNonce, initialBalance, _, err := accountsProvider.GetAccountNonceAndBalance([]byte(address))
			if err != nil {
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks",
					"err", err,
					"address", address,
					"tracked block rootHash", tb.rootHash,
					"tracked block hash", tb.hash,
					"tracked block nonce", tb.nonce)
				return err
			}

			if !validator.validateNonceContinuityOfBreadcrumb(address, initialNonce, breadcrumb) {
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks",
					"err", errDiscontinuousBreadcrumbs,
					"address", address,
					"tracked block rootHash", tb.rootHash,
					"tracked block hash", tb.hash,
					"tracked block nonce", tb.nonce)
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
					"tracked block rootHash", tb.rootHash,
					"tracked block hash", tb.hash,
					"tracked block nonce", tb.nonce)
				return err
			}
		}
	}

	return nil
}

// addNewTrackedBlockNoLock adds a new tracked block into the map of tracked blocks,
// replaces an existing block which has the same nonce with the one received.
func (st *selectionTracker) addNewTrackedBlockNoLock(blockToBeAddedHash []byte, blockToBeAdded *trackedBlock) error {
	// remove all the blocks with nonce equal or above the given nonce
	err := st.removeBlockEqualOrAboveNoLock(blockToBeAddedHash, blockToBeAdded)
	if err != nil {
		return err
	}

	// add the new block
	st.blocks[string(blockToBeAddedHash)] = blockToBeAdded
	st.globalBreadcrumbsCompiler.updateOnAddedBlock(blockToBeAdded)

	return nil
}

// removeBlockEqualOrAboveNoLock removes blocks higher or equal to the nonce of the given block.
// The removeBlockEqualOrAboveNoLock is used on the OnProposedBlock flow.
func (st *selectionTracker) removeBlockEqualOrAboveNoLock(blockToBeAddedHash []byte, blockToBeAdded *trackedBlock) error {
	// search if in the tracked blocks we already have one with same nonce or greater
	for bHash, b := range st.blocks {
		if b.hasSameNonceOrHigher(blockToBeAdded) {
			err := st.globalBreadcrumbsCompiler.updateOnRemovedBlockWithSameNonceOrAbove(b)
			if err != nil {
				return err
			}

			delete(st.blocks, bHash)

			log.Trace("selectionTracker.addNewTrackedBlockNoLock block with same nonce was deleted, to be replaced",
				"nonce", blockToBeAdded.nonce,
				"hash of replaced block", b.hash,
				"hash of new block", blockToBeAddedHash,
			)
		}
	}

	return nil
}

// OnExecutedBlock notifies when a block is executed and updates the state of the selectionTracker
// by removing each tracked block with nonce equal or lower than the one received in the blockHeader.
func (st *selectionTracker) OnExecutedBlock(blockHeader data.HeaderHandler) error {
	if check.IfNil(blockHeader) {
		return errNilBlockHeader
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

	err := st.removeUpToBlockNoLock(tempTrackedBlock)
	if err != nil {
		return err
	}

	st.updateLatestRootHashNoLock(nonce, rootHash)
	return nil
}

// removeUpToBlockNoLock removes all the tracked blocks with nonce equal or lower than the given nonce.
// The removeUpToBlockNoLock is called on the OnExecutedBlock flow.
func (st *selectionTracker) removeUpToBlockNoLock(searchedBlock *trackedBlock) error {
	removedBlocks := 0
	for blockHash, b := range st.blocks {
		if b.hasSameNonceOrLower(searchedBlock) {
			err := st.globalBreadcrumbsCompiler.updateOnRemovedBlockWithSameNonceOrBelow(b)
			if err != nil {
				return err
			}

			delete(st.blocks, blockHash)
			removedBlocks++
		}
	}

	log.Trace("selectionTracker.removeUpToBlockNoLock",
		"searched block nonce", searchedBlock.nonce,
		"searched block hash", searchedBlock.hash,
		"searched block rootHash", searchedBlock.rootHash,
		"searched block prevHash", searchedBlock.prevHash,
		"removed blocks", removedBlocks,
	)

	return nil
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

// ResetTrackedBlocks resets the tracked blocks, the global account breadcrumbs and the state saved on the OnExecutedBlock.
func (st *selectionTracker) ResetTrackedBlocks() {
	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	log.Debug("selectionTracker.ResetTrackedBlocks removing all tracked blocks",
		"len(trackedBlocks)", len(st.blocks),
	)

	st.latestRootHash = nil
	st.latestNonce = 0

	st.blocks = make(map[string]*trackedBlock)
	st.globalBreadcrumbsCompiler.cleanGlobalBreadcrumbs()
}

// deriveVirtualSelectionSession creates a virtual selection session by transforming the global accounts breadcrumbs into virtual records
// The deriveVirtualSelectionSession methods needs a SelectionSession and the nonce of the block on which the selection is built.
func (st *selectionTracker) deriveVirtualSelectionSession(
	session SelectionSession,
	nonce uint64,
	shouldRemoveTrackedBlocks bool,
) (*virtualSelectionSession, error) {
	if !shouldRemoveTrackedBlocks {
		err := st.removeBlocksAboveNonce(nonce)
		if err != nil {
			return nil, err
		}
	}

	rootHash, err := session.GetRootHash()
	if err != nil {
		log.Debug("selectionTracker.deriveVirtualSelectionSession",
			"err", err)
		return nil, err
	}

	log.Debug("selectionTracker.deriveVirtualSelectionSession",
		"rootHash", rootHash,
		"nonce", nonce,
	)

	st.displayTrackedBlocks(log, "trackedBlocks")

	computer := newVirtualSessionComputer(session)

	globalAccountsBreadcrumbs := st.getGlobalAccountsBreadcrumbs()
	return computer.createVirtualSelectionSession(globalAccountsBreadcrumbs)
}

// removeBlocksAboveNonce removes blocks with nonce higher than the given nonce.
// The removeBlocksAboveNonce is used on the deriveVirtualSelectionSession flow.
func (st *selectionTracker) removeBlocksAboveNonce(nonce uint64) error {
	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	for blockHash, tb := range st.blocks {
		if tb.hasHigherNonce(nonce) {
			err := st.globalBreadcrumbsCompiler.updateOnRemovedBlockWithSameNonceOrAbove(tb)
			if err != nil {
				return err
			}

			delete(st.blocks, blockHash)

			log.Trace("selectionTracker.removeBlocksAboveNonce",
				"nonce", nonce,
				"nonce of deleted block", tb.nonce,
				"hash of deleted block", blockHash,
			)
		}
	}

	return nil
}

// getChainOfTrackedPendingBlocks finds the chain of tracked blocks, iterating from tail to head,
// following the previous hash of each block, in order to avoid fork scenarios.
// The iteration stops when the previous hash of a block is equal to latestExecutedBlockHash.
func (st *selectionTracker) getChainOfTrackedPendingBlocks(
	latestExecutedBlockHash []byte,
	previousHashToBeFound []byte,
	nonceOfNextBlock uint64,
) ([]*trackedBlock, error) {
	chain := make([]*trackedBlock, 0)

	// If the previous hash to be found is equal to the latest executed hash,
	// it means that we do not have any tracked proposed block on top.
	// The block found would be the actual executed block, but that one is not tracked anymore.
	if bytes.Equal(latestExecutedBlockHash, previousHashToBeFound) {
		return chain, nil
	}

	// search for the block with the hash equal to the previous hash.
	// NOTE: we expect a nil value for a key (block hash) which is not in the map of tracked blocks.
	previousBlock := st.blocks[string(previousHashToBeFound)]

	for {
		if nonceOfNextBlock == 0 {
			// should never actually happen (e.g. genesis)
			break
		}

		// if no block was found, it means there is a gap and we have to return an error
		if previousBlock == nil {
			return nil, errBlockNotFound
		}

		// extra check for a block gap, to assure there are no missing tracked blocks
		hasDiscontinuousBlockNonce := previousBlock.nonce != nonceOfNextBlock-1
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
		nonceOfNextBlock -= 1

		// find the previous block
		previousBlock = st.blocks[string(previousBlockHash)]
	}

	// return the blocks in their natural order (from head to tail)
	slices.Reverse(chain)
	return chain, nil
}

func (st *selectionTracker) getGlobalAccountsBreadcrumbs() map[string]*globalAccountBreadcrumb {
	return st.globalBreadcrumbsCompiler.getGlobalBreadcrumbs()
}

// getVirtualNonceOfAccountWithRootHash searches the global breadcrumb of the given address and return its nonce
func (st *selectionTracker) getVirtualNonceOfAccountWithRootHash(
	address []byte,
) (uint64, []byte, error) {
	breadcrumb, err := st.globalBreadcrumbsCompiler.getGlobalBreadcrumbByAddress(string(address))
	if err != nil {
		return 0, nil, errGlobalBreadcrumbDoesNotExist
	}

	if !breadcrumb.lastNonce.HasValue {
		return 0, nil, errLastNonceNotFound
	}

	return breadcrumb.lastNonce.Value + 1, st.latestRootHash, nil
}

// GetBulkOfUntrackedTransactions returns the hashes of the untracked transactions
func (st *selectionTracker) GetBulkOfUntrackedTransactions(transactions []*WrappedTransaction) [][]byte {
	untrackedTransactions := make([][]byte, 0)
	for _, tx := range transactions {
		if tx == nil || tx.Tx == nil {
			continue
		}

		if !st.isTransactionTracked(tx) {
			untrackedTransactions = append(untrackedTransactions, tx.TxHash)
		}
	}

	return untrackedTransactions
}

// isTransactionTracked checks if a transaction is still in the tracked blocks of the SelectionTracker
// TODO analyze if the forks are still an issue here
func (st *selectionTracker) isTransactionTracked(transaction *WrappedTransaction) bool {
	if transaction == nil || transaction.Tx == nil {
		return false
	}

	sender := transaction.Tx.GetSndAddr()
	txNonce := transaction.Tx.GetNonce()

	senderGlobalBreadcrumb, err := st.globalBreadcrumbsCompiler.getGlobalBreadcrumbByAddress(string(sender))
	if err != nil {
		return false
	}

	minNonce := senderGlobalBreadcrumb.firstNonce
	maxNonce := senderGlobalBreadcrumb.lastNonce

	if !minNonce.HasValue || !maxNonce.HasValue {
		// we consider the transaction as not tracked because the account is tracked only as a relayer
		return false
	}

	if txNonce < minNonce.Value || txNonce > maxNonce.Value {
		// we consider the transaction as not tracked because it's outside the tracked range
		return false
	}

	return true
}

func (st *selectionTracker) displayTrackedBlocks(contextualLogger logger.Logger, linePrefix string) {
	if contextualLogger.GetLevel() > logger.LogTrace {
		return
	}

	st.mutTracker.RLock()
	defer st.mutTracker.RUnlock()

	log.Debug("selectionTracker.deriveVirtualSelectionSession",
		"len(trackedBlocks)", len(st.blocks))

	if len(st.blocks) > 0 {
		contextualLogger.Trace("displayTrackedBlocks - trackedBlocks (as newline-separated JSON):")
		contextualLogger.Trace(marshalTrackedBlockToNewlineDelimitedJSON(st.blocks, linePrefix))
	} else {
		contextualLogger.Trace("displayTrackedBlocks - trackedBlocks: none")
	}
}

// getNumTrackedBlocks returns the dimension of tracked blocks
// TODO the number of tracked accounts could also be returned when the globalAccountsBreadcrumbs is integrated
func (st *selectionTracker) getNumTrackedBlocks() uint64 {
	st.mutTracker.RLock()
	defer st.mutTracker.RUnlock()

	return uint64(len(st.blocks))
}
