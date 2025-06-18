package txcache

import (
	"bytes"
	"math/big"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type accountBreadcrumb struct {
	initialNonce    uint64
	lastNonce       uint64
	consumedBalance *big.Int
}

type trackedBlock struct {
	nonce                uint64
	hash                 []byte
	rootHash             []byte
	prevHash             []byte
	breadcrumbsByAddress map[string]*accountBreadcrumb
}

type selectionTracker struct {
	mutState          sync.RWMutex
	mutLatestRootHash sync.RWMutex
	latestNonce       uint64
	latestRootHash    []byte
	state             []*trackedBlock
	txCache           *TxCache
}

func newTrackedBlock(nonce uint64, blockHash []byte, rootHash []byte, prevHash []byte) *trackedBlock {
	return &trackedBlock{
		nonce:                nonce,
		hash:                 blockHash,
		rootHash:             rootHash,
		prevHash:             prevHash,
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	}
}

// NewSelectionTracker - creates a new selectionTracker
func NewSelectionTracker(txCache *TxCache) (*selectionTracker, error) {
	if txCache == nil {
		return nil, errNilTxCache
	}
	return &selectionTracker{
		mutState:          sync.RWMutex{},
		mutLatestRootHash: sync.RWMutex{},
		txCache:           txCache,
		state:             make([]*trackedBlock, 0),
	}, nil
}

// OnProposedBlock - notifies when a block is proposed and updates the state of the selectionTracker
func (st *selectionTracker) OnProposedBlock(blockHash []byte, blockBody *block.Body, handler data.HeaderHandler) error {
	if blockHash == nil {
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

	st.mutState.Lock()
	defer st.mutState.Unlock()

	logSelectionTracker.Info("selectionTracker.OnProposedBlock",
		"blockHash", blockHash,
		"nonce", nonce,
		"rootHash", rootHash,
		"prevHash", prevHash)

	st.state = append(st.state, newTrackedBlock(nonce, blockHash, rootHash, prevHash))
	return nil
}

// OnExecutedBlock - notifies when a block is executed and updates the state of the selectionTracker
func (st *selectionTracker) OnExecutedBlock(handler data.HeaderHandler) error {
	if check.IfNil(handler) {
		return errNilHeaderHandler
	}

	nonce := handler.GetNonce()
	rootHash := handler.GetRootHash()
	prevHash := handler.GetPrevHash()

	tempTrackedBlock := newTrackedBlock(nonce, rootHash, nil, prevHash)
	st.removeFromTrackedBlocks(tempTrackedBlock)
	st.updateLatestRootHash(nonce, rootHash)

	return nil
}

func (st *selectionTracker) removeFromTrackedBlocks(searchedBlock *trackedBlock) {
	st.mutState.Lock()
	defer st.mutState.Unlock()

	remainedBlocks := make([]*trackedBlock, 0)
	for i := 0; i < len(st.state); i++ {
		if !st.equalBlocks(searchedBlock, st.state[i]) {
			remainedBlocks = append(remainedBlocks, st.state[i])
		}
	}
	st.state = remainedBlocks
}

func (st *selectionTracker) updateLatestRootHash(receivedNonce uint64, receivedRootHash []byte) {
	st.mutLatestRootHash.Lock()
	defer st.mutLatestRootHash.Unlock()

	logSelectionTracker.Info("selectionTracker.updateLatestRootHash",
		"received root hash", receivedRootHash,
		"received nonce", receivedNonce)

	if st.latestRootHash == nil {
		st.latestRootHash = receivedRootHash
		st.latestNonce = receivedNonce
		return
	}
	if bytes.Equal(st.latestRootHash, receivedRootHash) {
		return
	}

	if receivedNonce > st.latestNonce {
		st.latestRootHash = receivedRootHash
		st.latestNonce = receivedNonce
	}
}

func (st *selectionTracker) equalBlocks(trackedBlock1, trackedBlock2 *trackedBlock) bool {
	if trackedBlock1.nonce != trackedBlock2.nonce || !bytes.Equal(trackedBlock1.prevHash, trackedBlock2.prevHash) {
		return false
	}

	return true
}
