package preprocess_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createEnableEpochsHandler() common.EnableEpochsHandler {
	return &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsSCDeployFlagEnabledField: true,
	}
}

func TestNewGasComputation_NilEconomicsFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		nil,
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	assert.Nil(t, gc)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewGasComputation_NilEnableEpochsHandlerShouldErr(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		nil,
	)

	assert.Nil(t, gc)
	assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
}

func TestNewGasComputation_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, err := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	assert.NotNil(t, gc)
	assert.Nil(t, err)
}

func TestGasProvided_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	key := []byte("key")
	gc.Reset(key)

	gc.SetGasProvided(2, []byte("hash1"))
	assert.Equal(t, uint64(2), gc.GasProvided([]byte("hash1")))
	require.Equal(t, 1, len(gc.GetTxHashesWithGasProvidedSinceLastReset(key)))
	assert.Equal(t, []byte("hash1"), gc.GetTxHashesWithGasProvidedSinceLastReset(key)[0])

	gc.SetGasProvided(3, []byte("hash2"))
	assert.Equal(t, uint64(3), gc.GasProvided([]byte("hash2")))
	require.Equal(t, 2, len(gc.GetTxHashesWithGasProvidedSinceLastReset(key)))
	assert.Equal(t, []byte("hash1"), gc.GetTxHashesWithGasProvidedSinceLastReset(key)[0])
	assert.Equal(t, []byte("hash2"), gc.GetTxHashesWithGasProvidedSinceLastReset(key)[1])

	assert.Equal(t, uint64(5), gc.TotalGasProvided())

	gc.RemoveGasProvided([][]byte{[]byte("hash1")})
	assert.Equal(t, uint64(3), gc.TotalGasProvided())

	gc.Init()
	assert.Equal(t, uint64(0), gc.TotalGasProvided())
}

func TestGasRefunded_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	key := []byte("key")
	gc.Reset(key)

	gc.SetGasRefunded(2, []byte("hash1"))
	assert.Equal(t, uint64(2), gc.GasRefunded([]byte("hash1")))
	require.Equal(t, 1, len(gc.GetTxHashesWithGasRefundedSinceLastReset(key)))
	assert.Equal(t, []byte("hash1"), gc.GetTxHashesWithGasRefundedSinceLastReset(key)[0])

	gc.SetGasRefunded(3, []byte("hash2"))
	assert.Equal(t, uint64(3), gc.GasRefunded([]byte("hash2")))
	require.Equal(t, 2, len(gc.GetTxHashesWithGasRefundedSinceLastReset(key)))
	assert.Equal(t, []byte("hash1"), gc.GetTxHashesWithGasRefundedSinceLastReset(key)[0])
	assert.Equal(t, []byte("hash2"), gc.GetTxHashesWithGasRefundedSinceLastReset(key)[1])

	assert.Equal(t, uint64(5), gc.TotalGasRefunded())

	gc.RemoveGasRefunded([][]byte{[]byte("hash1")})
	assert.Equal(t, uint64(3), gc.TotalGasRefunded())

	gc.Init()
	assert.Equal(t, uint64(0), gc.TotalGasRefunded())
}

func TestGasPenalized_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	key := []byte("key")
	gc.Reset(key)

	gc.SetGasPenalized(2, []byte("hash1"))
	assert.Equal(t, uint64(2), gc.GasPenalized([]byte("hash1")))
	require.Equal(t, 1, len(gc.GetTxHashesWithGasPenalizedSinceLastReset(key)))
	assert.Equal(t, []byte("hash1"), gc.GetTxHashesWithGasPenalizedSinceLastReset(key)[0])

	gc.SetGasPenalized(3, []byte("hash2"))
	assert.Equal(t, uint64(3), gc.GasPenalized([]byte("hash2")))
	require.Equal(t, 2, len(gc.GetTxHashesWithGasPenalizedSinceLastReset(key)))
	assert.Equal(t, []byte("hash1"), gc.GetTxHashesWithGasPenalizedSinceLastReset(key)[0])
	assert.Equal(t, []byte("hash2"), gc.GetTxHashesWithGasPenalizedSinceLastReset(key)[1])

	assert.Equal(t, uint64(5), gc.TotalGasPenalized())

	gc.RemoveGasPenalized([][]byte{[]byte("hash1")})
	assert.Equal(t, uint64(3), gc.TotalGasPenalized())

	gc.Init()
	assert.Equal(t, uint64(0), gc.TotalGasPenalized())
}

func TestComputeGasProvidedByTx_ShouldErrWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	_, _, err := gc.ComputeGasProvidedByTx(0, 1, nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsNotASmartContract(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	tx := transaction.Transaction{GasLimit: 7}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(6), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractInShard(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.SCInvoking, process.SCInvoking
			}},
		createEnableEpochsHandler(),
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 0, &tx)
	assert.Equal(t, uint64(7), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractCrossShard(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			}},
		createEnableEpochsHandler(),
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldReturnZeroIf0GasLimit(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			}},
		createEnableEpochsHandler(),
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 0, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &scr)
	assert.Equal(t, uint64(0), gasInSnd)
	assert.Equal(t, uint64(0), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldReturnGasLimitIfLessThanMoveBalance(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.MoveBalance, process.SCInvoking
			}},
		createEnableEpochsHandler(),
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 3, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &scr)
	assert.Equal(t, uint64(3), gasInSnd)
	assert.Equal(t, uint64(3), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldReturnGasLimitWhenRelayed(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.RelayedTx, process.RelayedTx
			}},
		createEnableEpochsHandler(),
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 3, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &scr)
	assert.Equal(t, uint64(3), gasInSnd)
	assert.Equal(t, uint64(3), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldReturnGasLimitWhenRelayedV2(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 0
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.RelayedTxV2, process.RelayedTxV2
			}},
		createEnableEpochsHandler(),
	)

	scr := smartContractResult.SmartContractResult{GasLimit: 3, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &scr)
	assert.Equal(t, uint64(3), gasInSnd)
	assert.Equal(t, uint64(3), gasInRcv)
}

func TestComputeGasProvidedByMiniBlock_ShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
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
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
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
		&economicsmocks.EconomicsHandlerStub{
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
		createEnableEpochsHandler(),
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
		&economicsmocks.EconomicsHandlerStub{
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
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
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
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	tx := transaction.Transaction{GasLimit: 7}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(6), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractInShardV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.SCInvoking, process.SCInvoking
			}},
		createEnableEpochsHandler(),
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 0, &tx)
	assert.Equal(t, uint64(7), gasInSnd)
	assert.Equal(t, uint64(7), gasInRcv)
}

func TestComputeGasProvidedByTx_ShouldWorkWhenTxReceiverAddressIsASmartContractCrossShardV1(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{
			ComputeGasLimitCalled: func(tx data.TransactionWithFeeHandler) uint64 {
				return 6
			},
		},
		&testscommon.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.SCInvoking, process.SCInvoking
			}},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	tx := transaction.Transaction{GasLimit: 7, RcvAddr: make([]byte, core.NumInitCharactersForScAddress+1)}

	gasInSnd, gasInRcv, _ := gc.ComputeGasProvidedByTx(0, 1, &tx)
	assert.Equal(t, uint64(6), gasInSnd)
	assert.Equal(t, uint64(1), gasInRcv)
}

func TestReset_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	key := []byte("key")
	gc.Reset(key)

	gc.SetGasProvided(5, []byte("hash1"))
	gc.SetGasProvidedAsScheduled(7, []byte("hash2"))
	gc.SetGasRefunded(2, []byte("hash1"))
	gc.SetGasPenalized(1, []byte("hash2"))

	require.Equal(t, 1, len(gc.GetTxHashesWithGasProvidedSinceLastReset(key)))
	assert.Equal(t, []byte("hash1"), gc.GetTxHashesWithGasProvidedSinceLastReset(key)[0])

	require.Equal(t, 1, len(gc.GetTxHashesWithGasProvidedAsScheduledSinceLastReset(key)))
	assert.Equal(t, []byte("hash2"), gc.GetTxHashesWithGasProvidedAsScheduledSinceLastReset(key)[0])

	require.Equal(t, 1, len(gc.GetTxHashesWithGasRefundedSinceLastReset(key)))
	assert.Equal(t, []byte("hash1"), gc.GetTxHashesWithGasRefundedSinceLastReset(key)[0])

	require.Equal(t, 1, len(gc.GetTxHashesWithGasPenalizedSinceLastReset(key)))
	assert.Equal(t, []byte("hash2"), gc.GetTxHashesWithGasPenalizedSinceLastReset(key)[0])

	gc.Reset(key)

	require.Equal(t, 0, len(gc.GetTxHashesWithGasProvidedSinceLastReset(key)))
	require.Equal(t, 0, len(gc.GetTxHashesWithGasProvidedAsScheduledSinceLastReset(key)))
	require.Equal(t, 0, len(gc.GetTxHashesWithGasRefundedSinceLastReset(key)))
	require.Equal(t, 0, len(gc.GetTxHashesWithGasPenalizedSinceLastReset(key)))
}

func TestRestoreGasSinceLastReset_ShouldWork(t *testing.T) {
	t.Parallel()

	gc, _ := preprocess.NewGasComputation(
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.TxTypeHandlerMock{},
		createEnableEpochsHandler(),
	)

	gc.SetGasProvided(5, []byte("hash1"))
	gc.SetGasProvidedAsScheduled(7, []byte("hash2"))
	gc.SetGasRefunded(2, []byte("hash1"))
	gc.SetGasPenalized(1, []byte("hash2"))

	assert.Equal(t, uint64(5), gc.TotalGasProvided())
	assert.Equal(t, uint64(7), gc.TotalGasProvidedAsScheduled())
	assert.Equal(t, uint64(2), gc.TotalGasRefunded())
	assert.Equal(t, uint64(1), gc.TotalGasPenalized())

	gc.Reset([]byte("key"))

	gc.SetGasProvided(5, []byte("hash3"))
	gc.SetGasProvidedAsScheduled(7, []byte("hash4"))
	gc.SetGasRefunded(2, []byte("hash3"))
	gc.SetGasPenalized(1, []byte("hash4"))

	assert.Equal(t, uint64(10), gc.TotalGasProvided())
	assert.Equal(t, uint64(14), gc.TotalGasProvidedAsScheduled())
	assert.Equal(t, uint64(4), gc.TotalGasRefunded())
	assert.Equal(t, uint64(2), gc.TotalGasPenalized())

	gc.RestoreGasSinceLastReset([]byte("key"))

	assert.Equal(t, uint64(5), gc.TotalGasProvided())
	assert.Equal(t, uint64(7), gc.TotalGasProvidedAsScheduled())
	assert.Equal(t, uint64(2), gc.TotalGasRefunded())
	assert.Equal(t, uint64(1), gc.TotalGasPenalized())
}
