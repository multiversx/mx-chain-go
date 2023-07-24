package preprocess

import (
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewShardScheduledTxsExecutionFactory(t *testing.T) {
	stef, err := NewShardScheduledTxsExecutionFactory()
	require.Nil(t, err)
	require.NotNil(t, stef)
}

func TestShardScheduledTxsExecutionFactory_CreateScheduledTxsExecutionHandler(t *testing.T) {
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
}

func TestShardScheduledTxsExecutionFactory_IsInterfaceNil(t *testing.T) {
	stef, _ := NewShardScheduledTxsExecutionFactory()
	require.False(t, stef.IsInterfaceNil())
}
