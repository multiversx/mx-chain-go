package data

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialIndexingData_NewIndexingData(t *testing.T) {
	t.Parallel()

	id := NewIndexingData()

	require.False(t, check.IfNil(id))
}

func TestInitialIndexingData_Getters(t *testing.T) {
	t.Parallel()

	delegationTxs := make([]data.TransactionHandler, 0)
	scrsTxs := make(map[string]data.TransactionHandler)
	stakingTxs := make([]data.TransactionHandler, 0)
	deploySystemScTxs := make([]data.TransactionHandler, 0)

	id := &IndexingData{
		DelegationTxs:     delegationTxs,
		ScrsTxs:           scrsTxs,
		StakingTxs:        stakingTxs,
		DeploySystemScTxs: deploySystemScTxs,
	}

	require.False(t, check.IfNil(id))
	assert.Equal(t, delegationTxs, id.GetDelegationTxs())
	assert.Equal(t, scrsTxs, id.GetScrsTxs())
	assert.Equal(t, stakingTxs, id.GetStakingTxs())
	assert.Equal(t, deploySystemScTxs, id.GetDeploySystemScTxs())
}
