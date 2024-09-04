package metachain

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignEconomics(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createMockEpochEconomicsArguments()
		ec, _ := NewEndOfEpochEconomicsDataCreator(args)
		sovEc, err := NewSovereignEconomics(ec)
		require.Nil(t, err)
		require.False(t, sovEc.IsInterfaceNil())
	})
	t.Run("nil input, should fail", func(t *testing.T) {
		sovEc, err := NewSovereignEconomics(nil)
		require.Equal(t, process.ErrNilEconomicsData, err)
		require.Nil(t, sovEc)
	})

}
