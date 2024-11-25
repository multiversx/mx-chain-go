package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestWrappedTransaction_precomputeFields(t *testing.T) {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	t.Run("only move balance gas limit", func(t *testing.T) {
		tx := createTx([]byte("a"), "a", 1).withDataLength(1).withGasLimit(51500).withGasPrice(oneBillion)
		tx.precomputeFields(txGasHandler)

		require.Equal(t, "51500000000000", tx.Fee.String())
		require.Equal(t, oneBillion, int(tx.PricePerUnit))
	})

	t.Run("move balance gas limit and execution gas limit (a)", func(t *testing.T) {
		tx := createTx([]byte("b"), "b", 1).withDataLength(1).withGasLimit(51501).withGasPrice(oneBillion)
		tx.precomputeFields(txGasHandler)

		require.Equal(t, "51500010000000", tx.Fee.String())
		require.Equal(t, 999_980_777, int(tx.PricePerUnit))
	})

	t.Run("move balance gas limit and execution gas limit (b)", func(t *testing.T) {
		tx := createTx([]byte("c"), "c", 1).withDataLength(1).withGasLimit(oneMilion).withGasPrice(oneBillion)
		tx.precomputeFields(txGasHandler)

		actualFee := 51500*oneBillion + (oneMilion-51500)*oneBillion/100
		require.Equal(t, "60985000000000", tx.Fee.String())
		require.Equal(t, 60_985_000_000_000, actualFee)
		require.Equal(t, actualFee/oneMilion, int(tx.PricePerUnit))
	})

	t.Run("with guardian", func(t *testing.T) {
		tx := createTx([]byte("a"), "a", 1)
		tx.precomputeFields(txGasHandler)

		require.Equal(t, "50000000000000", tx.Fee.String())
		require.Equal(t, oneBillion, int(tx.PricePerUnit))
	})
}

func TestWrappedTransaction_isTransactionMoreValuableForNetwork(t *testing.T) {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()

	t.Run("decide by price per unit", func(t *testing.T) {
		a := createTx([]byte("a-1"), "a", 1).withDataLength(1).withGasLimit(51500).withGasPrice(oneBillion)
		a.precomputeFields(txGasHandler)

		b := createTx([]byte("b-1"), "b", 1).withDataLength(1).withGasLimit(51501).withGasPrice(oneBillion)
		b.precomputeFields(txGasHandler)

		require.True(t, a.isTransactionMoreValuableForNetwork(b))
	})

	t.Run("decide by transaction hash (set them up to have the same PPU)", func(t *testing.T) {
		a := createTx([]byte("a-7"), "a", 7)
		a.precomputeFields(txGasHandler)

		b := createTx([]byte("b-7"), "b", 7)
		b.precomputeFields(txGasHandler)

		require.True(t, a.isTransactionMoreValuableForNetwork(b))
	})
}
