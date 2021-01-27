package preprocess_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewGasConsumption_NilEconomicsFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		nil,
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		0,
	)

	assert.Nil(t, gc)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewGasConsumption_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		0,
	)

	assert.NotNil(t, gc)
	assert.Nil(t, err)
}

func TestGasConsumed_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		0,
	)

	gc.SetGasConsumed(2, []byte("hash1"))
	assert.Equal(t, uint64(2), gc.GasConsumed([]byte("hash1")))

	gc.SetGasConsumed(3, []byte("hash2"))
	assert.Equal(t, uint64(3), gc.GasConsumed([]byte("hash2")))

	assert.Equal(t, uint64(5), gc.TotalGasConsumed())

	gc.RemoveGasConsumed([][]byte{[]byte("hash1")})
	assert.Equal(t, uint64(3), gc.TotalGasConsumed())

	gc.Init()
	assert.Equal(t, uint64(0), gc.TotalGasConsumed())
}

func TestGasRefunded_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		0,
	)

	gc.SetGasRefunded(2, []byte("hash1"))
	assert.Equal(t, uint64(2), gc.GasRefunded([]byte("hash1")))

	gc.SetGasRefunded(3, []byte("hash2"))
	assert.Equal(t, uint64(3), gc.GasRefunded([]byte("hash2")))

	assert.Equal(t, uint64(5), gc.TotalGasRefunded())

	gc.RemoveGasRefunded([][]byte{[]byte("hash1")})
	assert.Equal(t, uint64(3), gc.TotalGasRefunded())

	gc.Init()
	assert.Equal(t, uint64(0), gc.TotalGasRefunded())
}

func TestComputeGasConsumedByTx_ShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		0,
	)

	_, _, err := gc.ComputeGasConsumedByTx(0, 1, nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestComputeGasConsumedByTx_ShouldWorkWhenTxReceiverAddressIsNotASmartContract(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		0,
	)

	tx := transaction.Transaction{GasLimit: 7}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(6), gasInRcv)
}

func TestComputeGasConsumedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractInShard(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		}},
		&mock.EpochNotifierStub{},
		0,
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByTx(0, 0, &tx)
	assert.Equal(t, uint64(7), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasConsumedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractCrossShard(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.SCInvoking
		}},
		&mock.EpochNotifierStub{},
		0,
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasConsumedByTx_ShouldReturnZeroIf0GasLimit(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.SCInvoking
		}},
		&mock.EpochNotifierStub{},
		0,
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 0, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByTx(0, 1, &scr)
	assert.Equal(t, uint64(0), gasInSnd)
	assert.Equal(t, uint64(0), gasInRcv)
}

func TestComputeGasConsumedByTx_ShouldReturnGasLimitIfLessThanMoveBalance(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.SCInvoking
		}},
		&mock.EpochNotifierStub{},
		0,
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 3, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByTx(0, 1, &scr)
	assert.Equal(t, uint64(3), gasInSnd)
	assert.Equal(t, uint64(3), gasInRcv)
}

func TestComputeGasConsumedByMiniBlock_ShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		0,
	)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, []byte("hash1"))
	txHashes = append(txHashes, []byte("hash2"))

	miniBlock := block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 1,
		TxHashes:        txHashes,
	}

	mapHashTx := make(map[string]data.TransactionHandler)

	_, _, err := gc.ComputeGasConsumedByMiniBlock(&miniBlock, mapHashTx)
	assert.Equal(t, process.ErrMissingTransaction, err)
}

func TestComputeGasConsumedByMiniBlock_ShouldReturnZeroWhenOneTxIsMissing(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		0,
	)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, []byte("hash1"))
	txHashes = append(txHashes, []byte("hash2"))

	miniBlock := block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 1,
		TxHashes:        txHashes,
	}

	mapHashTx := make(map[string]data.TransactionHandler)
	mapHashTx["hash1"] = nil
	mapHashTx["hash2"] = nil

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByMiniBlock(&miniBlock, mapHashTx)
	assert.Equal(t, uint64(0), gasInSnd)
	assert.Equal(t, uint64(0), gasInRcv)
}

func TestComputeGasConsumedByMiniBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			if core.IsSmartContractAddress(tx.GetRcvAddr()) {
				return process.MoveBalance, process.SCInvoking
			}
			return process.MoveBalance, process.MoveBalance
		}},
		&mock.EpochNotifierStub{},
		0,
	)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, []byte("hash1"))
	txHashes = append(txHashes, []byte("hash2"))
	txHashes = append(txHashes, []byte("hash3"))

	miniBlock := block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 1,
		TxHashes:        txHashes,
	}

	mapHashTx := make(map[string]data.TransactionHandler)
	mapHashTx["hash1"] = &transaction.Transaction{GasLimit: 7}
	mapHashTx["hash2"] = &transaction.Transaction{GasLimit: 20, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}
	mapHashTx["hash3"] = &transaction.Transaction{GasLimit: 30, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByMiniBlock(&miniBlock, mapHashTx)
	assert.Equal(t, uint64(18), gasInSnd)
	assert.Equal(t, uint64(56), gasInRcv)
}

func TestComputeGasConsumedByMiniBlock_ShouldWorkV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			if core.IsSmartContractAddress(tx.GetRcvAddr()) {
				return process.SCInvoking, process.SCInvoking
			}
			return process.MoveBalance, process.MoveBalance
		}},
		&mock.EpochNotifierStub{},
		10,
	)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, []byte("hash1"))
	txHashes = append(txHashes, []byte("hash2"))
	txHashes = append(txHashes, []byte("hash3"))

	miniBlock := block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 1,
		TxHashes:        txHashes,
	}

	mapHashTx := make(map[string]data.TransactionHandler)
	mapHashTx["hash1"] = &transaction.Transaction{GasLimit: 7}
	mapHashTx["hash2"] = &transaction.Transaction{GasLimit: 20, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}
	mapHashTx["hash3"] = &transaction.Transaction{GasLimit: 30, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByMiniBlock(&miniBlock, mapHashTx)
	assert.Equal(t, uint64(18), gasInSnd)
	assert.Equal(t, uint64(44), gasInRcv)
}

func TestComputeGasConsumedByTx_ShouldWorkWhenTxReceiverAddressIsNotASmartContractV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{},
		&mock.EpochNotifierStub{},
		10,
	)

	tx := transaction.Transaction{GasLimit: 7}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(6), gasInRcv)
}

func TestComputeGasConsumedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractInShardV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		}},
		&mock.EpochNotifierStub{},
		10,
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByTx(0, 0, &tx)
	assert.Equal(t, uint64(7), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasConsumedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractCrossShardV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&mock.TxTypeHandlerMock{ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		}},
		&mock.EpochNotifierStub{},
		10,
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasConsumedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(1), gasInRcv)
}
