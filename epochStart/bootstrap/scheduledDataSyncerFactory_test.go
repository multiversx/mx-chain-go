package bootstrap

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/types"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	epochStartMocks "github.com/ElrondNetwork/elrond-go/testscommon/bootstrapMocks/epochStart"
	"github.com/ElrondNetwork/elrond-go/testscommon/syncer"
	"github.com/stretchr/testify/require"
)

func TestNewScheduledDataSyncerFactory(t *testing.T) {
	sdsFactory := NewScheduledDataSyncerFactory()
	require.NotNil(t, sdsFactory)
}

func TestScheduledDataSyncerFactory_CreateNilScheduledTxsHandler(t *testing.T) {
	sdsFactory := NewScheduledDataSyncerFactory()
	args := createDefaultDataSyncerFactoryArgs()
	args.ScheduledTxsHandler = nil

	dataSyncer, err := sdsFactory.Create(args)
	require.Nil(t, dataSyncer)
	require.Equal(t, epochStart.ErrNilScheduledTxsHandler, err)
}

func TestScheduledDataSyncerFactory_CreateNilHeadersSyncer(t *testing.T) {
	sdsFactory := NewScheduledDataSyncerFactory()
	args := createDefaultDataSyncerFactoryArgs()
	args.HeadersSyncer = nil

	dataSyncer, err := sdsFactory.Create(args)
	require.Nil(t, dataSyncer)
	require.Equal(t, epochStart.ErrNilHeadersSyncer, err)
}

func TestScheduledDataSyncerFactory_CreateNilMiniBlocksSyncer(t *testing.T) {
	sdsFactory := NewScheduledDataSyncerFactory()
	args := createDefaultDataSyncerFactoryArgs()
	args.MiniBlocksSyncer = nil

	dataSyncer, err := sdsFactory.Create(args)
	require.Nil(t, dataSyncer)
	require.Equal(t, epochStart.ErrNilMiniBlocksSyncer, err)
}

func TestScheduledDataSyncerFactory_CreateNilTxSyncer(t *testing.T) {
	sdsFactory := NewScheduledDataSyncerFactory()
	args := createDefaultDataSyncerFactoryArgs()
	args.TxSyncer = nil

	dataSyncer, err := sdsFactory.Create(args)
	require.Nil(t, dataSyncer)
	require.Equal(t, epochStart.ErrNilTransactionsSyncer, err)
}

func TestScheduledDataSyncerFactory_Create(t *testing.T) {
	sdsFactory := NewScheduledDataSyncerFactory()
	args := createDefaultDataSyncerFactoryArgs()

	dataSyncer, err := sdsFactory.Create(args)
	require.Nil(t, err)
	require.NotNil(t, dataSyncer)
}

func TestScheduledDataSyncerFactory_IsInterfaceNil(t *testing.T) {
	var sdsFactory *ScheduledDataSyncerFactory = nil
	require.True(t, sdsFactory.IsInterfaceNil())

	sdsFactory = &ScheduledDataSyncerFactory{}
	require.False(t, sdsFactory.IsInterfaceNil())
}

func createDefaultDataSyncerFactoryArgs() *types.ScheduledDataSyncerCreateArgs {
	return &types.ScheduledDataSyncerCreateArgs{
		ScheduledTxsHandler:  &testscommon.ScheduledTxsExecutionStub{},
		HeadersSyncer:        &epochStartMocks.HeadersByHashSyncerStub{},
		MiniBlocksSyncer:     &epochStartMocks.PendingMiniBlockSyncHandlerStub{},
		TxSyncer:             &syncer.TransactionsSyncHandlerMock{},
		ScheduledEnableEpoch: 0,
	}
}
