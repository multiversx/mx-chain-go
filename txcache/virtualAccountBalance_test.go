package txcache

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_validateBalance(t *testing.T) {
	t.Parallel()

	t.Run("should return errExceededBalance", func(t *testing.T) {
		t.Parallel()

		record, err := newVirtualAccountBalance(
			big.NewInt(2),
		)
		require.NoError(t, err)

		record.consumedBalance = big.NewInt(3)
		err = record.validateBalance()
		require.Equal(t, errExceededBalance, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		record, err := newVirtualAccountBalance(
			big.NewInt(2),
		)
		require.NoError(t, err)

		record.consumedBalance = big.NewInt(1)
		err = record.validateBalance()
		require.Nil(t, err)
	})
}
