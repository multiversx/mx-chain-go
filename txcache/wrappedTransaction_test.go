package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestWrappedTransaction_computePricePerGasUnit(t *testing.T) {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	t.Run("only move balance gas limit", func(t *testing.T) {
		tx := createTx([]byte("a"), "a", 1).withDataLength(1).withGasLimit(51500).withGasPrice(oneBillion)
		tx.computePricePerGasUnit(txGasHandler)

		require.Equal(t, oneBillion, int(tx.PricePerGasUnitQuotient))
		require.Equal(t, 0, int(tx.PricePerGasUnitRemainder))
	})

	t.Run("move balance gas limit and execution gas limit (1)", func(t *testing.T) {
		tx := createTx([]byte("a"), "a", 1).withDataLength(1).withGasLimit(51501).withGasPrice(oneBillion)
		tx.computePricePerGasUnit(txGasHandler)

		require.Equal(t, 999980777, int(tx.PricePerGasUnitQuotient))
		require.Equal(t, 3723, int(tx.PricePerGasUnitRemainder))
	})

	t.Run("move balance gas limit and execution gas limit (2)", func(t *testing.T) {
		tx := createTx([]byte("a"), "a", 1).withDataLength(1).withGasLimit(oneMilion).withGasPrice(oneBillion)
		tx.computePricePerGasUnit(txGasHandler)

		actualFee := 51500*oneBillion + (oneMilion-51500)*oneBillion/100
		require.Equal(t, 60985000000000, actualFee)

		require.Equal(t, actualFee/oneMilion, int(tx.PricePerGasUnitQuotient))
		require.Equal(t, 0, int(tx.PricePerGasUnitRemainder))
	})
}

func TestWrappedTransaction_isTransactionMoreDesirableByProtocol(t *testing.T) {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	t.Run("decide by price per unit", func(t *testing.T) {
		a := createTx([]byte("a-1"), "a", 1).withDataLength(1).withGasLimit(51500).withGasPrice(oneBillion)
		a.computePricePerGasUnit(txGasHandler)

		b := createTx([]byte("b-1"), "b", 1).withDataLength(1).withGasLimit(51501).withGasPrice(oneBillion)
		b.computePricePerGasUnit(txGasHandler)

		require.True(t, a.isTransactionMoreDesirableByProtocol(b))
	})
}
