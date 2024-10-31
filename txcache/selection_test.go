package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_mergeTwoBunches(t *testing.T) {
	t.Run("empty bunches", func(t *testing.T) {
		merged := mergeTwoBunches(BunchOfTransactions{}, BunchOfTransactions{})
		require.Len(t, merged, 0)
	})

	t.Run("alice and bob (1)", func(t *testing.T) {
		first := BunchOfTransactions{
			createTx([]byte("hash-alice-1"), "alice", 1).withGasPrice(42),
		}

		second := BunchOfTransactions{
			createTx([]byte("hash-bob-1"), "bob", 1).withGasPrice(43),
		}

		merged := mergeTwoBunches(first, second)

		require.Len(t, merged, 2)
		require.Equal(t, "hash-bob-1", string(merged[0].TxHash))
		require.Equal(t, "hash-alice-1", string(merged[1].TxHash))
	})
}
