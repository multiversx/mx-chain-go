package stateChanges

import (
	"bytes"
	"sync"

	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
)

var log = logger.GetOrCreate("state/stateChanges")

// StateChangesForTx is used to collect state changes for a transaction hash
type StateChangesForTx struct {
	TxHash       []byte              `json:"txHash"`
	StateChanges []state.StateChange `json:"stateChanges"`
}

type stateChangesCollector struct {
	collectRead  bool
	collectWrite bool

	stateChanges    []state.StateChange
	stateChangesMut sync.RWMutex
}

// NewStateChangesCollector creates a new StateChangesCollector
func NewStateChangesCollector(collectRead bool, collectWrite bool) *stateChangesCollector {
	return &stateChangesCollector{
		collectRead:  collectRead,
		collectWrite: collectWrite,
		stateChanges: make([]state.StateChange, 0),
	}
}

// AddSaveAccountStateChange adds a new state change for the save account operation
func (scc *stateChangesCollector) AddSaveAccountStateChange(_, _ vmcommon.AccountHandler, stateChange state.StateChange) {
	scc.AddStateChange(stateChange)
}

// AddStateChange adds a new state change to the collector
func (scc *stateChangesCollector) AddStateChange(stateChange state.StateChange) {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	if stateChange.GetType() == data.Write && scc.collectWrite {
		scc.stateChanges = append(scc.stateChanges, stateChange)
	}

	if stateChange.GetType() == data.Read && scc.collectRead {
		scc.stateChanges = append(scc.stateChanges, stateChange)
	}
}

func (scc *stateChangesCollector) getStateChangesForTxs() ([]StateChangesForTx, error) {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	stateChangesForTxs := make([]StateChangesForTx, 0)

	for i := 0; i < len(scc.stateChanges); i++ {
		txHash := scc.stateChanges[i].GetTxHash()

		if len(txHash) == 0 {
			log.Warn("empty tx hash, state change event not associated to a transaction")
			continue
		}

		innerStateChangesForTx := make([]state.StateChange, 0)
		for j := i; j < len(scc.stateChanges); j++ {
			txHash2 := scc.stateChanges[j].GetTxHash()
			if !bytes.Equal(txHash, txHash2) {
				i = j
				break
			}

			innerStateChangesForTx = append(innerStateChangesForTx, scc.stateChanges[j])
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

// GetStateChangesForTxs will retrieve the state changes linked with the tx hash.
func (scc *stateChangesCollector) GetStateChangesForTxs() map[string]*data.StateChanges {
	scc.stateChangesMut.RLock()
	defer scc.stateChangesMut.RUnlock()

	stateChangesForTxs := make(map[string]*data.StateChanges)

	for _, stateChange := range scc.stateChanges {
		txHash := string(stateChange.GetTxHash())

		st, ok := stateChange.(*data.StateChange)
		if !ok {
			continue
		}

		_, ok = stateChangesForTxs[txHash]
		if !ok {
			stateChangesForTxs[txHash] = &data.StateChanges{
				StateChanges: []*data.StateChange{st},
			}
		} else {
			stateChangesForTxs[txHash].StateChanges = append(stateChangesForTxs[txHash].StateChanges, st)
		}
	}

	return stateChangesForTxs
}

// Reset resets the state changes collector
func (scc *stateChangesCollector) Reset() {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	scc.resetStateChangesUnprotected()
}

func (scc *stateChangesCollector) resetStateChangesUnprotected() {
	scc.stateChanges = make([]state.StateChange, 0)
}

// AddTxHashToCollectedStateChanges will try to set txHash field to each state change
// if the field is not already set
func (scc *stateChangesCollector) AddTxHashToCollectedStateChanges(txHash []byte, _ *transaction.Transaction) {
	scc.addTxHashToCollectedStateChanges(txHash)
}

func (scc *stateChangesCollector) addTxHashToCollectedStateChanges(txHash []byte) {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	for i := len(scc.stateChanges) - 1; i >= 0; i-- {
		if len(scc.stateChanges[i].GetTxHash()) > 0 {
			break
		}

		scc.stateChanges[i].SetTxHash(txHash)
	}
}

// SetIndexToLastStateChange will set index to the last state change
func (scc *stateChangesCollector) SetIndexToLastStateChange(index int) error {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	if index > len(scc.stateChanges) || index < 0 {
		return state.ErrStateChangesIndexOutOfBounds
	}

	if len(scc.stateChanges) == 0 {
		return nil
	}

	scc.stateChanges[len(scc.stateChanges)-1].SetIndex(int32(index))

	return nil
}

// RevertToIndex will revert to index
func (scc *stateChangesCollector) RevertToIndex(index int) error {
	if index > len(scc.stateChanges) || index < 0 {
		return state.ErrStateChangesIndexOutOfBounds
	}

	if index == 0 {
		scc.Reset()
		return nil
	}

	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	for i := len(scc.stateChanges) - 1; i >= 0; i-- {
		if scc.stateChanges[i].GetIndex() == int32(index) {
			scc.stateChanges = scc.stateChanges[:i]
			break
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scc *stateChangesCollector) IsInterfaceNil() bool {
	return scc == nil
}
