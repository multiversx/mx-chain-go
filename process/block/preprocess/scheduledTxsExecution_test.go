package preprocess

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilTxProcessor(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		nil,
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilTxCoordinator(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		nil,
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilTransactionCoordinator, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilStorer(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		nil,
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilMarshaller(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		nil,
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilHasher(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		nil,
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilShardCoordinator(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		nil,
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionOk(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, scheduledTxsExec)
}

func TestScheduledTxsExecution_InitShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	scheduledTxsExec.AddScheduledTx([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	scheduledTxsExec.AddScheduledTx([]byte("txHash3"), &transaction.Transaction{Nonce: 2})

	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))

	scheduledTxsExec.Init()

	assert.Equal(t, 0, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 0, len(scheduledTxsExec.scheduledTxs))
}

func TestScheduledTxsExecution_AddShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	res := scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	assert.True(t, res)
	assert.Equal(t, 1, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 1, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	assert.False(t, res)
	assert.Equal(t, 1, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 1, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.AddScheduledTx([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	assert.True(t, res)
	assert.Equal(t, 2, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 2, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.AddScheduledTx([]byte("txHash3"), &transaction.Transaction{Nonce: 1})
	assert.True(t, res)
	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.AddScheduledTx([]byte("txHash2"), &transaction.Transaction{Nonce: 2})
	assert.False(t, res)
	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))
}

func TestScheduledTxsExecution_ExecuteShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.True(t, errors.Is(err, process.ErrMissingTransaction))
}

func TestScheduledTxsExecution_ExecuteShouldErr(t *testing.T) {
	t.Parallel()

	localError := errors.New("error")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, localError
			},
		},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Equal(t, localError, err)
}

func TestScheduledTxsExecution_ExecuteShouldWorkOnErrFailedTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, process.ErrFailedTransaction
			},
		},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, nil
			},
		},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrNilHaveTimeHandler(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.ExecuteAll(nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrTimeIsOut(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(-1) }
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrFailedTransaction(t *testing.T) {
	t.Parallel()

	localError := errors.New("error")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, localError
			},
		},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction)
	assert.Equal(t, localError, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldWorkOnErrFailedTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, process.ErrFailedTransaction
			},
		},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction)
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldWork(t *testing.T) {
	t.Parallel()

	numTxsExecuted := 0
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				numTxsExecuted++
				return vmcommon.Ok, nil
			},
		},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	scheduledTxsExec.AddScheduledTx([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	scheduledTxsExec.AddScheduledTx([]byte("txHash3"), &transaction.Transaction{Nonce: 2})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction)
	assert.Nil(t, err)
	assert.Equal(t, 3, numTxsExecuted)
}

func TestScheduledTxsExecution_executeShouldErr(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.execute(nil)
	assert.True(t, errors.Is(err, process.ErrWrongTypeAssertion))
}

func TestScheduledTxsExecution_executeShouldWork(t *testing.T) {
	t.Parallel()

	response := errors.New("response")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, response
			},
		},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.execute(&transaction.Transaction{Nonce: 0})
	assert.Equal(t, response, err)
}

func TestScheduledTxsExecution_computeScheduledSCRsShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorStub{
		SameShardCalled: func(_, _ []byte) bool {
			return false
		},
	}

	mapAllIntermediateTxsBeforeScheduledExecution := map[block.Type]map[string]data.TransactionHandler{
		0: {
			"txHash1": &transaction.Transaction{Nonce: 1},
			"txHash2": &transaction.Transaction{Nonce: 2},
		},
	}
	mapAllIntermediateTxsAfterScheduledExecution := map[block.Type]map[string]data.TransactionHandler{
		1: {
			"txHash3": &transaction.Transaction{Nonce: 3},
			"txHash4": &transaction.Transaction{Nonce: 4},
		},
	}

	t.Run("nil maps, empty scheduled scrs", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			shardCoordinator,
		)

		scheduledTxsExec.ComputeScheduledIntermediateTxs(nil, nil)

		assert.Equal(t, 0, len(scheduledTxsExec.GetMapScheduledIntermediateTxs()))
	})
	t.Run("nil map after txs execution, empty scheduled scrs", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			shardCoordinator,
		)

		scheduledTxsExec.ComputeScheduledIntermediateTxs(mapAllIntermediateTxsBeforeScheduledExecution, nil)

		assert.Equal(t, 0, len(scheduledTxsExec.GetMapScheduledIntermediateTxs()))
	})
	t.Run("nil map after txs execution, empty scheduled scrs", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			shardCoordinator,
		)

		localMapAllIntermediateTxsAfterScheduledExecution := map[block.Type]map[string]data.TransactionHandler{
			0: {
				"txHash1": &transaction.Transaction{Nonce: 1},
				"txHash2": &transaction.Transaction{Nonce: 2},
			},
		}
		scheduledTxsExec.ComputeScheduledIntermediateTxs(
			mapAllIntermediateTxsBeforeScheduledExecution,
			localMapAllIntermediateTxsAfterScheduledExecution,
		)

		assert.Equal(t, 0, len(scheduledTxsExec.mapScheduledIntermediateTxs))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			shardCoordinator,
		)

		scheduledTxsExec.ComputeScheduledIntermediateTxs(
			mapAllIntermediateTxsBeforeScheduledExecution,
			mapAllIntermediateTxsAfterScheduledExecution,
		)

		mapScheduledSCRs := scheduledTxsExec.GetMapScheduledIntermediateTxs()
		assert.Equal(t, 1, len(mapScheduledSCRs))
		assert.Equal(t, 2, len(mapScheduledSCRs[1]))
	})
}

func TestScheduledTxsExecution_computeScheduledSCRsShouldRemoveInvalidSCRs(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{
			SameShardCalled: func(_, _ []byte) bool {
				return false
			},
		},
	)

	txHash1 := "txHash1"
	txHash2 := "txHash2"
	txHash3 := "txHash3"

	tx1 := &transaction.Transaction{Nonce: 1}
	tx2 := &transaction.Transaction{Nonce: 2}
	tx3 := &transaction.Transaction{Nonce: 3}

	mapAllIntermediateTxsBeforeScheduledExecutionWithInvalid := map[block.Type]map[string]data.TransactionHandler{
		block.SmartContractResultBlock: {
			"txHashNoScheduled0": &transaction.Transaction{Nonce: 11},
			"txHashNoScheduled1": &transaction.Transaction{Nonce: 12},
		},
	}

	mapAllIntermediateTxsAfterScheduledExecutionWithInvalid := map[block.Type]map[string]data.TransactionHandler{
		block.SmartContractResultBlock: {
			txHash1: tx1,
			txHash3: tx3,
		},
		block.InvalidBlock: {
			txHash2: tx2,
		},
	}

	mb1TxHashes := [][]byte{
		[]byte(txHash1),
		[]byte(txHash2),
		[]byte(txHash3),
	}

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{TxHashes: mb1TxHashes},
	}

	scheduledTxsExec.AddScheduledMiniBlocks(mbs)
	scheduledTxsExec.ComputeScheduledIntermediateTxs(
		mapAllIntermediateTxsBeforeScheduledExecutionWithInvalid,
		mapAllIntermediateTxsAfterScheduledExecutionWithInvalid,
	)

	mapScheduledSCRs := scheduledTxsExec.mapScheduledIntermediateTxs

	assert.Equal(t, 2, len(mapScheduledSCRs))

	scheduledSCRs := mapScheduledSCRs[block.SmartContractResultBlock]
	assert.Equal(t, 2, len(scheduledSCRs))
	assert.True(t, reflect.DeepEqual(tx1, scheduledSCRs[0]))
	assert.True(t, reflect.DeepEqual(tx3, scheduledSCRs[1]))

	invalidSCRs := mapScheduledSCRs[block.InvalidBlock]
	assert.Equal(t, 1, len(invalidSCRs))
	assert.True(t, reflect.DeepEqual(tx2, invalidSCRs[0]))
}

func TestScheduledTxsExecution_getAllIntermediateTxsAfterScheduledExecution(t *testing.T) {
	t.Parallel()

	allTxsBeforeExec := map[block.Type]map[string]data.TransactionHandler{
		0: {
			"txHash1": &transaction.Transaction{Nonce: 1},
			"txHash2": &transaction.Transaction{Nonce: 2},
		},
	}
	allTxsAfterExec := map[string]data.TransactionHandler{
		"txHash3": &transaction.Transaction{Nonce: 3},
		"txHash4": &transaction.Transaction{Nonce: 4},
	}

	t.Run("not already existing txs, different shard", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return false
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			0,
		)

		assert.Equal(t, 2, len(scrsInfo))
	})
	t.Run("not already existing txs, same shard, scr", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return true
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			block.SmartContractResultBlock,
		)

		assert.Equal(t, 0, len(scrsInfo))
	})
	t.Run("not already existing txs, same shard, receipt", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return true
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			block.ReceiptBlock,
		)

		assert.Equal(t, 0, len(scrsInfo))
	})
	t.Run("not already existing txs, same shard, transaction", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return true
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			block.TxBlock,
		)

		assert.Equal(t, 2, len(scrsInfo))
	})
	t.Run("not already existing txs, same shard, invalid", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return true
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			block.InvalidBlock,
		)

		assert.Equal(t, 2, len(scrsInfo))
	})
	t.Run("not existing block type, different shard", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return false
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			0,
		)

		assert.Equal(t, 2, len(scrsInfo))
	})
	t.Run("already existing txs, different shard", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return false
				},
			},
		)

		allTxsAfterExec := map[string]data.TransactionHandler{
			"txHash1": &transaction.Transaction{Nonce: 1},
			"txHash2": &transaction.Transaction{Nonce: 2},
		}

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			0,
		)

		assert.Equal(t, 0, len(scrsInfo))
	})
}

func TestScheduledTxsExecution_GetScheduledIntermediateTxsNonEmptySCRsMap(t *testing.T) {
	t.Parallel()

	allTxsAfterExec := map[block.Type]map[string]data.TransactionHandler{
		0: {
			"txHash1": &transaction.Transaction{Nonce: 1},
			"txHash2": &transaction.Transaction{Nonce: 2},
		},
		1: {
			"txHash3": &transaction.Transaction{Nonce: 3},
			"txHash4": &transaction.Transaction{Nonce: 4},
		},
	}

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{
			SameShardCalled: func(_, _ []byte) bool {
				return false
			},
		},
	)

	scheduledTxsExec.ComputeScheduledIntermediateTxs(
		nil,
		allTxsAfterExec,
	)

	scheduledIntermediateTxs := scheduledTxsExec.GetScheduledIntermediateTxs()

	assert.Equal(t, 2, len(scheduledIntermediateTxs))
	assert.Equal(t, 2, len(scheduledIntermediateTxs[0]))
	assert.Equal(t, 2, len(scheduledIntermediateTxs[1]))
}

func TestScheduledTxsExecution_GetScheduledIntermediateTxsEmptySCRsMap(t *testing.T) {
	t.Parallel()

	allTxsAfterExec := make(map[block.Type]map[string]data.TransactionHandler)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{
			SameShardCalled: func(_, _ []byte) bool {
				return false
			},
		},
	)

	scheduledTxsExec.ComputeScheduledIntermediateTxs(
		nil,
		allTxsAfterExec,
	)

	scheduledIntermediateTxs := scheduledTxsExec.GetScheduledIntermediateTxs()

	assert.Equal(t, 0, len(scheduledIntermediateTxs))
}

func TestScheduledTxsExecution_SetScheduledInfo(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	rootHash := []byte("root hash")
	gasAndFees := scheduled.GasAndFees{}
	intermediateTxs := map[block.Type][]data.TransactionHandler{
		0: {
			&transaction.Transaction{Nonce: 1},
			&transaction.Transaction{Nonce: 2},
		},
		1: {
			&transaction.Transaction{Nonce: 3},
			&transaction.Transaction{Nonce: 4},
		},
	}
	mbs := block.MiniBlockSlice{
		0: {
			Type: block.InvalidBlock,
		},
	}

	scheduledInfo := &process.ScheduledInfo{
		RootHash:        rootHash,
		IntermediateTxs: intermediateTxs,
		GasAndFees:      gasAndFees,
		MiniBlocks:      mbs,
	}
	scheduledTxsExec.SetScheduledInfo(scheduledInfo)

	assert.Equal(t, rootHash, scheduledTxsExec.GetScheduledRootHash())
	assert.Equal(t, gasAndFees, scheduledTxsExec.GetScheduledGasAndFees())
	assert.Equal(t, intermediateTxs, scheduledTxsExec.GetScheduledIntermediateTxs())
	assert.Equal(t, mbs, scheduledTxsExec.GetScheduledMiniBlocks())
}

func TestScheduledTxsExecution_Setters(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	gasAndFees := scheduled.GasAndFees{}

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)
	scheduledTxsExec.SetTransactionCoordinator(&testscommon.TransactionCoordinatorMock{})
	scheduledTxsExec.SetTransactionProcessor(&testscommon.TxProcessorMock{})

	scheduledTxsExec.SetScheduledGasAndFees(gasAndFees)
	assert.Equal(t, gasAndFees, scheduledTxsExec.GetScheduledGasAndFees())

	scheduledTxsExec.SetScheduledRootHash(rootHash)
	assert.Equal(t, rootHash, scheduledTxsExec.GetScheduledRootHash())
}

func TestScheduledTxsExecution_getScheduledInfoForHeaderShouldFail(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")

	t.Run("failed to get SCRs saved data from storage", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("storer err")
		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			&storageMocks.StorerStub{
				GetCalled: func(_ []byte) ([]byte, error) {
					return nil, expectedErr
				},
			},
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{},
		)

		scheduledInfo, err := scheduledTxsExec.getScheduledInfoForHeader(rootHash, core.OptionalUint32{})
		assert.Nil(t, scheduledInfo)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("failed to unmarshal data", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("marshaller err")
		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			&storageMocks.StorerStub{
				GetCalled: func(_ []byte) ([]byte, error) {
					return nil, nil
				},
			},
			&marshallerMock.MarshalizerStub{
				UnmarshalCalled: func(_ interface{}, _ []byte) error {
					return expectedErr
				},
			},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{},
		)

		scheduledInfo, err := scheduledTxsExec.getScheduledInfoForHeader(rootHash, core.OptionalUint32{})
		assert.Nil(t, scheduledInfo)
		assert.Equal(t, expectedErr, err)
	})
}

func TestScheduledTxsExecution_getScheduledInfoForHeaderShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("root hash")
	expectedGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(102),
		GasProvided:     103,
		GasPenalized:    104,
		GasRefunded:     105,
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash:   headerHash,
		Scrs:       []*smartContractResult.SmartContractResult{},
		GasAndFees: &expectedGasAndFees,
	}
	marshalledSCRsSavedData, _ := json.Marshal(scheduledSCRs)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return marshalledSCRsSavedData, nil
			},
		},
		&marshallerMock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledInfo, _ := scheduledTxsExec.getScheduledInfoForHeader(headerHash, core.OptionalUint32{})

	assert.Equal(t, headerHash, scheduledInfo.RootHash)
	assert.Equal(t, expectedGasAndFees, scheduledInfo.GasAndFees)
	assert.Nil(t, scheduledInfo.IntermediateTxs)
	assert.Equal(t, block.MiniBlockSlice(nil), scheduledInfo.MiniBlocks)
}

func TestScheduledTxsExecution_getMarshalledScheduledInfoShouldWork(t *testing.T) {
	t.Parallel()

	scheduledRootHash := []byte("root hash")
	mapSCRs := map[block.Type][]data.TransactionHandler{
		block.SmartContractResultBlock: {
			&smartContractResult.SmartContractResult{
				Nonce: 1,
			},
		},
	}
	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(100),
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash: scheduledRootHash,
		Scrs: []*smartContractResult.SmartContractResult{
			{
				Nonce: 1,
			},
		},
		GasAndFees: &gasAndFees,
	}
	expectedScheduledSCRs, _ := json.Marshal(scheduledSCRs)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshallerMock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledInfo := &process.ScheduledInfo{
		RootHash:        scheduledRootHash,
		IntermediateTxs: mapSCRs,
		GasAndFees:      gasAndFees,
	}
	marshalledSCRs, err := scheduledTxsExec.getMarshalledScheduledInfo(scheduledInfo)
	assert.Nil(t, err)
	assert.Equal(t, expectedScheduledSCRs, marshalledSCRs)
}

func TestScheduledTxsExecution_RollBackToBlockShouldFail(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")

	expectedErr := errors.New("local err")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return nil, expectedErr
			},
		},
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.RollBackToBlock(rootHash)
	assert.Equal(t, expectedErr, err)
}

func TestScheduledTxsExecution_RollBackToBlockShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("root hash")
	expectedGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(102),
		GasProvided:     103,
		GasPenalized:    104,
		GasRefunded:     105,
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash: headerHash,
		Scrs: []*smartContractResult.SmartContractResult{
			{
				Nonce: 0,
			},
		},
		GasAndFees: &expectedGasAndFees,
	}
	marshalledSCRsSavedData, _ := json.Marshal(scheduledSCRs)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return marshalledSCRsSavedData, nil
			},
		},
		&marshallerMock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.RollBackToBlock(headerHash)
	assert.Nil(t, err)

	scheduledInfo, err := scheduledTxsExec.getScheduledInfoForHeader(headerHash, core.OptionalUint32{})
	require.Nil(t, err)
	assert.Equal(t, headerHash, scheduledInfo.RootHash)
	assert.Equal(t, expectedGasAndFees, scheduledInfo.GasAndFees)
	assert.NotNil(t, scheduledInfo.IntermediateTxs)
	assert.Equal(t, block.MiniBlockSlice(nil), scheduledInfo.MiniBlocks)
}

func TestScheduledTxsExecution_SaveState(t *testing.T) {
	t.Parallel()

	headerHash := []byte("header hash")
	scheduledRootHash := []byte("scheduled root hash")
	mapSCRs := map[block.Type][]data.TransactionHandler{
		block.SmartContractResultBlock: {
			&smartContractResult.SmartContractResult{
				Nonce: 1,
			},
		},
		block.InvalidBlock: {
			&transaction.Transaction{
				Nonce: 2,
			},
		},
	}
	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(100),
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash: scheduledRootHash,
		Scrs: []*smartContractResult.SmartContractResult{
			{
				Nonce: 1,
			},
		},
		InvalidTransactions: []*transaction.Transaction{
			{
				Nonce: 2,
			},
		},
		GasAndFees: &gasAndFees,
	}
	marshalledScheduledData, err := json.Marshal(scheduledSCRs)
	require.Nil(t, err)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			PutCalled: func(key, data []byte) error {
				require.Equal(t, headerHash, key)
				require.Equal(t, marshalledScheduledData, data)
				return nil
			},
		},
		&marshallerMock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledInfo := &process.ScheduledInfo{
		RootHash:        scheduledRootHash,
		IntermediateTxs: mapSCRs,
		GasAndFees:      gasAndFees,
	}
	scheduledTxsExec.SaveState(headerHash, scheduledInfo)
}

func TestScheduledTxsExecution_SaveStateIfNeeded(t *testing.T) {
	t.Parallel()

	headerHash := []byte("header hash")

	wasCalled := false
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			PutCalled: func(key, _ []byte) error {
				wasCalled = true
				require.Equal(t, headerHash, key)
				return nil
			},
		},
		&marshallerMock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.SaveStateIfNeeded(headerHash)
	assert.False(t, wasCalled)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	scheduledTxsExec.SaveStateIfNeeded(headerHash)
	assert.True(t, wasCalled)
}

func TestScheduledTxsExecution_IsScheduledTx(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)
	scheduledTxsExec.AddScheduledTx(txHash1, &transaction.Transaction{Nonce: 0})

	ok := scheduledTxsExec.IsScheduledTx(txHash1)
	assert.True(t, ok)

	ok = scheduledTxsExec.IsScheduledTx(txHash2)
	assert.False(t, ok)
}

func TestScheduledTxsExecution_AddMiniBlocksWithNilReservedNilTxHashes(t *testing.T) {
	t.Run("nil Reserved", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{},
		)

		miniBlocks := block.MiniBlockSlice{}
		mb := &block.MiniBlock{
			TxHashes:        make([][]byte, 10),
			ReceiverShardID: 1,
			SenderShardID:   2,
			Type:            3,
			Reserved:        nil,
		}
		miniBlocks = append(miniBlocks, mb)

		scheduledTxsExec.AddScheduledMiniBlocks(miniBlocks)

		assert.Nil(t, scheduledTxsExec.scheduledMbs[0].Reserved)
		assert.NotNil(t, scheduledTxsExec.scheduledMbs[0].TxHashes)
	})
	t.Run("nil TxHashes", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshal.GogoProtoMarshalizer{},
			&hashingMocks.HasherMock{},
			&mock.ShardCoordinatorStub{},
		)

		miniBlocks := block.MiniBlockSlice{}
		mb := &block.MiniBlock{
			TxHashes:        nil,
			ReceiverShardID: 1,
			SenderShardID:   2,
			Type:            3,
			Reserved:        make([]byte, 10),
		}
		miniBlocks = append(miniBlocks, mb)

		scheduledTxsExec.AddScheduledMiniBlocks(miniBlocks)

		assert.Nil(t, scheduledTxsExec.scheduledMbs[0].TxHashes)
		assert.NotNil(t, scheduledTxsExec.scheduledMbs[0].Reserved)
	})
}

func TestScheduledTxsExecution_AddMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	miniBlocks := block.MiniBlockSlice{}
	mb1 := &block.MiniBlock{
		TxHashes:        make([][]byte, 10),
		ReceiverShardID: 1,
		SenderShardID:   2,
		Type:            3,
		Reserved:        make([]byte, 2),
	}
	mb2 := &block.MiniBlock{
		TxHashes:        make([][]byte, 5),
		ReceiverShardID: 3,
		SenderShardID:   1,
		Type:            2,
		Reserved:        make([]byte, 7),
	}

	miniBlocks = append(miniBlocks, mb1)
	miniBlocks = append(miniBlocks, mb2)

	scheduledTxsExec.AddScheduledMiniBlocks(miniBlocks)

	expectedLen := 2
	assert.Equal(t, expectedLen, len(scheduledTxsExec.scheduledMbs))
	assert.True(t, reflect.DeepEqual(miniBlocks[0], scheduledTxsExec.scheduledMbs[0]))
	assert.True(t, reflect.DeepEqual(miniBlocks[1], scheduledTxsExec.scheduledMbs[1]))
}

func TestScheduledTxsExecution_GetScheduledTxs(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)
	firstTransaction := &transaction.Transaction{Nonce: 0}
	secondTransaction := &transaction.Transaction{Nonce: 1}
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), firstTransaction)
	scheduledTxsExec.AddScheduledTx([]byte("txHash2"), secondTransaction)

	scheduledTxs := scheduledTxsExec.GetScheduledTxs()

	expectedLen := 2
	assert.Equal(t, expectedLen, len(scheduledTxs))
	assert.True(t, reflect.DeepEqual(firstTransaction, scheduledTxs[0]))
	assert.True(t, reflect.DeepEqual(secondTransaction, scheduledTxs[1]))
}

func TestScheduledTxsExecution_GetScheduledMBs(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	miniBlocks := block.MiniBlockSlice{}
	mb1 := &block.MiniBlock{
		TxHashes:        nil,
		ReceiverShardID: 1,
		SenderShardID:   2,
		Type:            3,
		Reserved:        make([]byte, 2),
	}
	mb2 := &block.MiniBlock{
		TxHashes:        make([][]byte, 5),
		ReceiverShardID: 3,
		SenderShardID:   1,
		Type:            2,
		Reserved:        nil,
	}
	mb3 := &block.MiniBlock{
		TxHashes:        make([][]byte, 10),
		ReceiverShardID: 2,
		SenderShardID:   2,
		Type:            2,
		Reserved:        make([]byte, 10),
	}

	miniBlocks = append(miniBlocks, mb1)
	miniBlocks = append(miniBlocks, mb2)
	miniBlocks = append(miniBlocks, mb3)

	scheduledTxsExec.AddScheduledMiniBlocks(miniBlocks)

	scheduledMBs := scheduledTxsExec.GetScheduledMiniBlocks()

	expectedLen := 3
	assert.Equal(t, expectedLen, len(scheduledMBs))

	assert.True(t, reflect.DeepEqual(mb1, scheduledMBs[0]))
	assert.True(t, reflect.DeepEqual(mb2, scheduledMBs[1]))
	assert.True(t, reflect.DeepEqual(mb3, scheduledMBs[2]))
}

func TestScheduledTxsExecution_GetScheduledRootHashForHeaderWithErrorShouldFail(t *testing.T) {
	t.Parallel()

	headerHash := []byte("root hash")
	storerGetErr := errors.New("storer.Get() error")
	storerGetFromEpochErr := errors.New("storer.GetFromEpoch() err")

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return nil, storerGetErr
			},
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				return nil, storerGetFromEpochErr
			},
		},
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	t.Run("without epoch", func(t *testing.T) {
		scheduledInfo, err := scheduledTxsExec.GetScheduledRootHashForHeader(headerHash)
		assert.Nil(t, scheduledInfo)
		assert.Equal(t, storerGetErr, err)
	})

	t.Run("with epoch", func(t *testing.T) {
		scheduledInfo, err := scheduledTxsExec.GetScheduledRootHashForHeaderWithEpoch(headerHash, 7)
		assert.Nil(t, scheduledInfo)
		assert.Equal(t, storerGetFromEpochErr, err)
	})
}

func TestScheduledTxsExecution_GetScheduledRootHashForHeaderShouldWork(t *testing.T) {
	t.Parallel()

	scheduledHash := []byte("scheduled hash")
	expectedGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(102),
		GasProvided:     103,
		GasPenalized:    104,
		GasRefunded:     105,
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash:   scheduledHash,
		GasAndFees: &expectedGasAndFees,
	}
	marshalledSCRsSavedData, _ := json.Marshal(scheduledSCRs)

	var keyPassedToStorerGet []byte
	var epochPassedToStorerGet uint32

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				keyPassedToStorerGet = key
				return marshalledSCRsSavedData, nil
			},
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				keyPassedToStorerGet = key
				epochPassedToStorerGet = epoch
				return marshalledSCRsSavedData, nil
			},
		},
		&marshallerMock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	t.Run("without epoch", func(t *testing.T) {
		scheduledRootHash, err := scheduledTxsExec.GetScheduledRootHashForHeader([]byte("aabb"))
		assert.Nil(t, err)
		assert.Equal(t, scheduledHash, scheduledRootHash)
		assert.Equal(t, keyPassedToStorerGet, []byte("aabb"))
	})

	t.Run("with epoch", func(t *testing.T) {
		scheduledRootHash, err := scheduledTxsExec.GetScheduledRootHashForHeaderWithEpoch([]byte("ccdd"), 7)
		assert.Nil(t, err)
		assert.Equal(t, scheduledHash, scheduledRootHash)
		assert.Equal(t, []byte("ccdd"), keyPassedToStorerGet)
		assert.Equal(t, uint32(7), epochPassedToStorerGet)
	})
}

func TestScheduledTxsExecution_removeInvalidTxsFromScheduledMiniBlocks(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&hashingMocks.HasherMock{},
		&mock.ShardCoordinatorStub{},
	)

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	txHash3 := []byte("txHash3")
	txHash4 := []byte("txHash4")
	txHash5 := []byte("txHash5")
	txHash6 := []byte("txHash6")

	mb1TxHashes := [][]byte{
		txHash1,
		txHash2,
		txHash3,
		txHash4,
		txHash5,
	}
	mb2TxHashes := [][]byte{
		txHash6,
	}

	mbs := block.MiniBlockSlice{
		&block.MiniBlock{TxHashes: mb1TxHashes},
		&block.MiniBlock{TxHashes: mb2TxHashes},
	}

	scrsInfo := []*intermediateTxInfo{
		{txHash: txHash1},
		{txHash: txHash3},
		{txHash: txHash5},
		{txHash: txHash6},
	}

	scheduledTxsExec.scheduledMbs = mbs
	scheduledTxsExec.removeInvalidTxsFromScheduledMiniBlocks(scrsInfo)

	// a scheduled miniBlock without any txs is removed completely
	expectedNumScheduledMiniBlocks := 1
	assert.Equal(t, expectedNumScheduledMiniBlocks, len(scheduledTxsExec.scheduledMbs))

	expectedLen := 2
	assert.Equal(t, expectedLen, len(scheduledTxsExec.scheduledMbs[0].TxHashes))

	remainingTxHashes := scheduledTxsExec.scheduledMbs[0].TxHashes
	assert.True(t, bytes.Equal(txHash2, remainingTxHashes[0]))
	assert.True(t, bytes.Equal(txHash4, remainingTxHashes[1]))
}

func TestScheduledTxsExecution_getIndexOfTxHashInMiniBlock(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")

	txHash6 := []byte("txHash6")

	mb1TxHashes := [][]byte{
		txHash1,
		txHash2,
	}

	mb0 := &block.MiniBlock{TxHashes: mb1TxHashes}

	assert.Equal(t, 0, getIndexOfTxHashInMiniBlock(txHash1, mb0))
	assert.Equal(t, 1, getIndexOfTxHashInMiniBlock(txHash2, mb0))
	assert.Equal(t, -1, getIndexOfTxHashInMiniBlock(txHash6, mb0))
}

func TestScheduledTxsExecution_setScheduledMiniBlockHashes(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	t.Run("fail to calculate hash", func(t *testing.T) {
		expectedErr := errors.New("calculate hash err")
		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshallerMock.MarshalizerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					return nil, expectedErr
				},
			},
			&mock.HasherStub{},
			&mock.ShardCoordinatorStub{},
		)

		miniBlocks := block.MiniBlockSlice{&block.MiniBlock{
			TxHashes: [][]byte{[]byte("dummyhash")},
		}}
		scheduledTxsExec.AddScheduledMiniBlocks(miniBlocks)

		err := scheduledTxsExec.setScheduledMiniBlockHashes()
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		mb := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("dummyhash")},
		}

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&testscommon.TransactionCoordinatorMock{},
			genericMocks.NewStorerMock(),
			&marshallerMock.MarshalizerStub{
				MarshalCalled: func(obj interface{}) ([]byte, error) {
					assert.Equal(t, mb, obj)
					return nil, nil
				},
			},
			&mock.HasherStub{
				ComputeCalled: func(s string) []byte {
					return hash
				},
			},
			&mock.ShardCoordinatorStub{},
		)

		miniBlocks := block.MiniBlockSlice{mb}
		scheduledTxsExec.AddScheduledMiniBlocks(miniBlocks)

		expectedMiniBlockHashes := make(map[string]struct{})
		expectedMiniBlockHashes[string(hash)] = struct{}{}

		err := scheduledTxsExec.setScheduledMiniBlockHashes()
		assert.Nil(t, err)
		assert.Equal(t, expectedMiniBlockHashes, scheduledTxsExec.mapScheduledMbHashes)
	})
}

func TestScheduledTxsExecution_IsMiniBlockExecuted(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	hash2 := []byte("hash2")

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&testscommon.TransactionCoordinatorMock{},
		genericMocks.NewStorerMock(),
		&marshal.GogoProtoMarshalizer{},
		&mock.HasherStub{
			ComputeCalled: func(s string) []byte {
				return hash1
			},
		},
		&mock.ShardCoordinatorStub{},
	)

	miniBlocks := block.MiniBlockSlice{&block.MiniBlock{
		TxHashes: [][]byte{[]byte("dummyhash")},
	}}
	scheduledTxsExec.AddScheduledMiniBlocks(miniBlocks)

	err := scheduledTxsExec.setScheduledMiniBlockHashes()
	require.Nil(t, err)

	ok := scheduledTxsExec.IsMiniBlockExecuted(hash1)
	assert.True(t, ok)

	ok = scheduledTxsExec.IsScheduledTx(hash2)
	assert.False(t, ok)
}
