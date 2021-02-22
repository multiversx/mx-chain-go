package transaction

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func createGasMap(value uint64) map[string]map[string]uint64 {
	gasMap := make(map[string]map[string]uint64)

	baseOpMap := make(map[string]uint64)

	baseOpMap["StorePerByte"] = value
	baseOpMap["CompilePerByte"] = value
	baseOpMap["AoTPreparePerByte"] = value
	baseOpMap["GetCode"] = value
	gasMap[core.BaseOperationCost] = baseOpMap

	return gasMap
}

func TestTransactionCostEstimator_NilTxTypeHandler(t *testing.T) {
	t.Parallel()

	gasSchedule := mock.NewGasScheduleNotifierMock(createGasMap(1))
	tce, err := NewTransactionCostEstimator(nil, &mock.FeeHandlerStub{}, &mock.ScQueryStub{}, gasSchedule)

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilTxTypeHandler, err)
}

func TestTransactionCostEstimator_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	gasSchedule := mock.NewGasScheduleNotifierMock(createGasMap(1))
	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, nil, &mock.ScQueryStub{}, gasSchedule)

	require.Nil(t, tce)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestTransactionCostEstimator_NilQueryServiceShouldErr(t *testing.T) {
	t.Parallel()

	gasSchedule := mock.NewGasScheduleNotifierMock(createGasMap(1))
	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, nil, gasSchedule)

	require.Nil(t, tce)
	require.Equal(t, external.ErrNilSCQueryService, err)
}

func TestTransactionCostEstimator_Ok(t *testing.T) {
	t.Parallel()

	gasSchedule := mock.NewGasScheduleNotifierMock(createGasMap(1))
	tce, err := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{}, &mock.FeeHandlerStub{}, &mock.ScQueryStub{}, gasSchedule)

	require.Nil(t, err)
	require.False(t, check.IfNil(tce))
}

func TestComputeTransactionGasLimit_MoveBalance(t *testing.T) {
	t.Parallel()

	gasSchedule := mock.NewGasScheduleNotifierMock(createGasMap(1))
	consumedGasUnits := uint64(1000)
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.MoveBalance
		},
	}, &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return consumedGasUnits
		},
	}, &mock.ScQueryStub{}, gasSchedule)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits, cost)
}

func TestComputeTransactionGasLimit_SmartContractDeploy(t *testing.T) {
	t.Parallel()

	gasSchedule := mock.NewGasScheduleNotifierMock(createGasMap(2))
	gasLimitBaseTx := uint64(500)
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCDeployment, process.SCDeployment
		},
	}, &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return gasLimitBaseTx
		},
	}, &mock.ScQueryStub{}, gasSchedule)

	tx := &transaction.Transaction{
		Data: []byte("data"),
	}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, gasLimitBaseTx+uint64(16), cost)
}

func TestComputeTransactionGasLimit_SmartContractCall(t *testing.T) {
	t.Parallel()

	gasSchedule := mock.NewGasScheduleNotifierMock(createGasMap(1))
	gasLimitBaseTx := uint64(500)
	consumedGasUnits := big.NewInt(1000)
	tce, _ := NewTransactionCostEstimator(&mock.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.SCInvoking, process.SCInvoking
		},
	}, &mock.FeeHandlerStub{
		ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
			return gasLimitBaseTx
		},
	}, &mock.ScQueryStub{
		ComputeScCallGasLimitHandler: func(tx *transaction.Transaction) (u uint64, err error) {
			return consumedGasUnits.Uint64(), nil
		},
	}, gasSchedule)

	tx := &transaction.Transaction{}
	cost, err := tce.ComputeTransactionGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGasUnits.Uint64()+gasLimitBaseTx, cost)
}

func TestTransactionCostEstimator_RelayedTxShouldErr(t *testing.T) {
	t.Parallel()

	gasSchedule := mock.NewGasScheduleNotifierMock(createGasMap(1))
	tce, _ := NewTransactionCostEstimator(
		&mock.TxTypeHandlerMock{
			ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
				return process.RelayedTx, process.RelayedTx
			},
		},
		&mock.FeeHandlerStub{},
		&mock.ScQueryStub{},
		gasSchedule,
	)

	tx := &transaction.Transaction{}
	_, err := tce.ComputeTransactionGasLimit(tx)
	require.Error(t, err)
}
