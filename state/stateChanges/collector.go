package stateChanges

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
)

var log = logger.GetOrCreate("state/stateChanges")

// StateChangesForTx is used to collect state changes for a transaction hash
type StateChangesForTx struct {
	TxHash       []byte              `json:"txHash"`
	StateChanges []state.StateChange `json:"stateChanges"`
}

type collector struct {
	collectRead     bool
	collectWrite    bool
	stateChanges    []state.StateChange
	stateChangesMut sync.RWMutex
	cachedTxs       map[string]coreData.TransactionHandler
	storer          storage.Persister
}

// NewCollector will collect based on the options the state changes.
func NewCollector(opts ...CollectorOption) *collector {
	c := &collector{stateChanges: make([]state.StateChange, 0)}
	for _, opt := range opts {
		opt(c)
	}

	if c.storer != nil {
		c.cachedTxs = make(map[string]coreData.TransactionHandler)
	}

	log.Debug("created new state changes collector",
		"withRead", c.collectRead,
		"withWrite", c.collectWrite,
	)

	return c
}

// AddStateChange adds a new state change to the collector
func (c *collector) AddStateChange(stateChange state.StateChange) {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	if stateChange.GetType() == data.Write && c.collectWrite {
		c.stateChanges = append(c.stateChanges, stateChange)
		log.Trace("added WRITE state change to collector", "operation", stateChange.GetOperation())
	}

	if stateChange.GetType() == data.Read && c.collectRead {
		c.stateChanges = append(c.stateChanges, stateChange)
		log.Trace("added READ state change to collector", "operation", stateChange.GetOperation())
	}
}

// AddSaveAccountStateChange adds a new state change for the save account operation
func (c *collector) AddSaveAccountStateChange(oldAccount, account vmcommon.AccountHandler, stateChange state.StateChange) {
	if c.storer != nil {
		dataAnalysisStateChange := &dataAnalysisStateChangeDTO{
			StateChange: stateChange,
		}

		checkAccountChanges(oldAccount, account, dataAnalysisStateChange)

		c.AddStateChange(dataAnalysisStateChange)
		return
	}

	c.AddStateChange(stateChange)
}

// Reset resets the state changes collector
func (c *collector) Reset() {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	c.stateChanges = make([]state.StateChange, 0)
	if c.storer != nil {
		c.cachedTxs = make(map[string]coreData.TransactionHandler)
	}
	log.Trace("reset state changes collector")
}

// Publish will export state changes
func (c *collector) Publish() (map[string]*data.StateChanges, error) {
	c.stateChangesMut.RLock()
	defer c.stateChangesMut.RUnlock()

	stateChangesForTxs := make(map[string]*data.StateChanges)
	for _, stateChange := range c.stateChanges {
		txHash := string(stateChange.GetTxHash())
		if len(txHash) == 0 {
			log.Warn("empty tx hash, state change event not associated to a transaction", "stateChange", stateChangeToString(stateChange))
			continue
		}

		st, ok := stateChange.(*data.StateChange)
		if !ok {
			log.Warn("failed to cast state change to data.StateChange", "stateChange", stateChangeToString(stateChange))
			continue
		}

		_, ok = stateChangesForTxs[txHash]
		if !ok {
			log.Trace("created new state changes for tx", "txHash", txHash)
			stateChangesForTxs[txHash] = &data.StateChanges{
				StateChanges: []*data.StateChange{st},
			}
		} else {
			log.Trace("appended state change to existing state changes for tx", "txHash", txHash)
			stateChangesForTxs[txHash].StateChanges = append(stateChangesForTxs[txHash].StateChanges, st)
		}
	}
	log.Trace("published state changes", "numTxs", len(stateChangesForTxs))

	return stateChangesForTxs, nil
}

func stateChangeToString(stateChange state.StateChange) string {
	dataTrieChanges := make([]string, len(stateChange.GetDataTrieChanges()))
	for i, dataTrieChange := range stateChange.GetDataTrieChanges() {
		dataTrieChanges[i] = fmt.Sprintf("key: %v, val: %v, type: %v", hex.EncodeToString(dataTrieChange.Key), hex.EncodeToString(dataTrieChange.Val), dataTrieChange.Type)
	}
	return fmt.Sprintf("type: %v, operation: %v, mainTrieKey: %v, mainTrieVal: %v, index: %v, dataTrieChanges: %v",
		stateChange.GetType(),
		stateChange.GetOperation(),
		hex.EncodeToString(stateChange.GetMainTrieKey()),
		hex.EncodeToString(stateChange.GetMainTrieVal()),
		stateChange.GetIndex(),
		strings.Join(dataTrieChanges, ", "),
	)
}

// Store will store the collected state changes if it has been configured with a storer
func (c *collector) Store() error {
	// TODO: evaluate adding a more explicit field check here
	if check.IfNil(c.storer) {
		return nil
	}

	stateChangesForTx, err := c.getDataAnalysisStateChangesForTxs()
	if err != nil {
		return fmt.Errorf("failed to retrieve data analysis state changes for tx: %w", err)
	}

	for _, stateChange := range stateChangesForTx {
		marshalledData, err := json.Marshal(stateChange)
		if err != nil {
			return fmt.Errorf("failed to marshal state changes to JSON: %w", err)
		}

		err = c.storer.Put(stateChange.TxHash, marshalledData)
		if err != nil {
			return fmt.Errorf("failed to store marshalled data: %w", err)
		}
	}

	return nil
}

// AddTxHashToCollectedStateChanges will try to set txHash field to each state change
// if the field is not already set
func (c *collector) AddTxHashToCollectedStateChanges(txHash []byte, tx coreData.TransactionHandler) {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	if c.storer != nil {
		c.cachedTxs[string(txHash)] = tx
	}

	for i := len(c.stateChanges) - 1; i >= 0; i-- {
		if len(c.stateChanges[i].GetTxHash()) > 0 {
			break
		}

		log.Trace("added tx hash to state change", "txHash", txHash, "index", c.stateChanges[i].GetIndex())
		c.stateChanges[i].SetTxHash(txHash)
	}
}

// SetIndexToLastStateChange will set index to the last state change
func (c *collector) SetIndexToLastStateChange(index int) error {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	if index > len(c.stateChanges) || index < 0 {
		return fmt.Errorf("SetIndexToLastStateChange: %w for index %v, num state changes %v", state.ErrStateChangesIndexOutOfBounds, index, len(c.stateChanges))
	}

	if len(c.stateChanges) == 0 {
		return nil
	}

	log.Trace("set index to last state change", "index", index)
	c.stateChanges[len(c.stateChanges)-1].SetIndex(int32(index))

	return nil
}

// RevertToIndex will revert to index
func (c *collector) RevertToIndex(index int) error {
	if index > len(c.stateChanges) || index < 0 {
		return fmt.Errorf("RevertToIndex: %w for index %v, num state changes %v", state.ErrStateChangesIndexOutOfBounds, index, len(c.stateChanges))
	}

	if index == 0 {
		c.Reset()
		return nil
	}

	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	log.Trace("num state changes before revert", "num", len(c.stateChanges))
	for i := len(c.stateChanges) - 1; i >= 0; i-- {
		if c.stateChanges[i].GetIndex() == int32(index) {
			c.stateChanges = c.stateChanges[:i]
			log.Trace("reverted to index", "index", index, "num state changes after revert", len(c.stateChanges))
			break
		}
	}

	return nil
}

func (c *collector) getStateChangesForTxs() ([]StateChangesForTx, error) {
	c.stateChangesMut.Lock()
	defer c.stateChangesMut.Unlock()

	stateChangesForTxs := make([]StateChangesForTx, 0)

	for i := 0; i < len(c.stateChanges); i++ {
		txHash := c.stateChanges[i].GetTxHash()

		if len(txHash) == 0 {
			log.Warn("empty tx hash, state change event not associated to a transaction")
			continue
		}

		innerStateChangesForTx := make([]state.StateChange, 0)
		for j := i; j < len(c.stateChanges); j++ {
			txHash2 := c.stateChanges[j].GetTxHash()
			if !bytes.Equal(txHash, txHash2) {
				i = j
				break
			}

			innerStateChangesForTx = append(innerStateChangesForTx, c.stateChanges[j])
			i = j
		}

		stateChangesForTx := StateChangesForTx{
			TxHash:       txHash,
			StateChanges: innerStateChangesForTx,
		}
		stateChangesForTxs = append(stateChangesForTxs, stateChangesForTx)
	}

	return stateChangesForTxs, nil
}

func (c *collector) getDataAnalysisStateChangesForTxs() ([]dataAnalysisStateChangesForTx, error) {
	stateChangesForTxs, err := c.getStateChangesForTxs()
	if err != nil {
		return nil, err
	}

	dataAnalysisStateChangesForTxs := make([]dataAnalysisStateChangesForTx, 0)

	for _, stateChangeForTx := range stateChangesForTxs {
		cachedTx, txOk := c.cachedTxs[string(stateChangeForTx.TxHash)]
		if !txOk {
			return nil, fmt.Errorf("did not find tx in cache")
		}

		stateChangesForTx := dataAnalysisStateChangesForTx{
			StateChangesForTx: stateChangeForTx,
			Tx:                cachedTx,
		}
		dataAnalysisStateChangesForTxs = append(dataAnalysisStateChangesForTxs, stateChangesForTx)
	}

	return dataAnalysisStateChangesForTxs, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *collector) IsInterfaceNil() bool {
	return c == nil
}
