package transaction_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	txproc "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTxProcessorSimulate_NewTxProcErr(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	args.Accounts = nil

	txProc, err := txproc.NewTxProcessorSimulate(args)
	require.Nil(t, txProc)
	require.NotNil(t, err)
}

func TestNewTxProcessorSimulate_ShouldCheckBalanceAlwaysFalse(t *testing.T) {
	t.Parallel()

	args := createArgsForTxProcessor()
	txProc, _ := txproc.NewTxProcessorSimulate(args)

	assert.False(t, txProc.ShouldCheckBalance())
}

func TestProcessorSimulate_CheckTxValuesShouldIgnoreBalanceCheck(t *testing.T) {
	t.Parallel()

	adr1 := []byte{65}
	acnt1, err := state.NewUserAccount(adr1)
	assert.Nil(t, err)

	txProc, _ := txproc.NewTxProcessorSimulate(createArgsForTxProcessor())

	acnt1.Nonce = 5

	err = txProc.CheckTxValues(&transaction.Transaction{Nonce: 5}, acnt1, nil, false)
	require.Nil(t, err)
}
