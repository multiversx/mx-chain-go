package stateChanges

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const (
	workingDir              = "."
	defaultStateChangesPath = "stateChanges"
)

// TODO: use proto stucts
type DataAnalysisStateChangeDTO struct {
	StateChange
	Operation       string `json:"operation"`
	Nonce           bool   `json:"nonceChanged"`
	Balance         bool   `json:"balanceChanged"`
	CodeHash        bool   `json:"codeHashChanged"`
	RootHash        bool   `json:"rootHashChanged"`
	DeveloperReward bool   `json:"developerRewardChanged"`
	OwnerAddress    bool   `json:"ownerAddressChanged"`
	UserName        bool   `json:"userNameChanged"`
	CodeMetadata    bool   `json:"codeMetadataChanged"`
}

type DataAnalysisStateChangesForTx struct {
	TxHash       []byte                   `json:"txHash"`
	Tx           *transaction.Transaction `json:"tx"`
	StateChanges []StateChange            `json:"stateChanges"`
}

type userAccountHandler interface {
	GetCodeMetadata() []byte
	GetCodeHash() []byte
	GetRootHash() []byte
	GetBalance() *big.Int
	GetDeveloperReward() *big.Int
	GetOwnerAddress() []byte
	GetUserName() []byte
	vmcommon.AccountHandler
}

type dataAnalysisCollector struct {
	*stateChangesCollector

	cachedTxs map[string]*transaction.Transaction
	storer    storage.Persister
}

// NewDataAnalysisStateChangesCollector will create a new instance of data analysis collector
func NewDataAnalysisStateChangesCollector(storer storage.Persister) (*dataAnalysisCollector, error) {
	if check.IfNil(storer) {
		return nil, storage.ErrNilPersisterFactory
	}

	return &dataAnalysisCollector{
		stateChangesCollector: &stateChangesCollector{
			stateChanges: make([]StateChange, 0),
		},
		cachedTxs: make(map[string]*transaction.Transaction),
		storer:    storer,
	}, nil
}

// AddSaveAccountStateChange adds a new state change for the save account operation
func (scc *dataAnalysisCollector) AddSaveAccountStateChange(oldAccount, account vmcommon.AccountHandler, stateChange StateChange) {
	dataAnalysisStateChange := &DataAnalysisStateChangeDTO{
		StateChange: stateChange,
	}

	checkAccountChanges(oldAccount, account, dataAnalysisStateChange)

	scc.AddStateChange(stateChange)
}

func checkAccountChanges(oldAcc, newAcc vmcommon.AccountHandler, stateChange *DataAnalysisStateChangeDTO) {
	baseNewAcc, newAccOk := newAcc.(userAccountHandler)
	if !newAccOk {
		return
	}
	baseOldAccount, oldAccOk := oldAcc.(userAccountHandler)
	if !oldAccOk {
		return
	}

	if baseNewAcc.GetNonce() != baseOldAccount.GetNonce() {
		stateChange.Nonce = true
	}

	if baseNewAcc.GetBalance().Uint64() != baseOldAccount.GetBalance().Uint64() {
		stateChange.Balance = true
	}

	if !bytes.Equal(baseNewAcc.GetCodeHash(), baseOldAccount.GetCodeHash()) {
		stateChange.CodeHash = true
	}

	if !bytes.Equal(baseNewAcc.GetRootHash(), baseOldAccount.GetRootHash()) {
		stateChange.RootHash = true
	}

	if !bytes.Equal(baseNewAcc.GetDeveloperReward().Bytes(), baseOldAccount.GetDeveloperReward().Bytes()) {
		stateChange.DeveloperReward = true
	}

	if !bytes.Equal(baseNewAcc.GetOwnerAddress(), baseOldAccount.GetOwnerAddress()) {
		stateChange.OwnerAddress = true
	}

	if !bytes.Equal(baseNewAcc.GetUserName(), baseOldAccount.GetUserName()) {
		stateChange.UserName = true
	}

	if !bytes.Equal(baseNewAcc.GetCodeMetadata(), baseOldAccount.GetCodeMetadata()) {
		stateChange.CodeMetadata = true
	}
}

// AddStateChange adds a new state change to the collector
func (scc *dataAnalysisCollector) AddStateChange(stateChange StateChange) {
	scc.stateChangesMut.Lock()
	scc.stateChanges = append(scc.stateChanges, stateChange)
	scc.stateChangesMut.Unlock()
}

func (scc *dataAnalysisCollector) getStateChangesForTxs() ([]DataAnalysisStateChangesForTx, error) {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	stateChangesForTxs := make([]DataAnalysisStateChangesForTx, 0)

	for i := 0; i < len(scc.stateChanges); i++ {
		txHash := scc.stateChanges[i].GetTxHash()

		if len(txHash) == 0 {
			log.Warn("empty tx hash, state change event not associated to a transaction")
			break
		}

		cachedTx, txOk := scc.cachedTxs[string(txHash)]
		if !txOk {
			return nil, fmt.Errorf("did not find tx in cache")
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

		stateChangesForTx := DataAnalysisStateChangesForTx{
			TxHash:       txHash,
			Tx:           cachedTx,
			StateChanges: innerStateChangesForTx,
		}
		stateChangesForTxs = append(stateChangesForTxs, stateChangesForTx)
	}

	return stateChangesForTxs, nil
}

// Reset resets the state changes collector
func (scc *dataAnalysisCollector) Reset() {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	scc.stateChanges = make([]StateChange, 0)
	scc.cachedTxs = make(map[string]*transaction.Transaction)
}

// Publish will export state changes
func (scc *dataAnalysisCollector) Publish() error {
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
func (scc *dataAnalysisCollector) IsInterfaceNil() bool {
	return scc == nil
}
