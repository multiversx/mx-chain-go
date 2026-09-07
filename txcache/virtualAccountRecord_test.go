package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func Test_newVirtualAccountRecord(t *testing.T) {
	t.Parallel()

	initialNonce := core.OptionalUint64{
		Value:    uint64(1),
		HasValue: true,
	}

	record, err := newVirtualAccountRecord(initialNonce, big.NewInt(0))
	require.NoError(t, err)

	require.Equal(t, record.initialNonce, initialNonce)
	require.Equal(t, record.virtualBalance.initialBalance, big.NewInt(0))
	require.Equal(t, record.virtualBalance.consumedBalance, big.NewInt(0))
}

func Test_getInitialNonce(t *testing.T) {
	t.Parallel()

	expectedInitialNonce := core.OptionalUint64{
		Value:    uint64(1),
		HasValue: true,
	}

	record, err := newVirtualAccountRecord(expectedInitialNonce, big.NewInt(0))
	require.NoError(t, err)

	initialNonce, err := record.getInitialNonce()
	require.NoError(t, err)
	require.Equal(t, expectedInitialNonce.Value, initialNonce)
}

func Test_getConsumedBalance(t *testing.T) {
	t.Parallel()

	initialNonce := core.OptionalUint64{
		Value:    uint64(1),
		HasValue: true,
	}

	record, err := newVirtualAccountRecord(initialNonce, big.NewInt(0))
	require.NoError(t, err)

	balance := record.getConsumedBalance()
	require.Equal(t, balance, big.NewInt(0))
}
