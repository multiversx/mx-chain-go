package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
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

// ErrStateChangesIndexOutOfBounds signals that the state changes index is out of bounds
var ErrStateChangesIndexOutOfBounds = errors.New("state changes index out of bounds")

// StateChangeDTO is used to collect state changes
type StateChangeDTO struct {
	Index           int              `json:"-"`
	TxHash          []byte           `json:"-"`
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

type StateChangesTx struct {
	TxHash []byte                   `json:"txHash"`
	Tx     *transaction.Transaction `json:"tx"`
}

type stateChangesCollector struct {
	stateChanges    []StateChangeDTO
	stateChangesMut sync.RWMutex

	cachedTxs map[string]*transaction.Transaction

	storer types.Persister
}

// NewStateChangesCollector creates a new StateChangesCollector
func NewStateChangesCollector() *stateChangesCollector {
	// dbPath := filepath.Join(workingDir, "stateChangesDB", "StateChanges")

	// db, err := leveldb.NewSerialDB(dbPath, 2, 100, 10)
	// if err != nil {
	// 	log.Error("NewStateChangesCollector: failed to create level db")
	// }

	return &stateChangesCollector{
		stateChanges: make([]StateChangeDTO, 0),
		cachedTxs:    make(map[string]*transaction.Transaction),
		storer:       nil,
	}
}

// AddStateChange adds a new state change to the collector
func (scc *stateChangesCollector) AddStateChange(stateChange StateChangeDTO) {
	scc.stateChangesMut.Lock()
	scc.stateChanges = append(scc.stateChanges, stateChange)
	scc.stateChangesMut.Unlock()
}

// GetStateChanges returns the accumulated state changes
func (scc *stateChangesCollector) GetStateChanges() []StateChangesForTx {
	stateChangesForTx, err := scc.getStateChangesForTxs()
	if err != nil {
		log.Warn("failed to get state changes for tx", "error", err)
		return make([]StateChangesForTx, 0)
	}

	return stateChangesForTx
}

func (scc *stateChangesCollector) getStateChangesForTxs() ([]StateChangesForTx, error) {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	stateChangesForTxs := make([]StateChangesForTx, 0)

	for i := 0; i < len(scc.stateChanges); i++ {
		txHash := scc.stateChanges[i].TxHash

		if len(txHash) == 0 {
			log.Warn("empty tx hash, state change event not associated to a transaction")
			break
		}

		log.Warn("txHash", "txHash", string(txHash))

		cachedTx, txOk := scc.cachedTxs[string(txHash)]
		if !txOk {
			return nil, fmt.Errorf("did not find tx in cache")
		}

		innerStateChangesForTx := make([]StateChangeDTO, 0)
		for j := i; j < len(scc.stateChanges); j++ {
			txHash2 := scc.stateChanges[j].TxHash
			if !bytes.Equal(txHash, txHash2) {
				i = j
				break
			}

			innerStateChangesForTx = append(innerStateChangesForTx, scc.stateChanges[j])
			i = j
		}

		stateChangesForTx := StateChangesForTx{
			TxHash:       txHash,
			Tx:           cachedTx,
			StateChanges: innerStateChangesForTx,
		}
		stateChangesForTxs = append(stateChangesForTxs, stateChangesForTx)
	}

	return stateChangesForTxs, nil
}

// Reset resets the state changes collector
func (scc *stateChangesCollector) Reset() {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	scc.stateChanges = make([]StateChangeDTO, 0)
	scc.cachedTxs = make(map[string]*transaction.Transaction)
}

func (scc *stateChangesCollector) AddTxHashToCollectedStateChanges(txHash []byte, tx *transaction.Transaction) {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	scc.cachedTxs[string(txHash)] = tx

	for i := len(scc.stateChanges) - 1; i >= 0; i-- {
		if len(scc.stateChanges[i].TxHash) > 0 {
			break
		}

		scc.stateChanges[i].TxHash = txHash
	}
}

func (scc *stateChangesCollector) SetIndexToLastStateChange(index int) error {
	if index > len(scc.stateChanges) || index < 0 {
		return ErrStateChangesIndexOutOfBounds
	}

	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	scc.stateChanges[len(scc.stateChanges)-1].Index = index

	return nil
}

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
		if scc.stateChanges[i].Index == index {
			scc.stateChanges = scc.stateChanges[:i]
			break
		}
	}

	return nil
}

func (scc *stateChangesCollector) DumpToJSONFile() error {
	stateChangesForTx, err := scc.getStateChangesForTxs()
	if err != nil {
		return err
	}

	for _, stateChange := range stateChangesForTx {
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
