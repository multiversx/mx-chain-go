package coordinator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShardTransactionCoordinatorFactory_NewShardTransactionCoordinatorFactory(t *testing.T) {
	t.Parallel()

	tcf, err := NewShardTransactionCoordinatorFactory()

	require.Nil(t, err)
	require.NotNil(t, tcf)
	require.IsType(t, new(shardTransactionCoordinatorFactory), tcf)
}

func TestShardTransactionCoordinatorFactory_CreateTransactionCoordinator(t *testing.T) {
	t.Parallel()

	tcf, _ := NewShardTransactionCoordinatorFactory()
	tc, err := tcf.CreateTransactionCoordinator(ArgTransactionCoordinator{})
	require.NotNil(t, err)
	require.Nil(t, tc)

	tc, err = tcf.CreateTransactionCoordinator(createMockTransactionCoordinatorArguments())
	require.Nil(t, err)
	require.NotNil(t, tc)
	require.IsType(t, new(transactionCoordinator), tc)
}

func TestShardTransactionCoordinatorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	tcf, _ := NewShardTransactionCoordinatorFactory()
	require.False(t, tcf.IsInterfaceNil())
}
