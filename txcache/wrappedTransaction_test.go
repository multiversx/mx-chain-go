package txcache

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func TestWrappedTransaction_precomputeFields(t *testing.T) {
	t.Run("only move balance gas limit", func(t *testing.T) {
		host := txcachemocks.NewMempoolHostMock()

		tx := createTx([]byte("a"), "a", 1).withValue(oneQuintillionBig).withDataLength(1).withGasLimit(51500).withGasPrice(oneBillion)
		tx.precomputeFields(host)

		require.Equal(t, "51500000000000", tx.Fee.String())
		require.Equal(t, oneBillion, int(tx.PricePerUnit))
		require.Equal(t, "1000000000000000000", tx.TransferredValue.String())
		require.Equal(t, []byte("a"), tx.FeePayer)
	})

	t.Run("move balance gas limit and execution gas limit (a)", func(t *testing.T) {
		host := txcachemocks.NewMempoolHostMock()

		tx := createTx([]byte("b"), "b", 1).withDataLength(1).withGasLimit(51501).withGasPrice(oneBillion)
		tx.precomputeFields(host)

		require.Equal(t, "51500010000000", tx.Fee.String())
		require.Equal(t, 999_980_777, int(tx.PricePerUnit))
		require.Equal(t, []byte("b"), tx.FeePayer)
	})

	t.Run("move balance gas limit and execution gas limit (b)", func(t *testing.T) {
		host := txcachemocks.NewMempoolHostMock()

		tx := createTx([]byte("c"), "c", 1).withDataLength(1).withGasLimit(oneMilion).withGasPrice(oneBillion)
		tx.precomputeFields(host)

		actualFee := 51500*oneBillion + (oneMilion-51500)*oneBillion/100
		require.Equal(t, "60985000000000", tx.Fee.String())
		require.Equal(t, 60_985_000_000_000, actualFee)
		require.Equal(t, actualFee/oneMilion, int(tx.PricePerUnit))
		require.Equal(t, []byte("c"), tx.FeePayer)
	})

	t.Run("with guardian", func(t *testing.T) {
		host := txcachemocks.NewMempoolHostMock()

		tx := createTx([]byte("a"), "a", 1).withValue(oneQuintillionBig)
		tx.precomputeFields(host)

		require.Equal(t, "50000000000000", tx.Fee.String())
		require.Equal(t, oneBillion, int(tx.PricePerUnit))
		require.Equal(t, "1000000000000000000", tx.TransferredValue.String())
		require.Equal(t, []byte("a"), tx.FeePayer)
	})

	t.Run("with nil transferred value", func(t *testing.T) {
		host := txcachemocks.NewMempoolHostMock()

		tx := createTx([]byte("a"), "a", 1)
		tx.precomputeFields(host)

		require.Nil(t, tx.TransferredValue)
		require.Equal(t, []byte("a"), tx.FeePayer)
	})

	t.Run("queries host", func(t *testing.T) {
		host := txcachemocks.NewMempoolHostMock()
		host.ComputeTxFeeCalled = func(_ data.TransactionWithFeeHandler) *big.Int {
			return big.NewInt(42)
		}
		host.GetTransferredValueCalled = func(_ data.TransactionHandler) *big.Int {
			return big.NewInt(43)
		}

		tx := createTx([]byte("a"), "a", 1).withGasLimit(50_000)
		tx.precomputeFields(host)

		require.Equal(t, "42", tx.Fee.String())
		require.Equal(t, "43", tx.TransferredValue.String())
	})
}

func TestWrappedTransaction_isTransactionMoreValuableForNetwork(t *testing.T) {
	host := txcachemocks.NewMempoolHostMock()

	t.Run("decide by price per unit", func(t *testing.T) {
		a := createTx([]byte("a-1"), "a", 1).withDataLength(1).withGasLimit(51500).withGasPrice(oneBillion)
		a.precomputeFields(host)

		b := createTx([]byte("b-1"), "b", 1).withDataLength(1).withGasLimit(51501).withGasPrice(oneBillion)
		b.precomputeFields(host)

		require.True(t, a.isTransactionMoreValuableForNetwork(b))
	})

	t.Run("decide by gas limit (set them up to have the same PPU)", func(t *testing.T) {
		a := createTx([]byte("a-7"), "a", 7).withDataLength(30).withGasLimit(95_000).withGasPrice(oneBillion)
		a.precomputeFields(host)

		b := createTx([]byte("b-7"), "b", 7).withDataLength(60).withGasLimit(140_000).withGasPrice(oneBillion)
		b.precomputeFields(host)

		require.Equal(t, a.PricePerUnit, b.PricePerUnit)
		require.True(t, b.isTransactionMoreValuableForNetwork(a))
	})

	t.Run("decide by transaction hash (set them up to have the same PPU and gas limit)", func(t *testing.T) {
		a := createTx([]byte("a-7"), "a", 7)
		a.precomputeFields(host)

		b := createTx([]byte("b-7"), "b", 7)
		b.precomputeFields(host)

		require.Equal(t, a.PricePerUnit, b.PricePerUnit)
		require.True(t, a.isTransactionMoreValuableForNetwork(b))
	})
}
