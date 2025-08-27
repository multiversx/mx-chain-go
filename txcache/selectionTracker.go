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
// TODO add an upper bound MaxTrackedBlocks
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
// TODO the selection session might be unusable in the flow of OnProposed
// TODO log in case MaxTrackedBlocks is reached and brainstorm how to solve this case
func (st *selectionTracker) OnProposedBlock(
	blockHash []byte,
	blockBody *block.Body,
	handler data.HeaderHandler,
	session SelectionSession,
	blockchainInfo common.BlockchainInfo,
) error {
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

	// TODO brainstorm if this could be moved after getChainOfTrackedBlocks
	txs, err := st.getTransactionsFromBlock(blockBody)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error getting transactions from block", "err", err)
		return err
	}

	tBlock, err := newTrackedBlock(nonce, blockHash, rootHash, prevHash, txs)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error creating tracked block", "err", err)
		return err
	}

	blocksToBeValidated, err := st.getChainOfTrackedBlocks(
		blockchainInfo.GetLatestExecutedBlockHash(),
		prevHash,
		nonce,
	)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error creating chain of tracked blocks", "err", err)
		return err
	}

	// add the new block in the chain
	blocksToBeValidated = append(blocksToBeValidated, tBlock)

	// make sure that the proposed block is valid (continuous with the other proposed blocks and no balance issues)
	err = st.validateTrackedBlocks(blocksToBeValidated, session)
	if err != nil {
		log.Debug("selectionTracker.OnProposedBlock: error validating tracked blocks", "err", err)
		return err
	}

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

func (st *selectionTracker) validateTrackedBlocks(chainOfTrackedBlocks []*trackedBlock, session SelectionSession) error {
	validator := newBreadcrumbValidator()

	for _, tb := range chainOfTrackedBlocks {
		for address, breadcrumb := range tb.breadcrumbsByAddress {
			initialNonce, initialBalance, _, err := session.GetAccountNonceAndBalance([]byte(address))
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

func (st *selectionTracker) getTransactionsFromBlock(blockBody *block.Body) ([]*WrappedTransaction, error) {
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
	previousBlock := st.findBlockInChainByPreviousHash(previousHashToBeFound)

	for {
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
		previousBlock = st.findBlockInChainByPreviousHash(previousBlockHash)
	}

	// to be able to validate the blocks later, reverse the order of the blocks to have them from head to tail
	return st.reverseOrderOfBlocks(chain), nil
}

// findBlockInChainByPreviousHash finds the block which has the hash equal to the given previous hash
func (st *selectionTracker) findBlockInChainByPreviousHash(previousHash []byte) *trackedBlock {
	for _, b := range st.blocks {
		if bytes.Equal(b.hash, previousHash) {
			return b
		}
	}

	return nil
}

func (st *selectionTracker) reverseOrderOfBlocks(chainOfTrackedBlocks []*trackedBlock) []*trackedBlock {
	reversedChainOfTrackedBlocks := make([]*trackedBlock, 0, len(chainOfTrackedBlocks))
	for i := len(chainOfTrackedBlocks) - 1; i >= 0; i-- {
		reversedChainOfTrackedBlocks = append(reversedChainOfTrackedBlocks, chainOfTrackedBlocks[i])
	}

	return reversedChainOfTrackedBlocks
}
