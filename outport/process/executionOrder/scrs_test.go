package executionOrder

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func newScr(nonce uint64, originalTxHash string, execOrder int) data.TransactionHandlerWithGasUsedAndFee {
	return &outport.TransactionHandlerWithGasAndFee{
		TransactionHandler: &smartContractResult.SmartContractResult{
			Nonce:          nonce,
			OriginalTxHash: []byte(originalTxHash),
		},
		ExecutionOrder: execOrder,
	}
}

func TestSetOrderSmartContractResults(t *testing.T) {
	t.Parallel()

	txHash, txHashNotInPool, scrHash1, scrsHash2, scrsHash3 := "tx", "txHashNotInPool", "scr1", "scr2", "scr3"
	pool := &outport.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			txHash: &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{}, ExecutionOrder: 1},
		},
		Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			scrHash1:  newScr(0, txHash, 0),
			scrsHash2: newScr(1, txHashNotInPool, 0),
			scrsHash3: newScr(2, txHashNotInPool, 2),
		},
	}

	setOrderSmartContractResults(pool, []*block.MiniBlock{})

	require.Equal(t, &outport.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			txHash: &outport.TransactionHandlerWithGasAndFee{TransactionHandler: &transaction.Transaction{}, ExecutionOrder: 1},
		},
		Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			scrHash1:  newScr(0, txHash, 1),
			scrsHash2: newScr(1, txHashNotInPool, 2),
			scrsHash3: newScr(2, txHashNotInPool, 2),
		},
	}, pool)
}
