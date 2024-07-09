package state

import (
	"encoding/json"
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
	Type string `json:"type"`
	Key  []byte `json:"key"`
	Val  []byte `json:"-"`
}

// StateChangeDTO is used to collect state changes
type StateChangeDTO struct {
	Type            string           `json:"type"`
	MainTrieKey     []byte           `json:"mainTrieKey"`
	MainTrieVal     []byte           `json:"-"`
	Operation       string           `json:"operation,omitempty"`
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

	jsonFile    *os.File
	jsonEncoder *json.Encoder
}

// NewStateChangesCollector creates a new StateChangesCollector
func NewStateChangesCollector() *stateChangesCollector {
	directory := filepath.Join(workingDir, defaultStateChangesPath)
	args := core.ArgCreateFileArgument{
		Directory:     directory,
		Prefix:        "stateChanges",
		FileExtension: "json",
	}

	jsonFile, err := core.CreateFile(args)
	if err != nil {
		log.Error("NewStateChangesCollector: failed to create json file")
	}

	// _, err = jsonFile.Write([]byte("["))
	// if err != nil {
	// 	log.Error("NewStateChangesCollector: failed to write to json file")
	// }

	encoder := json.NewEncoder(jsonFile)

	return &stateChangesCollector{
		stateChanges: []StateChangeDTO{},
		jsonFile:     jsonFile,
		jsonEncoder:  encoder,
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
	for _, stateChange := range scc.stateChangesForTx {
		err := scc.jsonEncoder.Encode(stateChange)
		if err != nil {
			return err
		}
	}

	// _, err := scc.jsonFile.Write([]byte(","))
	// if err != nil {
	// 	return err
	// }

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scc *stateChangesCollector) IsInterfaceNil() bool {
	return scc == nil
}
