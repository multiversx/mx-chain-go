package transaction

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func TestTransactionCostEstimator_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(nil, &mock.FeeHandlerStub{}, &mock.ScQueryStub{})

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTransactionCostEstimator_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, nil, &mock.ScQueryStub{})

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTransactionCostEstimator_NilQueryServiceShouldErr(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, nil)

	require.Nil(t, tce)
	require.Equal(t, external.ErrNilSCQueryService, err)
}

func TestTransactionCostEstimator_Ok(t *testing.T) {
	t.Parallel()

	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, &mock.ScQueryStub{})

	require.Nil(t, err)
	require.False(t, check.IfNil(tce))
}

func TestComputeTransactionGasLimit_TxTypeHandlerErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error")
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, err error) {
			return 0, expectedErr
		},
	}, &mock.FeeHandlerStub{}, &mock.ScQueryStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Zero(t, cost)
	require.Equal(t, expectedErr, err)
}

func TestComputeTransactionGasLimit_MoveBalance(t *testing.T) {
	t.Parallel()

	consumedGasUnits := big.NewInt(1000)
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, err error) {
			return process.MoveBalance, nil
		},
	}, &mock.FeeHandlerStub{
		ComputeFeeCalled: func(tx process.TransactionWithFeeHandler) *big.Int {
			return consumedGasUnits
		},
	}, &mock.ScQueryStub{})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits.Uint64(), cost)
}

func TestComputeTransactionGasLimit_SmartContractCall(t *testing.T) {
	t.Parallel()

	consumedGasUnits := big.NewInt(1000)
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (transactionType process.TransactionType, err error) {
			return process.SCInvoking, nil
		},
	}, &mock.FeeHandlerStub{}, &mock.ScQueryStub{
		ComputeScCallCostHandler: func(tx *transaction.Transaction) (u uint64, err error) {
			return consumedGasUnits.Uint64(), nil
		},
	})

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits.Uint64(), cost)
}
