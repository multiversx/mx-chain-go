package txcache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state"
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

func (st *selectionTracker) createVirtualSelectionSession(session SelectionSession,
	chainOfTrackedBlocks []*trackedBlock) (*virtualSelectionSession, error) {
	virtualAccountsByRecords := make(map[string]*virtualAccountRecord)

	for _, tb := range chainOfTrackedBlocks {
		for address, breadcrumb := range tb.breadcrumbsByAddress {
			accountState, err := session.GetAccountState([]byte(address))
			if err != nil {
				log.Debug("selectionTracker.createVirtualSelectionSession",
					"err", err)
				return nil, err
			}

			st.fromBreadcrumbToVirtualRecord(virtualAccountsByRecords, accountState, address, breadcrumb)
		}
	}

	return &virtualSelectionSession{
		session:                  session,
		virtualAccountsByAddress: virtualAccountsByRecords,
	}, nil
}

func (st *selectionTracker) fromBreadcrumbToVirtualRecord(virtualAccountsByRecords map[string]*virtualAccountRecord,
	accountState state.UserAccountHandler, address string, breadcrumb *accountBreadcrumb) {

	virtualRecord, ok := virtualAccountsByRecords[address]
	if !ok {
		virtualRecord = st.createVirtualRecord(accountState, breadcrumb)
	}

	st.updateVirtualRecord(virtualRecord, breadcrumb)
}

func (st *selectionTracker) createVirtualRecord(accountState state.UserAccountHandler,
	breadcrumb *accountBreadcrumb) *virtualAccountRecord {
	initialBalance := accountState.GetBalance()
	virtualRecord := &virtualAccountRecord{
		initialNonce:   breadcrumb.initialNonce,
		initialBalance: initialBalance,
	}

	return virtualRecord
}

func (st *selectionTracker) updateVirtualRecord(virtualRecord *virtualAccountRecord,
	breadcrumb *accountBreadcrumb) {
	_ = virtualRecord.initialBalance.Add(virtualRecord.initialBalance, breadcrumb.consumedBalance)

	if !virtualRecord.initialNonce.HasValue {
		virtualRecord.initialNonce = breadcrumb.initialNonce
		return
	}

	if breadcrumb.initialNonce.HasValue {
		virtualRecord.initialNonce = core.OptionalUint64{
			Value:    max(breadcrumb.initialNonce.Value, virtualRecord.initialNonce.Value),
			HasValue: true,
		}
	}
}
