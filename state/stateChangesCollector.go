package state

import (
	"encoding/json"
	"path/filepath"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-storage-go/leveldb"
	"github.com/multiversx/mx-chain-storage-go/types"
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
	Operation       string           `json:"operation"`
	Nonce           bool             `json:"nonceChanged"`
	Balance         bool             `json:"balanceChanged"`
	CodeHash        bool             `json:"codeHashChanged"`
	RootHash        bool             `json:"rootHashChanged"`
	DeveloperReward bool             `json:"developerRewardChanged"`
	OwnerAddress    bool             `json:"ownerAddressChanged"`
	UserName        bool             `json:"userNameChanged"`
	CodeMetadata    bool             `json:"codeMetadataChanged"`
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

	storer types.Persister
}

// NewStateChangesCollector creates a new StateChangesCollector
func NewStateChangesCollector() *stateChangesCollector {
	dbPath := filepath.Join(workingDir, "stateChangesDB", "StateChanges")

	db, err := leveldb.NewSerialDB(dbPath, 2, 100, 10)
	if err != nil {
		log.Error("NewStateChangesCollector: failed to create level db")
	}

	return &stateChangesCollector{
		stateChanges: []StateChangeDTO{},
		storer:       db,
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
		marshalledData, err := json.Marshal(stateChange)
		if err != nil {
			return err
		}

		err = scc.storer.Put(stateChange.TxHash, marshalledData)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scc *stateChangesCollector) IsInterfaceNil() bool {
	return scc == nil
}
