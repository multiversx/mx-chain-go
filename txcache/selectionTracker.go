package txcache

import (
	"bytes"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"
	"golang.org/x/exp/slices"

	"github.com/multiversx/mx-chain-go/common"
)

const (
	maxAccountsPerBlock = 12000
)

type blockWithHash struct {
	hash  string
	block *trackedBlock
}

type selectionTracker struct {
	mutTracker                sync.RWMutex
	latestNonce               uint64 // last OnExecuted nonce
	latestRootHash            []byte // last OnExecuted rootHash
	blocks                    map[string]*trackedBlock
	globalBreadcrumbsCompiler *globalAccountBreadcrumbsCompiler
	txCache                   txCacheForSelectionTracker
	selfShardId               uint32
	maxTrackedBlocks          uint32
	aotSelectionPreempter     common.AOTSelectionPreempter
}

// NewSelectionTracker creates a new selectionTracker
func NewSelectionTracker(
	txCache txCacheForSelectionTracker,
	selfShardId uint32,
	maxTrackedBlocks uint32,
) (*selectionTracker, error) {
	if check.IfNil(txCache) {
		return nil, errNilTxCache
	}
	if maxTrackedBlocks == 0 {
		return nil, errInvalidMaxTrackedBlocks
	}
	return &selectionTracker{
		mutTracker:                sync.RWMutex{},
		blocks:                    make(map[string]*trackedBlock),
		globalBreadcrumbsCompiler: newGlobalAccountBreadcrumbsCompiler(),
		txCache:                   txCache,
		selfShardId:               selfShardId,
		maxTrackedBlocks:          maxTrackedBlocks,
	}, nil
}

// OnProposedBlock notifies when a block is proposed and updates the state of the selectionTracker.
// blockHash is the hash of the new proposed block.
// blockBody contains the transactions of the new block (required for creating the breadcrumbs and validating the block).
// blockHeader contains the nonce, the rootHash and the previousHash of the new proposed block.
// accountsProvider is a wrapper over the current blockchain state.
// latestExecutedHash represents the hash of the last executed block.
func (st *selectionTracker) OnProposedBlock(
	blockHash []byte,
	bodyHandler data.BodyHandler,
	blockHeader data.HeaderHandler,
	accountsProvider common.AccountNonceAndBalanceProvider,
	latestExecutedHash []byte,
) error {
	blockBody, ok := bodyHandler.(*block.Body)
	if !ok {
		return errWrongTypeAssertion
	}

	err := st.verifyBlockArgs(blockHash, blockBody, blockHeader)
	if err != nil {
		return err
	}
	if check.IfNil(accountsProvider) {
		return errNilAccountNonceAndBalanceProvider
	}

	accountsRootHash, err := accountsProvider.GetRootHash()
	if err != nil {
		return err
	}

	// Preempt any ongoing AOT selection before acquiring the lock
	st.cancelAOTOngoingSelection()

	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	err = st.checkUniqueAccountsLimit(blockBody)
	if err != nil {
		return err
	}

	if !bytes.Equal(st.latestRootHash, accountsRootHash) {
		log.Error("selectionTracker.OnProposedBlock",
			"err", errRootHashMismatch,
			"latestRootHash", st.latestRootHash,
			"accountsRootHash", accountsRootHash,
		)
		return errRootHashMismatch
	}

	nonce := blockHeader.GetNonce()
	prevHash := blockHeader.GetPrevHash()

	tBlock := newTrackedBlock(nonce, blockHash, prevHash)

	log.Debug("selectionTracker.OnProposedBlock",
		"nonce", nonce,
		"blockHash", blockHash,
		"prevHash", prevHash,
		"latestRootHash", st.latestRootHash,
	)

	err = st.checkReceivedBlockNoLock(blockBody, blockHeader)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error checking the received block", "err", err)
		return err
	}

	lastNoncePerSender, err := st.validateTrackedBlocksAndCompileBreadcrumbsNoLock(blockBody, tBlock, accountsProvider, latestExecutedHash)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error validating the tracked blocks", "err", err)
		return err
	}

	err = st.addNewTrackedBlockNoLock(blockHash, tBlock)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error adding the new tracked block", "err", err)
		return err
	}

	// Set selection offsets to skip transactions up to and including the last proposed nonce per sender
	// This skips already-proposed transactions during future selections
	st.txCache.SetSelectionOffsetsByLastNonce(lastNoncePerSender)

	return nil
}

// OnBackfilledBlock adds a previously consensus-agreed block as tracked without breadcrumb validation.
// This is used during sync start to backfill blocks between the last execution result and the current
// committed header. Since these blocks were already validated during consensus, breadcrumb continuity
// and balance checks are skipped. The block's breadcrumbs are still compiled and tracked.
func (st *selectionTracker) OnBackfilledBlock(
	blockHash []byte,
	bodyHandler data.BodyHandler,
	blockHeader data.HeaderHandler,
) error {
	blockBody, ok := bodyHandler.(*block.Body)
	if !ok {
		return errWrongTypeAssertion
	}

	err := st.verifyBlockArgs(blockHash, blockBody, blockHeader)
	if err != nil {
		return err
	}

	// Preempt any ongoing AOT selection before acquiring the lock
	st.cancelAOTOngoingSelection()

	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	nonce := blockHeader.GetNonce()
	prevHash := blockHeader.GetPrevHash()

	tBlock := newTrackedBlock(nonce, blockHash, prevHash)

	log.Debug("selectionTracker.OnBackfilledBlock",
		"nonce", nonce,
		"blockHash", blockHash,
		"prevHash", prevHash,
	)

	txs, err := getTransactionsInBlock(blockBody, st.txCache, st.selfShardId)
	if err != nil {
		return err
	}

	lastNoncePerSender, err := tBlock.compileBreadcrumbs(txs)
	if err != nil {
		return err
	}

	err = st.addNewTrackedBlockNoLock(blockHash, tBlock)
	if err != nil {
		return err
	}

	st.txCache.SetSelectionOffsetsByLastNonce(lastNoncePerSender)

	return nil
}

func (st *selectionTracker) verifyBlockArgs(
	blockHash []byte,
	blockBody *block.Body,
	blockHeader data.HeaderHandler,
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
// Returns lastNoncePerSender map for updating selection offsets.
func (st *selectionTracker) validateTrackedBlocksAndCompileBreadcrumbsNoLock(
	blockBody *block.Body,
	blockToTrack *trackedBlock,
	accountsProvider common.AccountNonceAndBalanceProvider,
	latestExecutedHash []byte,
) (map[string]uint64, error) {
	blocksToBeValidated, err := st.getChainOfTrackedPendingBlocks(
		latestExecutedHash,
		blockToTrack.prevHash,
		blockToTrack.nonce,
	)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocksAndCompileBreadcrumbsNoLock: error creating chain of tracked blocks", "err", err)
		return nil, err
	}

	// if we pass the first validation, only then we extract the txs to compile the breadcrumbs
	txs, err := getTransactionsInBlock(blockBody, st.txCache, st.selfShardId)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocksAndCompileBreadcrumbsNoLock: error getting transactions from block", "err", err)
		return nil, err
	}

	lastNoncePerSender, err := blockToTrack.compileBreadcrumbs(txs)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocksAndCompileBreadcrumbsNoLock: error compiling breadcrumbs",
			"error", err)
		return nil, err
	}

	// add the new block in the returned chain
	blocksToBeValidated = append(blocksToBeValidated, blockToTrack)

	// make sure that the breadcrumbs of the proposed block are valid
	// i.e. continuous with the other proposed blocks and no balance issues
	err = st.validateBreadcrumbsOfTrackedBlocks(blocksToBeValidated, accountsProvider)
	if err != nil {
		log.Debug("selectionTracker.validateTrackedBlocksAndCompileBreadcrumbsNoLock: error validating tracked blocks", "err", err)
		return nil, err
	}

	return lastNoncePerSender, nil
}

// validateBreadcrumbsOfTrackedBlocks validates the breadcrumbs of each tracked block.
// For predecessor blocks (all except the last), discontinuous breadcrumbs are tolerated:
// the address is marked as discontinuous and its nonce/balance validation is skipped.
// For the new block (the last one), if it has breadcrumbs for an address that was marked
// as discontinuous, validation fails - the new block must not include transactions
// for accounts with stale/discontinuous history.
func (st *selectionTracker) validateBreadcrumbsOfTrackedBlocks(
	chainOfTrackedBlocks []*trackedBlock,
	accountsProvider common.AccountNonceAndBalanceProvider,
) error {
	validator := newBreadcrumbValidator()

	numBlocks := len(chainOfTrackedBlocks)

	for i, tb := range chainOfTrackedBlocks {
		isNewBlock := i == numBlocks-1

		for address, breadcrumb := range tb.breadcrumbsByAddress {
			if isNewBlock && validator.isAddressDiscontinuous(address) {
				// The new block has transactions for an account with discontinuous breadcrumbs
				// in predecessor blocks. This must be rejected.
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks: new block has breadcrumb for discontinuous address",
					"err", errDiscontinuousBreadcrumbs,
					"address", address,
					"tracked block hash", tb.hash,
					"tracked block nonce", tb.nonce)
				return errDiscontinuousBreadcrumbs
			}

			// NOTE: the initial balance should never change during this validation, because we use an account provider which is not affected by the actual execution.
			initialNonce, initialBalance, _, err := accountsProvider.GetAccountNonceAndBalance([]byte(address))
			if err != nil {
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks",
					"err", err,
					"address", address,
					"tracked block hash", tb.hash,
					"tracked block nonce", tb.nonce)
				return err
			}

			if !validator.validateNonceContinuityOfBreadcrumb(address, initialNonce, breadcrumb) {
				if isNewBlock {
					// The new block itself has discontinuous breadcrumbs - reject
					log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks: new block has discontinuous breadcrumbs",
						"err", errDiscontinuousBreadcrumbs,
						"address", address,
						"tracked block hash", tb.hash,
						"tracked block nonce", tb.nonce)
					return errDiscontinuousBreadcrumbs
				}

				// Predecessor block has discontinuous breadcrumbs - tolerate, mark address
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks: tolerating discontinuous breadcrumbs in predecessor block",
					"address", address,
					"tracked block hash", tb.hash,
					"tracked block nonce", tb.nonce)
				validator.markAddressAsDiscontinuous(address)
				continue
			}

			// Skip balance validation for addresses already marked as discontinuous.
			// This is safe because predecessor blocks were already validated at their original proposal time.
			if validator.isAddressDiscontinuous(address) {
				continue
			}

			// use its balance to accumulate and validate (make sure is < than initialBalance from the session)
			err = validator.validateBalance(address, initialBalance, breadcrumb)
			if err != nil {
				// exit at the first failure
				log.Debug("selectionTracker.validateBreadcrumbsOfTrackedBlocks validation failed",
					"err", err,
					"address", address,
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
	if blockToBeAdded == nil {
		return errNilTrackedBlock
	}

	// remove all the blocks with nonce equal or above the given nonce
	err := st.removeBlocksAboveOrEqualToNonceNoLock(blockToBeAdded.nonce)
	if err != nil {
		return err
	}

	// add the new block
	st.blocks[string(blockToBeAddedHash)] = blockToBeAdded
	st.globalBreadcrumbsCompiler.updateOnAddedBlock(blockToBeAdded)

	return nil
}

// OnExecutedBlock notifies when a block is executed and updates the state of the selectionTracker
// by removing each tracked block with nonce equal or lower than the one received in the blockHeader.
func (st *selectionTracker) OnExecutedBlock(blockHeader data.HeaderHandler, rootHash []byte) error {
	if check.IfNil(blockHeader) {
		return errNilBlockHeader
	}

	nonce := blockHeader.GetNonce()
	prevHash := blockHeader.GetPrevHash()

	log.Debug("selectionTracker.OnExecutedBlock",
		"nonce", nonce,
		"rootHash", rootHash,
		"prevHash", prevHash,
	)

	tempTrackedBlock := newTrackedBlock(nonce, nil, prevHash)

	// Preempt any ongoing AOT selection before acquiring the lock
	st.cancelAOTOngoingSelection()

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
// Blocks are sorted by ascending nonce before processing to ensure correct global breadcrumb updates.
func (st *selectionTracker) removeUpToBlockNoLock(searchedBlock *trackedBlock) error {
	// Collect matching blocks first
	var matchingBlocks []blockWithHash
	for blockHash, b := range st.blocks {
		if b.hasSameNonceOrLower(searchedBlock) {
			matchingBlocks = append(matchingBlocks, blockWithHash{hash: blockHash, block: b})
		}
	}

	// Sort by ascending nonce to ensure correct global breadcrumb updates
	sort.Slice(matchingBlocks, func(i, j int) bool {
		return matchingBlocks[i].block.nonce < matchingBlocks[j].block.nonce
	})

	// Process in sorted order
	for _, mb := range matchingBlocks {
		// first delete, then update the global breadcrumbs
		delete(st.blocks, mb.hash)

		err := st.globalBreadcrumbsCompiler.updateOnRemovedBlockWithSameNonceOrBelow(mb.block)
		if err != nil {
			return err
		}
	}

	log.Trace("selectionTracker.removeUpToBlockNoLock",
		"searched block nonce", searchedBlock.nonce,
		"searched block hash", searchedBlock.hash,
		"searched block prevHash", searchedBlock.prevHash,
		"removed blocks", len(matchingBlocks),
	)

	return nil
}

func (st *selectionTracker) updateLatestRootHashNoLock(receivedNonce uint64, receivedRootHash []byte) {
	log.Debug("selectionTracker.updateLatestRootHashNoLock",
		"latest root hash", st.latestRootHash,
		"received root hash", receivedRootHash,
		"latest nonce", st.latestNonce,
		"received nonce", receivedNonce,
	)

	if st.latestRootHash == nil {
		st.latestRootHash = receivedRootHash
		st.latestNonce = receivedNonce
		return
	}

	if st.latestNonce >= receivedNonce {
		log.Debug("selectionTracker.updateLatestRootHashNoLock received a lower or equal nonce than the latest nonce")
		return
	}

	st.latestRootHash = receivedRootHash
	st.latestNonce = receivedNonce
}

func (st *selectionTracker) checkUniqueAccountsLimit(blockBody *block.Body) error {
	txsInBlock, err := getTransactionsInBlock(blockBody, st.txCache, st.selfShardId)
	if err != nil {
		return nil
	}

	uniqueAccounts := make(map[string]struct{})
	for _, tx := range txsInBlock {
		uniqueAccounts[string(tx.Tx.GetSndAddr())] = struct{}{}
		if len(uniqueAccounts) > maxAccountsPerBlock {
			log.Warn("selectionTracker.OnProposedBlock: too many unique accounts in block",
				"count", len(uniqueAccounts),
				"limit", maxAccountsPerBlock)
			return errToManyUniqueAccountsInBlock
		}
	}

	return nil
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

func (st *selectionTracker) canDoSimulateSelection(nonce uint64) bool {
	// nonce 0 will select over current tracker state
	if nonce == 0 {
		return true
	}

	lastTrackedBlockNonce := st.latestNonce
	// drop the selection if not matching the tracker state
	for _, tb := range st.blocks {
		if tb.nonce > lastTrackedBlockNonce {
			lastTrackedBlockNonce = tb.nonce
		}
	}

	return nonce == lastTrackedBlockNonce+1
}

// deriveVirtualSelectionSession creates a virtual selection session by transforming the global accounts breadcrumbs into virtual records
// The deriveVirtualSelectionSession methods needs a SelectionSession and the nonce of the block for which the selection is built.
// Before the actual selection, all tracked blocks with greater or equal nonce are removed from the tracker.
func (st *selectionTracker) deriveVirtualSelectionSession(
	session SelectionSession,
	nonce uint64,
	isSimulation bool,
) (*virtualSelectionSession, error) {
	st.mutTracker.Lock()
	defer st.mutTracker.Unlock()

	if !isSimulation {
		err := st.removeBlocksAboveOrEqualToNonceNoLock(nonce)
		if err != nil {
			return nil, err
		}
	} else if !st.canDoSimulateSelection(nonce) {
		return nil, errSimulateSelectionContextInvalid
	}

	rootHash, err := session.GetRootHash()
	if err != nil {
		log.Debug("selectionTracker.deriveVirtualSelectionSession",
			"err", err)
		return nil, err
	}

	if !bytes.Equal(st.latestRootHash, rootHash) {
		log.Error("selectionTracker.deriveVirtualSelectionSession",
			"err", errRootHashMismatch,
			"latestRootHash", st.latestRootHash,
			"session rootHash", rootHash,
		)
		return nil, errRootHashMismatch
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

// removeBlocksAboveOrEqualToNonceNoLock removes blocks with nonce higher or equal than the given nonce.
// The removeBlocksAboveOrEqualToNonceNoLock is used on the deriveVirtualSelectionSession flow.
// Blocks are sorted by descending nonce before processing to ensure correct global breadcrumb updates.
func (st *selectionTracker) removeBlocksAboveOrEqualToNonceNoLock(nonce uint64) error {
	// Collect matching blocks first
	var matchingBlocks []blockWithHash
	for blockHash, tb := range st.blocks {
		if tb.hasSameNonceOrHigherThanGivenNonce(nonce) {
			matchingBlocks = append(matchingBlocks, blockWithHash{hash: blockHash, block: tb})
		}
	}

	log.Trace("selectionTracker.removeBlocksAboveOrEqualToNonceNoLock",
		"nonce", nonce,
		"num blocks to remove", len(matchingBlocks),
	)

	return st.deleteMatchedBlocksWithSameNonceOrAboveNoLock(matchingBlocks)
}

// deleteMatchedBlocksWithSameNonceOrAboveNoLock sorts the matched blocks by descending nonce,
// deletes them from the tracked blocks map, updates the global breadcrumbs, and resets selection offsets.
// The descending sort order ensures global breadcrumbs are reduced correctly (highest nonce first).
func (st *selectionTracker) deleteMatchedBlocksWithSameNonceOrAboveNoLock(matchingBlocks []blockWithHash) error {
	// Sort by descending nonce to ensure correct global breadcrumb updates
	sort.Slice(matchingBlocks, func(i, j int) bool {
		return matchingBlocks[i].block.nonce > matchingBlocks[j].block.nonce
	})

	sendersWithFirstNonce := make(map[string]uint64)

	for _, mb := range matchingBlocks {
		// Collect senders and their first nonce from removed blocks for offset reset
		for address, breadcrumb := range mb.block.breadcrumbsByAddress {
			if breadcrumb.firstNonce.HasValue {
				// Keep the lowest first nonce for each sender across all removed blocks
				if existingNonce, exists := sendersWithFirstNonce[address]; !exists || breadcrumb.firstNonce.Value < existingNonce {
					sendersWithFirstNonce[address] = breadcrumb.firstNonce.Value
				}
			}
		}

		// first delete, then update the global breadcrumbs
		delete(st.blocks, mb.hash)

		err := st.globalBreadcrumbsCompiler.updateOnRemovedBlockWithSameNonceOrAbove(mb.block)
		if err != nil {
			return err
		}

		log.Trace("selectionTracker.deleteMatchedBlocksWithSameNonceOrAboveNoLock",
			"nonce of deleted block", mb.block.nonce,
			"hash of deleted block", mb.hash,
		)
	}

	// Reset selection offsets for affected senders so their transactions can be re-selected
	if len(sendersWithFirstNonce) > 0 {
		st.txCache.ResetSelectionOffsetsToNonce(sendersWithFirstNonce)
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
			log.Debug("getChainOfTrackedPendingBlocks: hash not found",
				"previousHashToBeFound", previousHashToBeFound,
			)
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

// IsTransactionTracked checks if a transaction is still in the tracked blocks of the SelectionTracker.
// However, in the case of forks, IsTransactionTracked might return inaccurate results.
func (st *selectionTracker) IsTransactionTracked(transaction *WrappedTransaction) bool {
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

// should be called under mutex protection
func (st *selectionTracker) displayTrackedBlocks(contextualLogger logger.Logger, linePrefix string) {
	if contextualLogger.GetLevel() > logger.LogTrace {
		return
	}

	log.Debug("selectionTracker.deriveVirtualSelectionSession",
		"len(trackedBlocks)", len(st.blocks))

	if len(st.blocks) > 0 {
		contextualLogger.Trace("displayTrackedBlocks - trackedBlocks (as newline-separated JSON):")
		contextualLogger.Trace(marshalTrackedBlockToNewlineDelimitedJSON(st.blocks, linePrefix))
	} else {
		contextualLogger.Trace("displayTrackedBlocks - trackedBlocks: none")
	}
}

// getTrackerDiagnosis returns the dimension of tracked blocks and the number of global account breadcrumbs
func (st *selectionTracker) getTrackerDiagnosis() TrackerDiagnosis {
	st.mutTracker.RLock()
	defer st.mutTracker.RUnlock()

	return NewTrackerDiagnosis(uint64(len(st.blocks)), st.globalBreadcrumbsCompiler.getNumGlobalBreadcrumbs())
}

// SetAOTSelectionPreempter sets the AOT selection preempter for preemption support
func (st *selectionTracker) SetAOTSelectionPreempter(preempter common.AOTSelectionPreempter) {
	st.aotSelectionPreempter = preempter
}

// cancelAOTOngoingSelection cancels any ongoing AOT selection before critical operations
func (st *selectionTracker) cancelAOTOngoingSelection() {
	if !check.IfNil(st.aotSelectionPreempter) {
		st.aotSelectionPreempter.CancelOngoingSelection()
	}
}
