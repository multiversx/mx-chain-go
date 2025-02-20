package latestData

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignLatestDataProvider(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		sldp, err := NewSovereignLatestDataProvider(getLatestDataProviderArgs())
		require.Nil(t, err)
		require.False(t, sldp.IsInterfaceNil())
		require.Equal(t, "*latestData.sovereignEpochStartRoundLoader", fmt.Sprintf("%T", sldp.epochStartRoundLoader))
	})

	t.Run("nil input, should return error", func(t *testing.T) {
		args := getLatestDataProviderArgs()
		args.DirectoryReader = nil
		sldp, err := NewSovereignLatestDataProvider(args)
		require.Nil(t, sldp)
		require.Equal(t, storage.ErrNilDirectoryReader, err)
	})
}
