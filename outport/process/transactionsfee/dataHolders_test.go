package transactionsfee

import (
	"encoding/hex"
	"math/big"
	"testing"

	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestTransactionsAndScrsHolder(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	scrHash1 := []byte("scrHash1")
	scrHash2 := []byte("scrHash2")
	scrHash3 := []byte("scrHash3")
	pool := &outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			hex.EncodeToString(txHash): {
				Transaction: &transaction.Transaction{
					Nonce: 1,
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		SmartContractResults: map[string]*outportcore.SCRInfo{
			hex.EncodeToString(scrHash1): {
				SmartContractResult: &smartContractResult.SmartContractResult{
					Nonce:          2,
					OriginalTxHash: txHash,
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},

			hex.EncodeToString(scrHash2): {
				SmartContractResult: &smartContractResult.SmartContractResult{},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			hex.EncodeToString(scrHash3): {
				SmartContractResult: &smartContractResult.SmartContractResult{
					Nonce:          3,
					OriginalTxHash: txHash,
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		Logs: []*outportcore.LogData{
			{
				Log:    &transaction.Log{Address: []byte("addr")},
				TxHash: hex.EncodeToString(txHash),
			},
			{
				Log:    &transaction.Log{},
				TxHash: "hash",
			},
		},
	}

	res := prepareTransactionsAndScrs(pool)
	require.NotNil(t, res)
	require.Equal(t, 1, len(res.txsWithResults))
	require.Equal(t, 2, len(res.txsWithResults[hex.EncodeToString(txHash)].scrs))
	require.NotNil(t, res.txsWithResults[hex.EncodeToString(txHash)].log)
	require.Equal(t, 1, len(res.scrsNoTx))
}
