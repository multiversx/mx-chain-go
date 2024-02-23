package sync_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignForkDetectorFactory(t *testing.T) {
	t.Parallel()

	sfdf, err := sync.NewSovereignForkDetectorFactory(nil)
	require.Equal(t, process.ErrNilForkDetectorCreator, err)
	require.Nil(t, sfdf)

	sf, _ := sync.NewShardForkDetectorFactory()
	sfdf, err = sync.NewSovereignForkDetectorFactory(sf)
	require.Nil(t, err)
	require.NotNil(t, sfdf)
	require.Implements(t, new(sync.ForkDetectorCreator), sf)
}

func TestSovereignForkDetectorFactory_CreateForkDetector(t *testing.T) {
	t.Parallel()

	sf, _ := sync.NewShardForkDetectorFactory()
	sfdf, _ := sync.NewSovereignForkDetectorFactory(sf)

	forkDetector, err := sfdf.CreateForkDetector(sync.ForkDetectorFactoryArgs{})
	require.Nil(t, forkDetector)
	require.NotNil(t, err)

	args := sync.ForkDetectorFactoryArgs{
		RoundHandler:    &testscommon.RoundHandlerMock{},
		HeaderBlackList: &testscommon.TimeCacheStub{},
		BlockTracker:    &testscommon.BlockTrackerStub{},
		GenesisTime:     0,
	}
	forkDetector, err = sfdf.CreateForkDetector(args)
	require.Nil(t, err)
	require.NotNil(t, forkDetector)
	require.Implements(t, new(process.ForkDetector), forkDetector)
}

func TestSovereignForkDetectorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sf, _ := sync.NewShardForkDetectorFactory()
	sfdf, _ := sync.NewSovereignForkDetectorFactory(sf)
	require.False(t, sfdf.IsInterfaceNil())
}
