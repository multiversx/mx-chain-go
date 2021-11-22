package data

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

type IndexingData struct {
	DelegationTxs     []data.TransactionHandler
	ScrsTxs           map[string]data.TransactionHandler
	StakingTxs        []data.TransactionHandler
	DeploySystemScTxs []data.TransactionHandler
}

func NewIndexingData() *IndexingData {
	delegationTxs := make([]data.TransactionHandler, 0)
	scrsTxs := make(map[string]data.TransactionHandler)
	stakingTxs := make([]data.TransactionHandler, 0)
	deploySystemScTxs := make([]data.TransactionHandler, 0)

	return &IndexingData{
		DelegationTxs:     delegationTxs,
		ScrsTxs:           scrsTxs,
		StakingTxs:        stakingTxs,
		DeploySystemScTxs: deploySystemScTxs,
	}
}

// TODO: upate comments

// GetScrsTxs will return the smartContracts results transactions per shard
func (id *IndexingData) GetScrsTxs() map[string]data.TransactionHandler {
	return id.ScrsTxs
}

// GetDelegationTxs will return the delegation transactions per shard
func (id *IndexingData) GetDelegationTxs() []data.TransactionHandler {
	return id.DelegationTxs
}

// GetStakingTxs will return the staking transactions per shard
func (id *IndexingData) GetStakingTxs() []data.TransactionHandler {
	return id.StakingTxs
}

// GetDeploySystemScTxs
func (id *IndexingData) GetDeploySystemScTxs() []data.TransactionHandler {
	return id.DeploySystemScTxs
}

// IsInterfaceNil returns if underlying object is true
func (id *IndexingData) IsInterfaceNil() bool {
	return id == nil
}
