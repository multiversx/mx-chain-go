package filters

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestStatusFilters_ApplyStatusFilters(t *testing.T) {
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

func TestStatusFilters_ApplyStatusFiltersFailedMoveBalance(t *testing.T) {
	t.Parallel()

	userAddr1 := bytes.Repeat([]byte{0x01}, 32)
	userAddr2 := bytes.Repeat([]byte{0x02}, 32)
	scAddr := append(bytes.Repeat([]byte{0x00}, 8), bytes.Repeat([]byte{0x01}, 24)...)

	const senderStr = "sndAddr"
	const receiverStr = "rcvAddr"
	const txHash = "mbTxHash"

	createMoveBalanceTx := func(value *big.Int, sndAddr, rcvAddr []byte) *transaction.ApiTransactionResult {
		return &transaction.ApiTransactionResult{
			Hash:             txHash,
			Nonce:            1,
			SourceShard:      1,
			DestinationShard: 0,
			Status:           transaction.TxStatusSuccess,
			Sender:           senderStr,
			Receiver:         receiverStr,
			Value:            value.String(),
			Tx: &transaction.Transaction{
				Nonce:   1,
				Value:   new(big.Int).Set(value),
				SndAddr: sndAddr,
				RcvAddr: rcvAddr,
			},
		}
	}

	createRefundSCR := func(value *big.Int, sndStr, rcvStr string, originalHash string) *transaction.ApiTransactionResult {
		return &transaction.ApiTransactionResult{
			Hash:                    "scrHash",
			OriginalTransactionHash: originalHash,
			SourceShard:             0,
			DestinationShard:        1,
			Type:                    string(transaction.TxTypeUnsigned),
			Sender:                  sndStr,
			Receiver:                rcvStr,
			Value:                   value.String(),
			Tx: &smartContractResult.SmartContractResult{
				Value: new(big.Int).Set(value),
			},
		}
	}

	createMbs := func(tx *transaction.ApiTransactionResult, scrs ...*transaction.ApiTransactionResult) []*api.MiniBlock {
		return []*api.MiniBlock{
			{
				SourceShard:      1,
				DestinationShard: 0,
				Transactions:     []*transaction.ApiTransactionResult{tx},
				Type:             block.TxBlock.String(),
			},
			{
				SourceShard:      0,
				DestinationShard: 1,
				Transactions:     scrs,
				Type:             block.SmartContractResultBlock.String(),
			},
		}
	}

	t.Run("mirrored refund with same value should fail", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createMoveBalanceTx(big.NewInt(100), userAddr1, userAddr2)
		mbs := createMbs(tx, createRefundSCR(big.NewInt(100), receiverStr, senderStr, txHash))

		sf.ApplyStatusFilters(mbs)
		require.Equal(t, transaction.TxStatusFail, tx.Status)
	})

	t.Run("scr value differs should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createMoveBalanceTx(big.NewInt(100), userAddr1, userAddr2)
		mbs := createMbs(tx, createRefundSCR(big.NewInt(99), receiverStr, senderStr, txHash))

		sf.ApplyStatusFilters(mbs)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("scr not mirrored should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createMoveBalanceTx(big.NewInt(100), userAddr1, userAddr2)
		mbs := createMbs(tx, createRefundSCR(big.NewInt(100), senderStr, receiverStr, txHash))

		sf.ApplyStatusFilters(mbs)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("scr for another tx should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createMoveBalanceTx(big.NewInt(100), userAddr1, userAddr2)
		mbs := createMbs(tx, createRefundSCR(big.NewInt(100), receiverStr, senderStr, "otherHash"))

		sf.ApplyStatusFilters(mbs)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("zero value tx should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createMoveBalanceTx(big.NewInt(0), userAddr1, userAddr2)
		mbs := createMbs(tx, createRefundSCR(big.NewInt(0), receiverStr, senderStr, txHash))

		sf.ApplyStatusFilters(mbs)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("sc sender should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createMoveBalanceTx(big.NewInt(100), scAddr, userAddr2)
		mbs := createMbs(tx, createRefundSCR(big.NewInt(100), receiverStr, senderStr, txHash))

		sf.ApplyStatusFilters(mbs)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("no scr miniblock should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createMoveBalanceTx(big.NewInt(100), userAddr1, userAddr2)
		mbs := []*api.MiniBlock{
			{
				SourceShard:      1,
				DestinationShard: 0,
				Transactions:     []*transaction.ApiTransactionResult{tx},
				Type:             block.TxBlock.String(),
			},
		}

		sf.ApplyStatusFilters(mbs)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("non mirrored scr miniblock should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createMoveBalanceTx(big.NewInt(100), userAddr1, userAddr2)
		mbs := []*api.MiniBlock{
			{
				SourceShard:      1,
				DestinationShard: 0,
				Transactions:     []*transaction.ApiTransactionResult{tx},
				Type:             block.TxBlock.String(),
			},
			{
				SourceShard:      1,
				DestinationShard: 0,
				Transactions:     []*transaction.ApiTransactionResult{createRefundSCR(big.NewInt(100), receiverStr, senderStr, txHash)},
				Type:             block.SmartContractResultBlock.String(),
			},
		}

		sf.ApplyStatusFilters(mbs)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})
}

func TestStatusFilters_SetStatusIfIsFailedESDTTransfer(t *testing.T) {
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

func TestStatusFilters_SetStatusIfFailedMoveBalanceWithError(t *testing.T) {
	t.Parallel()

	userAddr1 := bytes.Repeat([]byte{0x01}, 32)
	userAddr2 := bytes.Repeat([]byte{0x02}, 32)
	scAddr := append(bytes.Repeat([]byte{0x00}, 8), bytes.Repeat([]byte{0x01}, 24)...)
	require.True(t, core.IsSmartContractAddress(scAddr))
	require.False(t, core.IsSmartContractAddress(userAddr1))
	require.False(t, core.IsSmartContractAddress(userAddr2))

	const senderStr = "sndAddr"
	const receiverStr = "rcvAddr"

	createTx := func(value *big.Int, sndAddr, rcvAddr []byte) *transaction.ApiTransactionResult {
		var scrValue *big.Int
		if value != nil {
			scrValue = new(big.Int).Set(value)
		}
		return &transaction.ApiTransactionResult{
			Status:   transaction.TxStatusSuccess,
			Sender:   senderStr,
			Receiver: receiverStr,
			Tx: &transaction.Transaction{
				Value:   value,
				SndAddr: sndAddr,
				RcvAddr: rcvAddr,
			},
			SmartContractResults: []*transaction.ApiSmartContractResult{
				{
					OriginalTxHash: "myHash",
					Nonce:          1,
					Value:          scrValue,
					SndAddr:        receiverStr,
					RcvAddr:        senderStr,
				},
			},
		}
	}

	t.Run("no SCRs should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)
		tx.SmartContractResults = nil

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("nil tx should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)
		tx.Tx = nil

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("nil value should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(nil, userAddr1, userAddr2)

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("zero value should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(0), userAddr1, userAddr2)

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("sender is SC should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), scAddr, userAddr2)

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("receiver is SC should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, scAddr)

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("scr value differs should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)
		tx.SmartContractResults[0].Value = big.NewInt(99)

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("scr not mirrored should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)
		tx.SmartContractResults[0].SndAddr = senderStr
		tx.SmartContractResults[0].RcvAddr = receiverStr

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("scr only sender matches should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)
		tx.SmartContractResults[0].RcvAddr = "other"

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("scr with nil value should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)
		tx.SmartContractResults[0].Value = nil

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("nil scr entry should be ignored", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)
		tx.SmartContractResults = []*transaction.ApiSmartContractResult{nil}

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusSuccess, tx.Status)
	})

	t.Run("mirrored refund among other SCRs should fail", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)
		tx.SmartContractResults = append([]*transaction.ApiSmartContractResult{
			{
				Value:   big.NewInt(1),
				SndAddr: "other",
				RcvAddr: "other",
			},
			nil,
		}, tx.SmartContractResults...)

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusFail, tx.Status)
	})

	t.Run("move balance with mirrored refund should fail", func(t *testing.T) {
		t.Parallel()

		sf := NewStatusFilters(0)
		tx := createTx(big.NewInt(100), userAddr1, userAddr2)

		sf.SetStatusIfFailedMoveBalanceWithError(tx)
		require.Equal(t, transaction.TxStatusFail, tx.Status)
	})
}
