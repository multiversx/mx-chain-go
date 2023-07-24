package track_test

import (
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewShardBlockTrackerFactory(t *testing.T) {
	sbtcf, err := track.NewShardBlockTrackerFactory()
	require.Nil(t, err)
	require.NotNil(t, sbtcf)
}

func TestShardBlockTrackerFactory_CreateBlockTracker(t *testing.T) {
	sbtcf, _ := track.NewShardBlockTrackerFactory()

	bt, err := sbtcf.CreateBlockTracker(track.ArgShardTracker{})
	require.NotNil(t, err)
	require.Nil(t, bt)

	shardArguments := CreateShardTrackerMockArguments()
	bt, err = sbtcf.CreateBlockTracker(shardArguments)
	require.Nil(t, err)
	require.NotNil(t, bt)
}

func TestShardBlockTrackerFactory_IsInterfaceNil(t *testing.T) {
	sbtcf, _ := track.NewShardBlockTrackerFactory()
	require.False(t, sbtcf.IsInterfaceNil())
}
