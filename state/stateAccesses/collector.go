package stateAccesses

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
)

var log = logger.GetOrCreate("state/stateAccesses")

const maxNumBlocksInMemory = 20

type stateAccessesForTxs map[string]*data.StateAccesses

type committedStateAccesses struct {
	rootHash []byte
	accesses stateAccessesForTxs
}

type collector struct {
	collectRead            bool
	collectWrite           bool
	withAccountChanges     bool
	stateAccesses          []*data.StateAccess
	stateAccessesForTxs    stateAccessesForTxs
	stateAccessesForHeader map[string]*committedStateAccesses
	retainedMut            sync.RWMutex
	headerHash             []byte
	headerGeneration       uint64
	headerScopeMut         sync.RWMutex

	storer           state.StateAccessesStorer
	stateAccessesMut sync.RWMutex
}

// NewCollector will create a new collector which gathers the state accesses based on the provided options.
func NewCollector(storer state.StateAccessesStorer, opts ...CollectorOption) (*collector, error) {
	if check.IfNil(storer) {
		return nil, state.ErrNilStateAccessesStorer
	}

	c := &collector{
		stateAccesses:          make([]*data.StateAccess, 0),
		stateAccessesForTxs:    make(map[string]*data.StateAccesses),
		stateAccessesForHeader: make(map[string]*committedStateAccesses),
	}
	for _, opt := range opts {
		opt(c)
	}

	c.storer = storer
	log.Debug("created new state accesses collector",
		"withRead", c.collectRead,
		"withWrite", c.collectWrite,
		"withAccountChanges", c.withAccountChanges,
	)

	return c, nil
}

// AddStateAccess adds a new state access to the collector
func (c *collector) AddStateAccess(stateAccess *data.StateAccess) {
	c.stateAccessesMut.Lock()
	defer c.stateAccessesMut.Unlock()

	if stateAccess.GetType() == data.Write && c.collectWrite {
		c.stateAccesses = append(c.stateAccesses, stateAccess)
	}

	if stateAccess.GetType() == data.Read && c.collectRead {
		c.stateAccesses = append(c.stateAccesses, stateAccess)
	}
}

// GetAccountChanges will return the account changes
func (c *collector) GetAccountChanges(oldAccount, account vmcommon.AccountHandler) uint32 {
	if c.withAccountChanges {
		return getAccountChanges(oldAccount, account)
	}

	return data.NoChange
}

// Reset resets the state accesses collector
func (c *collector) Reset() {
	c.stateAccessesMut.Lock()
	c.stateAccesses = make([]*data.StateAccess, 0)
	c.stateAccessesForTxs = make(map[string]*data.StateAccesses)
	c.stateAccessesMut.Unlock()

	c.retainedMut.RLock()
	if len(c.stateAccessesForHeader) > maxNumBlocksInMemory {
		log.Warn("max number of executions in memory exceeded", "numExecutionsInMemory", len(c.stateAccessesForHeader))
	}
	log.Trace("reset state accesses collector", "num executions", len(c.stateAccessesForHeader))
	c.retainedMut.RUnlock()
}

// BeginExecution sets the identity for the next state commit
func (c *collector) BeginExecution(headerHash []byte) uint64 {
	c.headerScopeMut.Lock()
	c.headerGeneration++
	if c.headerGeneration == 0 {
		c.headerGeneration++
	}
	generation := c.headerGeneration
	if len(c.headerHash) > 0 {
		log.Warn("replacing active state accesses header scope", "headerHash", c.headerHash)
	}
	c.headerHash = append(c.headerHash[:0], headerHash...)
	c.headerScopeMut.Unlock()

	return generation
}

// EndExecution clears the identity when the generation is still current
func (c *collector) EndExecution(generation uint64) {
	c.headerScopeMut.Lock()
	if c.headerGeneration == generation {
		c.headerHash = nil
	}
	c.headerScopeMut.Unlock()
}

// TakeStateAccessesForHeader atomically validates and removes retained state accesses
func (c *collector) TakeStateAccessesForHeader(headerHash, expectedRootHash []byte) (map[string]*data.StateAccesses, error) {
	c.retainedMut.Lock()
	retained, ok := c.stateAccessesForHeader[string(headerHash)]
	if !ok {
		c.retainedMut.Unlock()
		return nil, nil
	}
	delete(c.stateAccessesForHeader, string(headerHash))
	c.retainedMut.Unlock()

	if !bytes.Equal(retained.rootHash, expectedRootHash) {
		return nil, &state.StateAccessesRootMismatchError{
			HeaderHash:   append([]byte(nil), headerHash...),
			ExpectedRoot: append([]byte(nil), expectedRootHash...),
			ActualRoot:   append([]byte(nil), retained.rootHash...),
		}
	}

	return retained.accesses, nil
}

// DiscardStateAccessesForHeader removes retained accesses for one header
func (c *collector) DiscardStateAccessesForHeader(headerHash []byte) {
	c.retainedMut.Lock()
	delete(c.stateAccessesForHeader, string(headerHash))
	c.retainedMut.Unlock()
}

func (c *collector) getStateAccessesForTxs() map[string]*data.StateAccesses {
	if len(c.stateAccessesForTxs) != 0 && len(c.stateAccesses) == 0 {
		return c.stateAccessesForTxs
	}
	if len(c.stateAccessesForTxs) != 0 && len(c.stateAccesses) != 0 {
		log.Warn("state accesses already collected, but new state accesses were added after")
	}

	stateAccessWithNoAssociatedTx := 0
	for _, stateAccess := range c.stateAccesses {
		txHash := string(stateAccess.GetTxHash())
		if len(txHash) == 0 {
			stateAccessWithNoAssociatedTx++
			log.Trace("empty tx hash, state access event not associated to a transaction", "stateChange", stateAccessToString(stateAccess))
			continue
		}

		_, ok := c.stateAccessesForTxs[txHash]
		if !ok {
			c.stateAccessesForTxs[txHash] = &data.StateAccesses{
				StateAccess: []*data.StateAccess{stateAccess},
			}

			continue
		}

		c.mergeStateAccessesIfSameAccount(txHash, stateAccess)
	}

	if stateAccessWithNoAssociatedTx > 0 {
		log.Warn("state accesses with no associated tx; use state:TRACE for more data", "num", stateAccessWithNoAssociatedTx)
	}
	c.stateAccesses = make([]*data.StateAccess, 0)
	logCollectedStateAccesses("state accesses for txs", c.stateAccessesForTxs)

	return c.stateAccessesForTxs
}

func (c *collector) mergeStateAccessesIfSameAccount(txHash string, stateAccess *data.StateAccess) {
	stateAccesses := c.stateAccessesForTxs[txHash].StateAccess
	wasMerged := false
	for i := range stateAccesses {
		sameAccount := bytes.Equal(stateAccesses[i].MainTrieKey, stateAccess.MainTrieKey)
		if !sameAccount {
			continue
		}
		sameActionType := stateAccesses[i].Type == stateAccess.Type
		if !sameActionType {
			continue
		}
		wasMerged = true
		stateAccesses[i].MainTrieVal = stateAccess.MainTrieVal
		stateAccesses[i].Operation = data.MergeOperations(stateAccesses[i].Operation, stateAccess.Operation)
		stateAccesses[i].DataTrieChanges = data.MergeDataTrieChanges(stateAccesses[i].DataTrieChanges, stateAccess.DataTrieChanges)
		if c.withAccountChanges {
			stateAccesses[i].AccountChanges = stateAccesses[i].AccountChanges | stateAccess.AccountChanges
		}
	}
	if !wasMerged {
		c.stateAccessesForTxs[txHash].StateAccess = append(c.stateAccessesForTxs[txHash].StateAccess, stateAccess)
	}
}

func logCollectedStateAccesses(message string, stateAccessesForTx map[string]*data.StateAccesses) {
	if log.GetLevel() != logger.LogTrace {
		return
	}

	log.Trace(message, "numTxs", len(stateAccessesForTx))
	for txHash, stateAccesses := range stateAccessesForTx {
		log.Trace("state accesses for tx",
			"txHash", txHash,
			"numStateAccesses", len(stateAccesses.StateAccess),
		)
		for _, stateAccess := range stateAccesses.StateAccess {
			log.Trace("state access", "stateAccess", stateAccessToString(stateAccess))
		}
	}
}

func stateAccessToString(stateAccess *data.StateAccess) string {
	dataTrieChanges := make([]string, len(stateAccess.GetDataTrieChanges()))
	for i, dataTrieChange := range stateAccess.GetDataTrieChanges() {
		dataTrieChanges[i] = fmt.Sprintf("key: %v, val: %v, type: %v, operation: %v, version: %v", hex.EncodeToString(dataTrieChange.Key), hex.EncodeToString(dataTrieChange.Val), dataTrieChange.Type, dataTrieChange.Operation, dataTrieChange.Version)
	}
	return fmt.Sprintf("type: %v, operation: %v, mainTrieKey: %v, mainTrieVal: %v, index: %v, dataTrieChanges: %v, accountChanges: %v", stateAccess.GetType(),
		stateAccess.GetOperation(),
		hex.EncodeToString(stateAccess.GetMainTrieKey()),
		hex.EncodeToString(stateAccess.GetMainTrieVal()),
		stateAccess.GetIndex(),
		strings.Join(dataTrieChanges, ", "),
		stateAccess.GetAccountChanges(),
	)
}

// CommitCollectedAccesses will call the Store method of the underlying storer, giving it the collected state accesses.
func (c *collector) CommitCollectedAccesses(rootHash []byte) error {
	c.stateAccessesMut.Lock()
	collectedStateAccesses := c.getStateAccessesForTxs()
	if len(collectedStateAccesses) == 0 {
		c.stateAccessesMut.Unlock()
		log.Trace("no state accesses collected, skipping commit", "rootHash", rootHash)
		return nil
	}
	c.stateAccessesForTxs = make(map[string]*data.StateAccesses)
	c.stateAccesses = make([]*data.StateAccess, 0)
	c.stateAccessesMut.Unlock()

	c.headerScopeMut.RLock()
	headerHash := append([]byte(nil), c.headerHash...)
	c.headerScopeMut.RUnlock()

	if len(headerHash) > 0 {
		c.retainedMut.RLock()
		retained := c.stateAccessesForHeader[string(headerHash)]
		c.retainedMut.RUnlock()
		if retained != nil {
			if bytes.Equal(retained.rootHash, rootHash) {
				return nil
			}

			return &state.StateAccessesRootMismatchError{
				HeaderHash:   headerHash,
				ExpectedRoot: append([]byte(nil), retained.rootHash...),
				ActualRoot:   append([]byte(nil), rootHash...),
			}
		}
	}

	err := c.storer.Store(collectedStateAccesses)
	if err != nil {
		return err
	}

	if len(headerHash) == 0 {
		return nil
	}

	c.retainedMut.Lock()
	c.stateAccessesForHeader[string(headerHash)] = &committedStateAccesses{
		rootHash: append([]byte(nil), rootHash...),
		accesses: collectedStateAccesses,
	}
	numExecutions := len(c.stateAccessesForHeader)
	c.retainedMut.Unlock()

	if numExecutions > maxNumBlocksInMemory {
		log.Warn("max number of executions in memory exceeded", "numExecutionsInMemory", numExecutions)
	}

	return nil
}

// AddTxHashToCollectedStateAccesses will try to set txHash field to each state access
// if the field is not already set
func (c *collector) AddTxHashToCollectedStateAccesses(txHash []byte) {
	c.stateAccessesMut.Lock()
	defer c.stateAccessesMut.Unlock()

	for i := len(c.stateAccesses) - 1; i >= 0; i-- {
		if len(c.stateAccesses[i].GetTxHash()) > 0 {
			break
		}

		log.Trace("added tx hash to state access", "txHash", txHash, "index", c.stateAccesses[i].GetIndex())
		c.stateAccesses[i].TxHash = txHash
	}
}

// SetIndexToLatestStateAccesses will set the index to the latest collected state accesses
func (c *collector) SetIndexToLatestStateAccesses(index int) error {
	c.stateAccessesMut.Lock()
	defer c.stateAccessesMut.Unlock()

	if index < 0 {
		return fmt.Errorf("SetIndexToLatestStateAccesses: %w for index %v, num state accesses %v", state.ErrStateAccessesIndexOutOfBounds, index, len(c.stateAccesses))
	}

	if len(c.stateAccesses) == 0 {
		return nil
	}

	for i := len(c.stateAccesses) - 1; i >= 0; i-- {
		if c.stateAccesses[i].GetIndex() != 0 {
			return nil
		}
		log.Trace("set index to last state access", "stateChange num", i, "index", index)
		c.stateAccesses[i].Index = int32(index)
	}

	return nil
}

// RevertToIndex will revert to index
func (c *collector) RevertToIndex(index int) error {
	if index < 0 {
		return fmt.Errorf("RevertToIndex: %w for index %v, num state accesses %v", state.ErrStateAccessesIndexOutOfBounds, index, len(c.stateAccesses))
	}

	if index == 0 {
		c.Reset()
		return nil
	}

	c.stateAccessesMut.Lock()
	defer c.stateAccessesMut.Unlock()

	log.Trace("num state accesses before revert", "num", len(c.stateAccesses))
	for i := len(c.stateAccesses) - 1; i >= 0; i-- {
		if c.stateAccesses[i].GetIndex() == int32(index) {
			c.stateAccesses = c.stateAccesses[:i+1]
			log.Trace("reverted to index", "index", index, "num state accesses after revert", len(c.stateAccesses))
			break
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *collector) IsInterfaceNil() bool {
	return c == nil
}
