package coordinator

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSovereignTransactionCoordinatorFactory_NewShardTransactionCoordinatorFactory(t *testing.T) {
	t.Parallel()

	sovtcf, err := NewSovereignTransactionCoordinatorFactory(nil)
	require.Equal(t, process.ErrNilTransactionCoordinatorCreator, err)
	require.Nil(t, sovtcf)

	stcf, _ := NewShardTransactionCoordinatorFactory()
	sovtcf, err = NewSovereignTransactionCoordinatorFactory(stcf)

	require.Nil(t, err)
	require.NotNil(t, sovtcf)
	require.IsType(t, new(sovereignTransactionCoordinatorFactory), sovtcf)
}

func TestSovereignTransactionCoordinatorFactory_CreateTransactionCoordinator(t *testing.T) {
	t.Parallel()

	stcf, _ := NewShardTransactionCoordinatorFactory()
	sovtcf, err := NewSovereignTransactionCoordinatorFactory(stcf)
	tc, err := sovtcf.CreateTransactionCoordinator(ArgTransactionCoordinator{})
	require.NotNil(t, err)
	require.Nil(t, tc)

	tc, err = sovtcf.CreateTransactionCoordinator(createMockTransactionCoordinatorArguments())
	require.Nil(t, err)
	require.NotNil(t, tc)
	require.IsType(t, new(sovereignChainTransactionCoordinator), tc)
}

func TestSovereignTransactionCoordinatorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	stcf, _ := NewShardTransactionCoordinatorFactory()
	sovtcf, _ := NewSovereignTransactionCoordinatorFactory(stcf)
	require.False(t, sovtcf.IsInterfaceNil())
}
