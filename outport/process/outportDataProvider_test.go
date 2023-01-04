package process

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/ElrondNetwork/elrond-go/outport/process/transactionsfee"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/require"
)

func createArgOutportDataProvider() ArgOutportDataProvider {
	txsFeeProc, _ := transactionsfee.NewTransactionsFeeProcessor(transactionsfee.ArgTransactionsFeeProcessor{
		Marshaller:         &testscommon.MarshalizerMock{},
		TransactionsStorer: &genericMocks.StorerMock{},
		ShardCoordinator:   &testscommon.ShardsCoordinatorMock{},
		TxFeeCalculator:    &mock.EconomicsHandlerMock{},
	})

	return ArgOutportDataProvider{
		AlteredAccountsProvider:  &testscommon.AlteredAccountsProviderStub{},
		TransactionsFeeProcessor: txsFeeProc,
		TxCoordinator:            &testscommon.TransactionCoordinatorMock{},
		NodesCoordinator:         &shardingMocks.NodesCoordinatorMock{},
		GasConsumedProvider:      &testscommon.GasHandlerStub{},
		EconomicsData:            &mock.EconomicsHandlerMock{},
		ShardCoordinator:         &testscommon.ShardsCoordinatorMock{},
		ExecutionOrderHandler:    &mock.ExecutionOrderHandlerStub{},
	}
}

func TestNewOutportDataProvider(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProvider()
	outportDataP, err := NewOutportDataProvider(arg)
	require.Nil(t, err)
	require.False(t, outportDataP.IsInterfaceNil())
}

func TestPrepareOutportSaveBlockDataNilHeader(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProvider()
	outportDataP, _ := NewOutportDataProvider(arg)

	_, err := outportDataP.PrepareOutportSaveBlockData(ArgPrepareOutportSaveBlockData{})
	require.Equal(t, errNilHeaderHandler, err)
}

func TestPrepareOutportSaveBlockDataNilBody(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProvider()
	outportDataP, _ := NewOutportDataProvider(arg)

	_, err := outportDataP.PrepareOutportSaveBlockData(ArgPrepareOutportSaveBlockData{
		Header: &block.Header{},
	})
	require.Equal(t, errNilBodyHandler, err)
}

func TestPrepareOutportSaveBlockData(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProvider()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetValidatorsPublicKeysCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]string, error) {
			return nil, nil
		},
		GetValidatorsIndexesCalled: func(publicKeys []string, epoch uint32) ([]uint64, error) {
			return []uint64{0, 1}, nil
		},
	}
	outportDataP, _ := NewOutportDataProvider(arg)

	res, err := outportDataP.PrepareOutportSaveBlockData(ArgPrepareOutportSaveBlockData{
		Header:     &block.Header{},
		Body:       &block.Body{},
		HeaderHash: []byte("something"),
	})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.HeaderHash)
	require.NotNil(t, res.Body)
	require.NotNil(t, res.Header)
	require.NotNil(t, res.SignersIndexes)
	require.NotNil(t, res.HeaderGasConsumption)
	require.NotNil(t, res.TransactionsPool)
}
