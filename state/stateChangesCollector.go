package state

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

const (
	workingDir              = "."
	defaultStateChangesPath = "stateChanges"
)

// DataTrieChange represents a change in the data trie
type DataTrieChange struct {
	Key []byte `json:"key"`
	Val []byte `json:"-"`
}

// StateChangeDTO is used to collect state changes
type StateChangeDTO struct {
	MainTrieKey     []byte           `json:"mainTrieKey"`
	MainTrieVal     []byte           `json:"-"`
	DataTrieChanges []DataTrieChange `json:"dataTrieChanges"`
}

// StateChangesForTx is used to collect state changes for a transaction hash
type StateChangesForTx struct {
	TxHash       []byte                   `json:"txHash"`
	Tx           *transaction.Transaction `json:"tx"`
	StateChanges []StateChangeDTO         `json:"stateChanges"`
}

type stateChangesCollector struct {
	stateChanges      []StateChangeDTO
	stateChangesForTx []StateChangesForTx
}

// NewStateChangesCollector creates a new StateChangesCollector
func NewStateChangesCollector() *stateChangesCollector {
	return &stateChangesCollector{
		stateChanges: []StateChangeDTO{},
	}
}

// AddStateChange adds a new state change to the collector
func (scc *stateChangesCollector) AddStateChange(stateChange StateChangeDTO) {
	scc.stateChanges = append(scc.stateChanges, stateChange)
}

// GetStateChanges returns the accumulated state changes
func (scc *stateChangesCollector) GetStateChanges() []StateChangesForTx {
	if len(scc.stateChanges) > 0 {
		scc.AddTxHashToCollectedStateChanges([]byte{}, nil)
	}

	return scc.stateChangesForTx
}

// Reset resets the state changes collector
func (scc *stateChangesCollector) Reset() {
	scc.stateChanges = make([]StateChangeDTO, 0)
	scc.stateChangesForTx = make([]StateChangesForTx, 0)
}

func (scc *stateChangesCollector) AddTxHashToCollectedStateChanges(txHash []byte, tx *transaction.Transaction) {
	stateChangesForTx := StateChangesForTx{
		TxHash:       txHash,
		Tx:           tx,
		StateChanges: scc.stateChanges,
	}

	scc.stateChanges = make([]StateChangeDTO, 0)
	scc.stateChangesForTx = append(scc.stateChangesForTx, stateChangesForTx)
}

func (scc *stateChangesCollector) DumpToJSONFile() error {
	directory := filepath.Join(workingDir, defaultStateChangesPath)
	args := core.ArgCreateFileArgument{
		Directory:     directory,
		Prefix:        "stateChanges",
		FileExtension: "json",
	}
	jsonFile, err := core.CreateFile(args)
	if err != nil {
		return err
	}

	marshalledData, err := json.Marshal(scc.stateChangesForTx)
	if err != nil {
		return err
	}

	// encoder := json.NewEncoder(jsonFile)

	// err = encoder.Encode(marshalledData)
	// if err != nil {
	// 	return err
	// }

	err = ioutil.WriteFile(jsonFile.Name(), marshalledData, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scc *stateChangesCollector) IsInterfaceNil() bool {
	return scc == nil
}
