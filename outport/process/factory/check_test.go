package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/outport/process"
	"github.com/ElrondNetwork/elrond-go/outport/process/alteredaccounts"
	"github.com/ElrondNetwork/elrond-go/outport/process/transactionsfee"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/stretchr/testify/require"
)

func createArgCreateOutportDataProvider() ArgOutportDataProviderFactory {
	return ArgOutportDataProviderFactory{
		HasDrivers:             false,
		AddressConverter:       &testscommon.PubkeyConverterMock{},
		AccountsDB:             &state.AccountsStub{},
		Marshaller:             &testscommon.MarshalizerMock{},
		EsdtDataStorageHandler: &testscommon.EsdtStorageHandlerStub{},
		TransactionsStorer:     &genericMocks.StorerMock{},
		ShardCoordinator:       &testscommon.ShardsCoordinatorMock{},
		TxCoordinator:          &testscommon.TransactionCoordinatorMock{},
		NodesCoordinator:       &shardingMocks.NodesCoordinatorMock{},
		GasConsumedProvider:    &testscommon.GasHandlerStub{},
		EconomicsData:          &economicsmocks.EconomicsHandlerMock{},
	}
}

func TestCheckArgCreateOutportDataProvider(t *testing.T) {
	t.Parallel()

	arg := createArgCreateOutportDataProvider()
	arg.AddressConverter = nil
	require.Equal(t, alteredaccounts.ErrNilPubKeyConverter, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.AccountsDB = nil
	require.Equal(t, alteredaccounts.ErrNilAccountsDB, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.Marshaller = nil
	require.Equal(t, transactionsfee.ErrNilMarshaller, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.EsdtDataStorageHandler = nil
	require.Equal(t, alteredaccounts.ErrNilESDTDataStorageHandler, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.TransactionsStorer = nil
	require.Equal(t, transactionsfee.ErrNilStorage, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.ShardCoordinator = nil
	require.Equal(t, transactionsfee.ErrNilShardCoordinator, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.TxCoordinator = nil
	require.Equal(t, process.ErrNilTransactionCoordinator, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.NodesCoordinator = nil
	require.Equal(t, process.ErrNilNodesCoordinator, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.GasConsumedProvider = nil
	require.Equal(t, process.ErrNilGasConsumedProvider, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	arg.EconomicsData = nil
	require.Equal(t, transactionsfee.ErrNilTransactionFeeCalculator, checkArgCreateOutportDataProvider(arg))

	arg = createArgCreateOutportDataProvider()
	require.Nil(t, checkArgCreateOutportDataProvider(arg))
}
