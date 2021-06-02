package transaction_test

import (
	"testing"

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
