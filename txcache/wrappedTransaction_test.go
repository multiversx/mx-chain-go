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

		require.Equal(t, float64(oneBillion), tx.PricePerUnit)
	})

	t.Run("move balance gas limit and execution gas limit (1)", func(t *testing.T) {
		tx := createTx([]byte("a"), "a", 1).withDataLength(1).withGasLimit(51501).withGasPrice(oneBillion)
		tx.computePricePerGasUnit(txGasHandler)

		require.InDelta(t, float64(999980777), tx.PricePerUnit, 0.1)
	})

	t.Run("move balance gas limit and execution gas limit (2)", func(t *testing.T) {
		tx := createTx([]byte("a"), "a", 1).withDataLength(1).withGasLimit(oneMilion).withGasPrice(oneBillion)
		tx.computePricePerGasUnit(txGasHandler)

		actualFee := 51500*oneBillion + (oneMilion-51500)*oneBillion/100
		require.Equal(t, 60985000000000, actualFee)

		require.InDelta(t, actualFee/oneMilion, tx.PricePerUnit, 0.1)
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

	t.Run("decide by gas price (set them up to have the same PPU)", func(t *testing.T) {
		a := createTx([]byte("a-2"), "a", 1).withGasPrice(oneBillion + 1)
		b := createTx([]byte("b-2"), "b", 1).withGasPrice(oneBillion)

		a.PricePerUnit = 42
		b.PricePerUnit = 42

		require.True(t, a.isTransactionMoreDesirableByProtocol(b))
	})

	t.Run("decide by gas limit (set them up to have the same PPU and gas price)", func(t *testing.T) {
		a := createTx([]byte("a-2"), "a", 1).withGasLimit(55000)
		b := createTx([]byte("b-2"), "b", 1).withGasLimit(60000)

		a.PricePerUnit = 42
		b.PricePerUnit = 42

		require.True(t, a.isTransactionMoreDesirableByProtocol(b))
	})

	t.Run("decide by transaction hash (set them up to have the same PPU, gas price and gas limit)", func(t *testing.T) {
		a := createTx([]byte("a-2"), "a", 1)
		b := createTx([]byte("b-2"), "b", 1)

		require.True(t, b.isTransactionMoreDesirableByProtocol(a))
	})
}
