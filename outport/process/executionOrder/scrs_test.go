package executionOrder

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func newScr(nonce uint64, originalTxHash string, execOrder uint32) *outport.SCRInfo {
	return &outport.SCRInfo{
		SmartContractResult: &smartContractResult.SmartContractResult{
			Nonce:          nonce,
			OriginalTxHash: []byte(originalTxHash),
		},
		ExecutionOrder: execOrder,
	}
}

func TestSetOrderSmartContractResults(t *testing.T) {
	t.Parallel()

	txHash, txHashNotInPool, scrHash1, scrsHash2, scrsHash3, scrHashToMe := "tx", "txHashNotInPool", "scr1", "scr2", "scr3", "scrHashToMe"
	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			txHash: {Transaction: &transaction.Transaction{}, ExecutionOrder: 1},
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			scrHash1:    newScr(0, txHash, 0),
			scrsHash2:   newScr(1, txHashNotInPool, 0),
			scrsHash3:   newScr(2, txHashNotInPool, 2),
			scrHashToMe: newScr(3, txHashNotInPool, 1),
		},
	}

	setOrderSmartContractResults(pool, []*block.MiniBlock{}, map[string]data.TxWithExecutionOrderHandler{
		scrHashToMe: newScr(3, txHashNotInPool, 1)})

	require.Equal(t, &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			txHash: {Transaction: &transaction.Transaction{}, ExecutionOrder: 1},
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			scrHash1:    newScr(0, txHash, 1),
			scrsHash2:   newScr(1, txHashNotInPool, 2),
			scrsHash3:   newScr(2, txHashNotInPool, 2),
			scrHashToMe: newScr(3, txHashNotInPool, 1),
		},
	}, pool)
}
