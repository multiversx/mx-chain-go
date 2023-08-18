package preprocess

import (
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/stretchr/testify/require"
)

func TestNewShardScheduledTxsExecutionFactory(t *testing.T) {
	t.Parallel()

	stef, err := NewShardScheduledTxsExecutionFactory()
	require.Nil(t, err)
	require.NotNil(t, stef)
	require.IsType(t, &shardScheduledTxsExecutionFactory{}, stef)
}

func TestShardScheduledTxsExecutionFactory_CreateScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	stef, _ := NewShardScheduledTxsExecutionFactory()

	stxeh, err := stef.CreateScheduledTxsExecutionHandler(ScheduledTxsExecutionFactoryArgs{})
	require.NotNil(t, err)
	require.Nil(t, stxeh)

	stxeh, err = stef.CreateScheduledTxsExecutionHandler(ScheduledTxsExecutionFactoryArgs{
		TxProcessor:      &testscommon.TxProcessorMock{},
		TxCoordinator:    &testscommon.TransactionCoordinatorMock{},
		Storer:           &genericMocks.StorerMock{},
		Marshalizer:      &testscommon.ProtoMarshalizerMock{},
		Hasher:           &testscommon.HasherStub{},
		ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
	})
	require.Nil(t, err)
	require.NotNil(t, stxeh)
	require.IsType(t, &scheduledTxsExecution{}, stxeh)
}

func TestShardScheduledTxsExecutionFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	stef, _ := NewShardScheduledTxsExecutionFactory()
	require.False(t, stef.IsInterfaceNil())
}
