package txcache

import (
	"testing"

	"github.com/multiversx/mx-chain-storage-go/testscommon/txcachemocks"
	"github.com/stretchr/testify/require"
)

func Test_computeTxFee(t *testing.T) {
	txGasHandler := txcachemocks.NewTxGasHandlerMock()
	tx := createTx([]byte("a"), "a", 1).withDataLength(1).withGasLimit(51500).withGasPrice(oneBillion)
	txFee := tx.computeFee(txGasHandler)

	require.Equal(t, float64(51500000000000), txFee)
	require.Equal(t, txFee, tx.TxFee)
}
