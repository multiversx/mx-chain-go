package preprocess

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	commonMock "github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
)

func TestNewShardScheduledTxsExecutionFactory(t *testing.T) {
	t.Parallel()

	stef := NewShardScheduledTxsExecutionFactory()
	require.NotNil(t, stef)
	require.IsType(t, &shardScheduledTxsExecutionFactory{}, stef)
}

func TestShardScheduledTxsExecutionFactory_CreateScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	stef := NewShardScheduledTxsExecutionFactory()

	stxeh, err := stef.CreateScheduledTxsExecutionHandler(ScheduledTxsExecutionFactoryArgs{})
	require.NotNil(t, err)
	require.Nil(t, stxeh)

	stxeh, err = stef.CreateScheduledTxsExecutionHandler(ScheduledTxsExecutionFactoryArgs{
		TxProcessor:             &testscommon.TxProcessorMock{},
		TxCoordinator:           &testscommon.TransactionCoordinatorMock{},
		Storer:                  &genericMocks.StorerMock{},
		Marshalizer:             &testscommon.ProtoMarshalizerMock{},
		Hasher:                  &testscommon.HasherStub{},
		ShardCoordinator:        &testscommon.ShardsCoordinatorMock{},
		TxExecutionOrderHandler: &commonMock.TxExecutionOrderHandlerStub{},
	})
	require.Nil(t, err)
	require.NotNil(t, stxeh)
	require.IsType(t, &scheduledTxsExecution{}, stxeh)
}

func TestShardScheduledTxsExecutionFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	stef := NewShardScheduledTxsExecutionFactory()
	require.False(t, stef.IsInterfaceNil())
}
