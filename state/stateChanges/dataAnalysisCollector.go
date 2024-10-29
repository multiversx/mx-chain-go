package stateChanges

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/check"
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
)

type dataAnalysisStateChangeDTO struct {
	state.StateChange
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

type dataAnalysisStateChangesForTx struct {
	StateChangesForTx
	Tx *transaction.Transaction `json:"tx"`
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
		return nil, storage.ErrNilPersister
	}

	return &dataAnalysisCollector{
		stateChangesCollector: NewStateChangesCollector(true, true),
		cachedTxs:             make(map[string]*transaction.Transaction),
		storer:                storer,
	}, nil
}

// AddSaveAccountStateChange adds a new state change for the save account operation
func (scc *dataAnalysisCollector) AddSaveAccountStateChange(oldAccount, account vmcommon.AccountHandler, stateChange state.StateChange) {
	dataAnalysisStateChange := &dataAnalysisStateChangeDTO{
		StateChange: stateChange,
	}

	checkAccountChanges(oldAccount, account, dataAnalysisStateChange)

	scc.AddStateChange(dataAnalysisStateChange)
}

func checkAccountChanges(oldAcc, newAcc vmcommon.AccountHandler, stateChange *dataAnalysisStateChangeDTO) {
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
func (scc *dataAnalysisCollector) AddStateChange(stateChange state.StateChange) {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	scc.stateChanges = append(scc.stateChanges, stateChange)
}

func (scc *dataAnalysisCollector) getDataAnalysisStateChangesForTxs() ([]dataAnalysisStateChangesForTx, error) {
	stateChangesForTxs, err := scc.getStateChangesForTxs()
	if err != nil {
		return nil, err
	}

	dataAnalysisStateChangesForTxs := make([]dataAnalysisStateChangesForTx, 0)

	for _, stateChangeForTx := range stateChangesForTxs {
		cachedTx, txOk := scc.cachedTxs[string(stateChangeForTx.TxHash)]
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

// AddTxHashToCollectedStateChanges will add the transaction to the cache
func (scc *dataAnalysisCollector) AddTxHashToCollectedStateChanges(txHash []byte, tx *transaction.Transaction) {
	scc.cachedTxs[string(txHash)] = tx
	scc.addTxHashToCollectedStateChanges(txHash)
}

// Reset resets the state changes collector
func (scc *dataAnalysisCollector) Reset() {
	scc.stateChangesMut.Lock()
	defer scc.stateChangesMut.Unlock()

	scc.resetStateChangesUnprotected()
	scc.cachedTxs = make(map[string]*transaction.Transaction)
}

// Publish will write the stored state changes
func (scc *dataAnalysisCollector) Publish() (map[string]*data.StateChanges, error) {
	stateChangesForTx, err := scc.getDataAnalysisStateChangesForTxs()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data analysis state changes for tx: %w", err)
	}

	for _, stateChange := range stateChangesForTx {
		marshalledData, err := json.Marshal(stateChange)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal state changes to JSON: %w", err)
		}

		err = scc.storer.Put(stateChange.TxHash, marshalledData)
		if err != nil {
			return nil, fmt.Errorf("failed to store marshalled data: %w", err)
		}
	}

	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scc *dataAnalysisCollector) IsInterfaceNil() bool {
	return scc == nil
}
