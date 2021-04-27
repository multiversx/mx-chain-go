package preprocess

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilTxProcessor(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		nil,
	)

	assert.Nil(t, scheduledTxsExec)
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionOk(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&mock.TxProcessorMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, scheduledTxsExec)
}

func TestScheduledTxsExecution_InitShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{},
	)

	scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	scheduledTxsExec.Add([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	scheduledTxsExec.Add([]byte("txHash3"), &transaction.Transaction{Nonce: 2})

	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))

	scheduledTxsExec.Init()

	assert.Equal(t, 0, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 0, len(scheduledTxsExec.scheduledTxs))
}

func TestScheduledTxsExecution_AddShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{},
	)

	res := scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	assert.True(t, res)
	assert.Equal(t, 1, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 1, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	assert.False(t, res)
	assert.Equal(t, 1, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 1, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.Add([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	assert.True(t, res)
	assert.Equal(t, 2, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 2, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.Add([]byte("txHash3"), &transaction.Transaction{Nonce: 1})
	assert.True(t, res)
	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.Add([]byte("txHash2"), &transaction.Transaction{Nonce: 2})
	assert.False(t, res)
	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))
}

func TestScheduledTxsExecution_ExecuteShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{},
	)

	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.True(t, errors.Is(err, process.ErrMissingTransaction))
}

func TestScheduledTxsExecution_ExecuteShouldErr(t *testing.T) {
	t.Parallel()

	localError := errors.New("error")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, localError
			},
		},
	)

	scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Equal(t, localError, err)
}

func TestScheduledTxsExecution_ExecuteShouldWorkOnErrFailedTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, process.ErrFailedTransaction
			},
		},
	)

	scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, nil
			},
		},
	)

	scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrNilHaveTimeHandler(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{},
	)

	err := scheduledTxsExec.ExecuteAll(nil, nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrNilTransactionCoordinator(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(-1) }
	err := scheduledTxsExec.ExecuteAll(haveTimeFunction, nil)
	assert.Equal(t, process.ErrNilTransactionCoordinator, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrTimeIsOut(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(-1) }
	scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction, &mock.TransactionCoordinatorMock{})
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrFailedTransaction(t *testing.T) {
	t.Parallel()

	localError := errors.New("error")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, localError
			},
		},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction, &mock.TransactionCoordinatorMock{})
	assert.Equal(t, localError, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldWorkOnErrFailedTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, process.ErrFailedTransaction
			},
		},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction, &mock.TransactionCoordinatorMock{})
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldWork(t *testing.T) {
	t.Parallel()

	numTxsExecuted := 0
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				numTxsExecuted++
				return vmcommon.Ok, nil
			},
		},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.Add([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	scheduledTxsExec.Add([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	scheduledTxsExec.Add([]byte("txHash3"), &transaction.Transaction{Nonce: 2})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction, &mock.TransactionCoordinatorMock{})
	assert.Nil(t, err)
	assert.Equal(t, 3, numTxsExecuted)
}

func TestScheduledTxsExecution_executeShouldErr(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{},
	)

	err := scheduledTxsExec.execute(nil)
	assert.True(t, errors.Is(err, process.ErrWrongTypeAssertion))
}

func TestScheduledTxsExecution_executeShouldWork(t *testing.T) {
	t.Parallel()

	response := errors.New("response")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&mock.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, response
			},
		},
	)

	err := scheduledTxsExec.execute(&transaction.Transaction{Nonce: 0})
	assert.Equal(t, response, err)
}

//TODO: Add unit tests for computeScheduledSCRs, getSCRsInfo, GetScheduledSCRs and SetScheduledSCRs methods
