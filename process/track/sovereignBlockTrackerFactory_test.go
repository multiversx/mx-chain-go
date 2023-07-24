package track_test

import (
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSovereignBlockTrackerFactory(t *testing.T) {
	sbtcf, err := track.NewSovereignBlockTrackerFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, sbtcf)

	sf, _ := track.NewShardBlockTrackerFactory()
	sbtcf, err = track.NewSovereignBlockTrackerFactory(sf)
	require.Nil(t, err)
	require.NotNil(t, sbtcf)
}

func TestSovereignBlockTrackerFactory_CreateBlockTracker(t *testing.T) {
	sf, _ := track.NewShardBlockTrackerFactory()
	sbtcf, _ := track.NewSovereignBlockTrackerFactory(sf)

	bt, err := sbtcf.CreateBlockTracker(track.ArgShardTracker{})
	require.NotNil(t, err)
	require.Nil(t, bt)

	shardArguments := CreateShardTrackerMockArguments()
	shardArguments.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}
	bt, err = sbtcf.CreateBlockTracker(shardArguments)
	require.Nil(t, err)
	require.NotNil(t, bt)
}

func TestSovereignBlockTrackerFactory_IsInterfaceNil(t *testing.T) {
	sf, _ := track.NewShardBlockTrackerFactory()
	sbtcf, _ := track.NewSovereignBlockTrackerFactory(sf)

	require.False(t, sbtcf.IsInterfaceNil())
}
