package data

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// IndexingData specify the transactions sets that will be used for indexing
type IndexingData struct {
	DelegationTxs      []data.TransactionHandler
	ScrsTxs            map[string]data.TransactionHandler
	StakingTxs         []data.TransactionHandler
	DeploySystemScTxs  []data.TransactionHandler
	DeployInitialSCTxs []data.TransactionHandler
}

// NewIndexingData creates a new IndexingData structure
func NewIndexingData() *IndexingData {
	delegationTxs := make([]data.TransactionHandler, 0)
	scrsTxs := make(map[string]data.TransactionHandler)
	stakingTxs := make([]data.TransactionHandler, 0)
	deploySystemScTxs := make([]data.TransactionHandler, 0)
	deployInitialSCTxs := make([]data.TransactionHandler, 0)

	return &IndexingData{
		DelegationTxs:      delegationTxs,
		ScrsTxs:            scrsTxs,
		StakingTxs:         stakingTxs,
		DeploySystemScTxs:  deploySystemScTxs,
		DeployInitialSCTxs: deployInitialSCTxs,
	}
}

// GetScrsTxs will return the smartContract results transactions and
// corresponding hashes
func (id *IndexingData) GetScrsTxs() map[string]data.TransactionHandler {
	return id.ScrsTxs
}

// GetDelegationTxs will return the delegation transactions
func (id *IndexingData) GetDelegationTxs() []data.TransactionHandler {
	return id.DelegationTxs
}

// GetStakingTxs will return the staking transactions
func (id *IndexingData) GetStakingTxs() []data.TransactionHandler {
	return id.StakingTxs
}

// GetDeploySystemScTxs will return deploy system SC for metachain
func (id *IndexingData) GetDeploySystemScTxs() []data.TransactionHandler {
	return id.DeploySystemScTxs
}

// GetDeployInitialSCTxs will return deploy system SC for metachain
func (id *IndexingData) GetDeployInitialSCTxs() []data.TransactionHandler {
	return id.DeployInitialSCTxs
}

// IsInterfaceNil returns if underlying object is true
func (id *IndexingData) IsInterfaceNil() bool {
	return id == nil
}
