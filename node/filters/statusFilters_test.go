package filters

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestShardAPIBlockProcessor_SetStatusESDTTransfer(t *testing.T) {
	t.Parallel()

	sf := NewStatusFilters(0)

	esdtTransferTx := &transaction.ApiTransactionResult{
		Hash:             "myHash",
		Nonce:            1,
		SourceShard:      1,
		DestinationShard: 0,
		Data:             []byte("ESDTTransfer@42524f2d343663663439@a688906bd8b00000"),
	}
	mbs := []*api.MiniBlock{
		{
			SourceShard:      1,
			DestinationShard: 0,
			Transactions: []*transaction.ApiTransactionResult{
				esdtTransferTx,
				{},
			},
			Type: block.TxBlock.String(),
		},
		{
			Type: block.TxBlock.String(),
		},
		{
			DestinationShard: 1,
			SourceShard:      0,
			Type:             block.SmartContractResultBlock.String(),
			Transactions: []*transaction.ApiTransactionResult{
				{},
				{
					OriginalTransactionHash: "myHash",
					Nonce:                   1,
					SourceShard:             1,
					DestinationShard:        0,
					Data:                    []byte("ESDTTransfer@42524f2d343663663439@a688906bd8b00000@75736572206572726f72"),
				},
			},
		},
		{
			Type: block.RewardsBlock.String(),
		},
	}
	sf.ApplyStatusFilters(mbs)
	require.Equal(t, transaction.TxStatusFail, esdtTransferTx.Status)
}

func TestSetStatusIfIsESDTTransferFail(t *testing.T) {
	t.Parallel()

	sf := NewStatusFilters(0)
	// ESDT transfer fail
	tx1 := &transaction.ApiTransactionResult{
		Nonce:            1,
		Hash:             "myHash",
		SourceShard:      1,
		DestinationShard: 0,
		Data:             []byte("ESDTTransfer@42524f2d343663663439@a688906bd8b00000"),
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{
				OriginalTxHash: "myHash",
				Nonce:          1,
				Data:           "ESDTTransfer@42524f2d343663663439@a688906bd8b00000@75736572206572726f72",
			},
		},
	}

	sf.SetStatusIfIsFailedESDTTransfer(tx1)
	require.Equal(t, transaction.TxStatusFail, tx1.Status)

	// transaction with no SCR should be ignored
	tx2 := &transaction.ApiTransactionResult{
		Status: transaction.TxStatusSuccess,
	}
	sf.SetStatusIfIsFailedESDTTransfer(tx2)
	require.Equal(t, transaction.TxStatusSuccess, tx2.Status)

	// intra shard transaction should be ignored
	tx3 := &transaction.ApiTransactionResult{
		Status: transaction.TxStatusSuccess,
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{},
			{},
		},
	}
	sf.SetStatusIfIsFailedESDTTransfer(tx3)
	require.Equal(t, transaction.TxStatusSuccess, tx3.Status)

	// no ESDT transfer should be ignored
	tx4 := &transaction.ApiTransactionResult{
		Status:           transaction.TxStatusSuccess,
		SourceShard:      1,
		DestinationShard: 0,
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{},
			{},
		},
	}
	sf.SetStatusIfIsFailedESDTTransfer(tx4)
	require.Equal(t, transaction.TxStatusSuccess, tx4.Status)
}
