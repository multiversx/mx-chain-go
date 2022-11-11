package transactionsfee

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestTransactionsAndScrsHolder(t *testing.T) {
	t.Parallel()

	txHash := "txHash"
	scrHash1 := "scrHash1"
	scrHash2 := "scrHash2"
	scrHash3 := "scrHash3"
	pool := &outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			txHash: outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				Nonce: 1,
			}, 0, big.NewInt(0)),
		},
		Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			scrHash1: outportcore.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				Nonce:          2,
				OriginalTxHash: []byte(txHash),
			}, 0, big.NewInt(0)),

			scrHash2: outportcore.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{}, 0, big.NewInt(0)),
			scrHash3: outportcore.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				Nonce:          3,
				OriginalTxHash: []byte(txHash),
			}, 0, big.NewInt(0)),
		},
		Logs: []*data.LogData{
			{
				TxHash: "hash",
			},
			{
				TxHash: txHash,
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
				},
			},
		},
	}

	res := prepareTransactionsAndScrs(pool)
	require.NotNil(t, res)
	require.Equal(t, 1, len(res.txsWithResults))
	require.Equal(t, 2, len(res.txsWithResults[txHash].scrs))
	require.NotNil(t, res.txsWithResults[txHash].log)
	require.Equal(t, 1, len(res.scrsNoTx))
}
