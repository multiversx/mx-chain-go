package track_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/track"
)

func TestNewShardBlockTrackerFactory(t *testing.T) {
	t.Parallel()

	sbtcf := track.NewShardBlockTrackerFactory()
	require.NotNil(t, sbtcf)
	require.Implements(t, new(track.BlockTrackerCreator), sbtcf)
}

func TestShardBlockTrackerFactory_CreateBlockTracker(t *testing.T) {
	t.Parallel()

	sbtcf := track.NewShardBlockTrackerFactory()

	bt, err := sbtcf.CreateBlockTracker(track.ArgShardTracker{})
	require.NotNil(t, err)
	require.Nil(t, bt)

	shardArguments := CreateShardTrackerMockArguments()
	bt, err = sbtcf.CreateBlockTracker(shardArguments)
	require.Nil(t, err)
	require.NotNil(t, bt)
	require.Implements(t, new(process.BlockTracker), bt)
}

func TestShardBlockTrackerFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sbtcf := track.NewShardBlockTrackerFactory()
	require.False(t, sbtcf.IsInterfaceNil())
}
