package preprocess

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

func TestNewAccountStateProvider(t *testing.T) {
	t.Parallel()

	provider, err := newAccountStateProvider(nil, &testscommon.TxProcessorStub{})
	require.Nil(t, provider)
	require.ErrorIs(t, err, process.ErrNilAccountsAdapter)

	provider, err = newAccountStateProvider(&state.AccountsStub{}, nil)
	require.Nil(t, provider)
	require.ErrorIs(t, err, process.ErrNilTxProcessor)

	provider, err = newAccountStateProvider(&state.AccountsStub{}, &testscommon.TxProcessorStub{})
	require.NoError(t, err)
	require.NotNil(t, provider)
}
