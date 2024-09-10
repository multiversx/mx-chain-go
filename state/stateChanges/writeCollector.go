package stateChanges

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var log = logger.GetOrCreate("state/stateChanges")

// ErrStateChangesIndexOutOfBounds signals that the state changes index is out of bounds
var ErrStateChangesIndexOutOfBounds = errors.New("state changes index out of bounds")

type StateChange interface {
	GetType() string
	GetIndex() int32
	GetTxHash() []byte
	GetMainTrieKey() []byte
	GetMainTrieVal() []byte
	GetOperation() string
	GetDataTrieChanges() []*data.DataTrieChange

	SetTxHash(txHash []byte)
	SetIndex(index int32)
}

// StateChangesForTx is used to collect state changes for a transaction hash
type StateChangesForTx struct {
	TxHash       []byte        `json:"txHash"`
	StateChanges []StateChange `json:"stateChanges"`
}

type stateChangesCollector struct {
	stateChanges    []StateChange
	stateChangesMut sync.RWMutex
}

// NewStateChangesCollector creates a new StateChangesCollector
func NewStateChangesCollector() *stateChangesCollector {
	// TODO: add outport driver

	return &stateChangesCollector{
		stateChanges: make([]StateChange, 0),
	}
}

// AddSaveAccountStateChange adds a new state change for the save account operation
func (scc *stateChangesCollector) AddSaveAccountStateChange(_, _ vmcommon.AccountHandler, stateChange StateChange) {
	scc.AddStateChange(stateChange)
}

// AddStateChange adds a new state change to the collector
func (scc *stateChangesCollector) AddStateChange(stateChange StateChange) {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	// TODO: add custom type for stateChange type
	if stateChange.GetType() == "write" {
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
			break
		}

		innerStateChangesForTx := make([]StateChange, 0)
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
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	stateChangesForTxs := make(map[string]*data.StateChanges)

	for _, stateChange := range scc.stateChanges {
		txHash := string(stateChange.GetTxHash())

		if sc, ok := stateChangesForTxs[txHash]; !ok {
			stateChangesForTxs[txHash] = &data.StateChanges{StateChanges: []*data.StateChange{
				{
					Type:            stateChange.GetType(),
					Index:           stateChange.GetIndex(),
					TxHash:          stateChange.GetTxHash(),
					MainTrieKey:     stateChange.GetMainTrieKey(),
					MainTrieVal:     stateChange.GetMainTrieVal(),
					Operation:       stateChange.GetOperation(),
					DataTrieChanges: stateChange.GetDataTrieChanges(),
				},
			},
			}
		} else {
			stateChangesForTxs[txHash].StateChanges = append(sc.StateChanges,
				&data.StateChange{
					Type:            stateChange.GetType(),
					Index:           stateChange.GetIndex(),
					TxHash:          stateChange.GetTxHash(),
					MainTrieKey:     stateChange.GetMainTrieKey(),
					MainTrieVal:     stateChange.GetMainTrieVal(),
					Operation:       stateChange.GetOperation(),
					DataTrieChanges: stateChange.GetDataTrieChanges(),
				},
			)
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
	scc.stateChanges = make([]StateChange, 0)
}

// AddTxHashToCollectedStateChanges will try to set txHash field to each state change
// if the field is not already set
func (scc *stateChangesCollector) AddTxHashToCollectedStateChanges(txHash []byte, tx *transaction.Transaction) {
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
		return ErrStateChangesIndexOutOfBounds
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
		return ErrStateChangesIndexOutOfBounds
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

// Publish will export state changes
func (scc *stateChangesCollector) Publish() error {
	stateChangesForTx, err := scc.getStateChangesForTxs()
	if err != nil {
		return err
	}

	printStateChanges(stateChangesForTx)

	return nil
}

func printStateChanges(stateChanges []StateChangesForTx) {
	for _, stateChange := range stateChanges {

		if stateChange.TxHash != nil {
			fmt.Println(hex.EncodeToString(stateChange.TxHash))
		}

		for _, st := range stateChange.StateChanges {
			fmt.Println(st)
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (scc *stateChangesCollector) IsInterfaceNil() bool {
	return scc == nil
}
