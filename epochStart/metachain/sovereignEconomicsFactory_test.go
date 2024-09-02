package metachain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignEconomicsFactory_CreateEndOfEpochEconomics(t *testing.T) {
	t.Parallel()

	f := NewSovereignEconomicsFactory()
	require.False(t, f.IsInterfaceNil())

	t.Run("should work", func(t *testing.T) {
		args := createMockEpochEconomicsArguments()
		econ, err := f.CreateEndOfEpochEconomics(args)
		require.Nil(t, err)
		require.Equal(t, fmt.Sprintf("%T", econ), "*metachain.sovereignEconomics")
	})
	t.Run("nil input args, should not work", func(t *testing.T) {
		args := createMockEpochEconomicsArguments()
		args.ShardCoordinator = nil
		econ, err := f.CreateEndOfEpochEconomics(args)
		require.Nil(t, econ)
		require.NotNil(t, err)
	})
}
