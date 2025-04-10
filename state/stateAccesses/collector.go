package stateAccesses

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
)

var log = logger.GetOrCreate("state/stateAccesses")

type collector struct {
	collectRead     bool
	collectWrite    bool
	stateAccesses   []*data.StateAccess
	stateChangesMut sync.RWMutex
	storer          storage.Persister
	marshaller      marshal.Marshalizer
}

// NewCollector will collect based on the options the state changes.
func NewCollector(marshaller marshal.Marshalizer, opts ...CollectorOption) (*collector, error) {
	if check.IfNil(marshaller) {
		return nil, state.ErrNilMarshalizer
	}

	c := &collector{stateAccesses: make([]*data.StateAccess, 0)}
	for _, opt := range opts {
		opt(c)
	}

	c.marshaller = marshaller
	log.Debug("created new state changes collector",
		"withRead", c.collectRead,
		"withWrite", c.collectWrite,
		"withStorer", c.storer != nil,
	)

	return c, nil
}

// AddStateAccess adds a new state access to the collector
func (c *collector) AddStateAccess(stateAccess *data.StateAccess) {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	if stateAccess.GetType() == data.Write && c.collectWrite {
		c.stateAccesses = append(c.stateAccesses, stateAccess)
	}

	if stateAccess.GetType() == data.Read && c.collectRead {
		c.stateAccesses = append(c.stateAccesses, stateAccess)
	}
}

// AddSaveAccountStateAccess adds a new state access for the save account operation
func (c *collector) AddSaveAccountStateAccess(oldAccount, account vmcommon.AccountHandler, stateAccess *data.StateAccess) {
	if c.storer != nil {
		stateAccess.AccountChanges = getAccountChanges(oldAccount, account)
	}

	c.AddStateAccess(stateAccess)
}

// Reset resets the state changes collector
func (c *collector) Reset() {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	c.stateAccesses = make([]*data.StateAccess, 0)
	log.Trace("reset state changes collector")
}

// Publish will export state changes
func (c *collector) Publish() (map[string]*data.StateAccesses, error) {
	c.stateChangesMut.RLock()
	defer c.stateChangesMut.RUnlock()

	stateAccessesForTxs := getStateAccessesForTxs(c.stateAccesses, false)

	logCollectedStateAccesses("state accesses to publish", stateAccessesForTxs)
	return stateAccessesForTxs, nil
}

func getStateAccessesForTxs(collectedStateAccesses []*data.StateAccess, withAccountChanges bool) map[string]*data.StateAccesses {
	stateAccessesForTx := make(map[string]*data.StateAccesses)
	for _, stateAccess := range collectedStateAccesses {
		txHash := string(stateAccess.GetTxHash())
		if len(txHash) == 0 {
			log.Warn("empty tx hash, state change event not associated to a transaction", "stateChange", stateAccessToString(stateAccess))
			continue
		}

		if !withAccountChanges {
			stateAccess.AccountChanges = nil
		}

		_, ok := stateAccessesForTx[txHash]
		if !ok {
			stateAccessesForTx[txHash] = &data.StateAccesses{
				StateAccess: []*data.StateAccess{stateAccess},
			}

			continue
		}

		stateAccessesForTx[txHash].StateAccess = append(stateAccessesForTx[txHash].StateAccess, stateAccess)
	}

	return stateAccessesForTx
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
		dataTrieChanges[i] = fmt.Sprintf("key: %v, val: %v, type: %v", hex.EncodeToString(dataTrieChange.Key), hex.EncodeToString(dataTrieChange.Val), dataTrieChange.Type)
	}
	return fmt.Sprintf("type: %v, operation: %v, mainTrieKey: %v, mainTrieVal: %v, index: %v, dataTrieChanges: %v",
		stateAccess.GetType(),
		stateAccess.GetOperation(),
		hex.EncodeToString(stateAccess.GetMainTrieKey()),
		hex.EncodeToString(stateAccess.GetMainTrieVal()),
		stateAccess.GetIndex(),
		strings.Join(dataTrieChanges, ", "),
	)
}

// Store will store the collected state changes if it has been configured with a storer
func (c *collector) Store() error {
	// TODO: evaluate adding a more explicit field check here
	if check.IfNil(c.storer) {
		return nil
	}

	stateAccessesForTxs := getStateAccessesForTxs(c.stateAccesses, true)
	for txHash, stateChange := range stateAccessesForTxs {
		marshalledData, err := c.marshaller.Marshal(stateChange)
		if err != nil {
			return fmt.Errorf("failed to marshal state changes: %w", err)
		}

		err = c.storer.Put([]byte(txHash), marshalledData)
		if err != nil {
			return fmt.Errorf("failed to store marshalled data: %w", err)
		}
	}

	return nil
}

// AddTxHashToCollectedStateChanges will try to set txHash field to each state change
// if the field is not already set
func (c *collector) AddTxHashToCollectedStateChanges(txHash []byte) {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	for i := len(c.stateAccesses) - 1; i >= 0; i-- {
		if len(c.stateAccesses[i].GetTxHash()) > 0 {
			break
		}

		log.Trace("added tx hash to state change", "txHash", txHash, "index", c.stateAccesses[i].GetIndex())
		c.stateAccesses[i].TxHash = txHash
	}
}

// SetIndexToLastStateChange will set index to the last state change
func (c *collector) SetIndexToLastStateChange(index int) error {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	if index < 0 {
		return fmt.Errorf("SetIndexToLastStateChange: %w for index %v, num state changes %v", state.ErrStateChangesIndexOutOfBounds, index, len(c.stateAccesses))
	}

	if len(c.stateAccesses) == 0 {
		return nil
	}

	for i := len(c.stateAccesses) - 1; i >= 0; i-- {
		if c.stateAccesses[i].GetIndex() != 0 {
			return nil
		}
		log.Trace("set index to last state change", "stateChange num", i, "index", index)
		c.stateAccesses[i].Index = int32(index)
	}

	return nil
}

// RevertToIndex will revert to index
func (c *collector) RevertToIndex(index int) error {
	if index < 0 {
		return fmt.Errorf("RevertToIndex: %w for index %v, num state changes %v", state.ErrStateChangesIndexOutOfBounds, index, len(c.stateAccesses))
	}

	if index == 0 {
		c.Reset()
		return nil
	}

	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	log.Trace("num state changes before revert", "num", len(c.stateAccesses))
	for i := len(c.stateAccesses) - 1; i >= 0; i-- {
		if c.stateAccesses[i].GetIndex() == int32(index) {
			c.stateAccesses = c.stateAccesses[:i+1]
			log.Trace("reverted to index", "index", index, "num state changes after revert", len(c.stateAccesses))
			break
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *collector) IsInterfaceNil() bool {
	return c == nil
}
