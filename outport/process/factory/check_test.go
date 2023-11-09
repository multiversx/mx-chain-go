package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-go/outport/process"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts"
	"github.com/multiversx/mx-chain-go/outport/process/transactionsfee"
	"github.com/multiversx/mx-chain-go/testscommon"
	commonMocks "github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

func createArgOutportDataProviderFactory() ArgOutportDataProviderFactory {
	return ArgOutportDataProviderFactory{
		HasDrivers:             false,
		AddressConverter:       testscommon.NewPubkeyConverterMock(32),
		AccountsDB:             &state.AccountsStub{},
		Marshaller:             &marshallerMock.MarshalizerMock{},
		EsdtDataStorageHandler: &testscommon.EsdtStorageHandlerStub{},
		TransactionsStorer:     &genericMocks.StorerMock{},
		ShardCoordinator:       &testscommon.ShardsCoordinatorMock{},
		TxCoordinator:          &testscommon.TransactionCoordinatorMock{},
		NodesCoordinator:       &shardingMocks.NodesCoordinatorMock{},
		GasConsumedProvider:    &testscommon.GasHandlerStub{},
		EconomicsData:          &economicsmocks.EconomicsHandlerMock{},
		Hasher:                 &testscommon.KeccakMock{},
		MbsStorer:              &genericMocks.StorerMock{},
		EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		ExecutionOrderGetter:   &commonMocks.TxExecutionOrderHandlerStub{},
	}
}

func TestCheckArgCreateOutportDataProvider(t *testing.T) {
	t.Parallel()

	arg := createArgOutportDataProviderFactory()
	arg.AddressConverter = nil
	require.Equal(t, alteredaccounts.ErrNilPubKeyConverter, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.AccountsDB = nil
	require.Equal(t, alteredaccounts.ErrNilAccountsDB, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.Marshaller = nil
	require.Equal(t, transactionsfee.ErrNilMarshaller, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.EsdtDataStorageHandler = nil
	require.Equal(t, alteredaccounts.ErrNilESDTDataStorageHandler, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.TransactionsStorer = nil
	require.Equal(t, transactionsfee.ErrNilStorage, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.ShardCoordinator = nil
	require.Equal(t, transactionsfee.ErrNilShardCoordinator, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.TxCoordinator = nil
	require.Equal(t, process.ErrNilTransactionCoordinator, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.NodesCoordinator = nil
	require.Equal(t, process.ErrNilNodesCoordinator, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.GasConsumedProvider = nil
	require.Equal(t, process.ErrNilGasConsumedProvider, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.EconomicsData = nil
	require.Equal(t, transactionsfee.ErrNilTransactionFeeCalculator, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	arg.Hasher = nil
	require.Equal(t, process.ErrNilHasher, checkArgOutportDataProviderFactory(arg))

	arg = createArgOutportDataProviderFactory()
	require.Nil(t, checkArgOutportDataProviderFactory(arg))
}
