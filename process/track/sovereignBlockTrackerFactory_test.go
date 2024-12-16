package track_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewSovereignBlockTrackerFactory(t *testing.T) {
	t.Parallel()

	sbtcf, err := track.NewSovereignBlockTrackerFactory(nil)
	require.NotNil(t, err)
	require.Nil(t, sbtcf)

	sf := track.NewShardBlockTrackerFactory()
	sbtcf, err = track.NewSovereignBlockTrackerFactory(sf)
	require.Nil(t, err)
	require.NotNil(t, sbtcf)
	require.Implements(t, new(track.BlockTrackerCreator), sbtcf)
}

func TestSovereignBlockTrackerFactory_CreateBlockTracker(t *testing.T) {
	t.Parallel()

	sf := track.NewShardBlockTrackerFactory()
	sbtcf, _ := track.NewSovereignBlockTrackerFactory(sf)

	bt, err := sbtcf.CreateBlockTracker(track.ArgShardTracker{})
	require.NotNil(t, err)
	require.Nil(t, bt)

	shardArguments := CreateShardTrackerMockArguments()
	shardArguments.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}
	bt, err = sbtcf.CreateBlockTracker(shardArguments)
	require.Nil(t, err)
	require.NotNil(t, bt)
	require.Implements(t, new(process.BlockTracker), bt)
}

func TestSovereignBlockTrackerFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sf := track.NewShardBlockTrackerFactory()
	sbtcf, _ := track.NewSovereignBlockTrackerFactory(sf)

	require.False(t, sbtcf.IsInterfaceNil())
}
