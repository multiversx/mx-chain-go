package preprocess_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/assert"
)

func TestNewGasComputation_NilEconomicsFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		nil,
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	assert.Nil(t, gc)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewGasComputation_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	assert.NotNil(t, gc)
	assert.Nil(t, err)
}

func TestGasProvided_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	gc.SetGasProvided(2, []byte("hash1"))
	assert.Equal(t, uint64(2), gc.GasProvided([]byte("hash1")))

	gc.SetGasProvided(3, []byte("hash2"))
	assert.Equal(t, uint64(3), gc.GasProvided([]byte("hash2")))

	assert.Equal(t, uint64(5), gc.TotalGasProvided())

	gc.RemoveGasProvided([][]byte{[]byte("hash1")})
	assert.Equal(t, uint64(3), gc.TotalGasProvided())

	gc.Init()
	assert.Equal(t, uint64(0), gc.TotalGasProvided())
}

func TestGasRefunded_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
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

func TestGasPenalized_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	gc.SetGasPenalized(2, []byte("hash1"))
	assert.Equal(t, uint64(2), gc.GasPenalized([]byte("hash1")))

	gc.SetGasPenalized(3, []byte("hash2"))
	assert.Equal(t, uint64(3), gc.GasPenalized([]byte("hash2")))

	assert.Equal(t, uint64(5), gc.TotalGasPenalized())

	gc.RemoveGasPenalized([][]byte{[]byte("hash1")})
	assert.Equal(t, uint64(3), gc.TotalGasPenalized())

	gc.Init()
	assert.Equal(t, uint64(0), gc.TotalGasPenalized())
}

func TestComputeGasProvidedByTx_ShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	_, _, err := gc.ComputeGasProvidedByTx(0, 1, nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsNotASmartContract(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	tx := transaction.Transaction{GasLimit: 7}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(6), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractInShard(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.SCInvoking, process.SCInvoking
			}},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 0, &tx)
	assert.Equal(t, uint64(7), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractCrossShard(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			}},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldReturnZeroIf0GasLimit(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			}},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 0, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &scr)
	assert.Equal(t, uint64(0), gasInSnd)
	assert.Equal(t, uint64(0), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldReturnGasLimitIfLessThanMoveBalance(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			}},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 3, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &scr)
	assert.Equal(t, uint64(3), gasInSnd)
	assert.Equal(t, uint64(3), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldReturnGasLimitWhenRelayed(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.RelayedTx, process.RelayedTx
			}},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 3, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &scr)
	assert.Equal(t, uint64(3), gasInSnd)
	assert.Equal(t, uint64(3), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldReturnGasLimitWhenRelayedV2(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.RelayedTxV2, process.RelayedTxV2
			}},
		&epochNotifier.EpochNotifierStub{},
		0,
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 3, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &scr)
	assert.Equal(t, uint64(3), gasInSnd)
	assert.Equal(t, uint64(3), gasInRcv)
}

func TestComputeGasProvidedByMiniBlock_ShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
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

	_, _, err := gc.ComputeGasProvidedByMiniBlock(&miniBlock, mapHashTx)
	assert.Equal(t, process.ErrMissingTransaction, err)
}

func TestComputeGasProvidedByMiniBlock_ShouldReturnZeroWhenOneTxIsMissing(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
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

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByMiniBlock(&miniBlock, mapHashTx)
	assert.Equal(t, uint64(0), gasInSnd)
	assert.Equal(t, uint64(0), gasInRcv)
}

func TestComputeGasProvidedByMiniBlock_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if core.IsSmartContractAddress(tx.GetRcvAddr()) {
					return process.MoveBalance, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			}},
		&epochNotifier.EpochNotifierStub{},
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

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByMiniBlock(&miniBlock, mapHashTx)
	assert.Equal(t, uint64(18), gasInSnd)
	assert.Equal(t, uint64(56), gasInRcv)
}

func TestComputeGasProvidedByMiniBlock_ShouldWorkV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				if core.IsSmartContractAddress(tx.GetRcvAddr()) {
					return process.SCInvoking, process.SCInvoking
				}
				return process.MoveBalance, process.MoveBalance
			}},
		&epochNotifier.EpochNotifierStub{},
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

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByMiniBlock(&miniBlock, mapHashTx)
	assert.Equal(t, uint64(18), gasInSnd)
	assert.Equal(t, uint64(44), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsNotASmartContractV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{},
		&epochNotifier.EpochNotifierStub{},
		10,
	)

	tx := transaction.Transaction{GasLimit: 7}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(6), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractInShardV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.SCInvoking, process.SCInvoking
			}},
		&epochNotifier.EpochNotifierStub{},
		10,
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 0, &tx)
	assert.Equal(t, uint64(7), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractCrossShardV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&mock.FeeHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.SCInvoking, process.SCInvoking
			}},
		&epochNotifier.EpochNotifierStub{},
		10,
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(1), gasInRcv)
}
