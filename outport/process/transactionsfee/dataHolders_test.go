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

	txHash := "txHash"
	scrHash1 := "scrHash1"
	scrHash2 := "scrHash2"
	scrHash3 := "scrHash3"
	pool := &outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			hex.EncodeToString([]byte(txHash)): {
				Transaction: &transaction.Transaction{
					Nonce: 1,
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		SmartContractResults: map[string]*outportcore.SCRInfo{
			hex.EncodeToString([]byte(scrHash1)): {
				SmartContractResult: &smartContractResult.SmartContractResult{
					Nonce:          2,
					OriginalTxHash: []byte(txHash),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},

			hex.EncodeToString([]byte(scrHash2)): {
				SmartContractResult: &smartContractResult.SmartContractResult{},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			hex.EncodeToString([]byte(scrHash3)): {
				SmartContractResult: &smartContractResult.SmartContractResult{
					Nonce:          3,
					OriginalTxHash: []byte(txHash),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		Logs: []*outportcore.LogData{
			{
				Log:    &transaction.Log{Address: []byte("addr")},
				TxHash: hex.EncodeToString([]byte(txHash)),
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
	require.Equal(t, 2, len(res.txsWithResults[hex.EncodeToString([]byte(txHash))].scrs))
	require.NotNil(t, res.txsWithResults[hex.EncodeToString([]byte(txHash))].log)
	require.Equal(t, 1, len(res.scrsNoTx))
}
