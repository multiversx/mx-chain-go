package transaction_test

import (
	"testing"

	txproc "github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewMetaTxProcessorSimulate_NewMetaProcErr(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	args.Accounts = nil

	txProc, err := txproc.NewMetaTxProcessorSimulate(args)
	require.Nil(t, txProc)
	require.NotNil(t, err)
}

func TestNewMetaTxProcessorSimulate_ShouldCheckBalanceAlwaysFalse(t *testing.T) {
	t.Parallel()

	args := createMockNewMetaTxArgs()
	txProc, _ := txproc.NewMetaTxProcessorSimulate(args)

	require.False(t, txProc.ShouldCheckBalance())
}
