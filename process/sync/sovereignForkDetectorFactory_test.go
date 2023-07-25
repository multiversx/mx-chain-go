package sync_test

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"testing"

	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignForkDetectorFactory(t *testing.T) {
	sfdf, err := sync.NewSovereignForkDetectorFactory(nil)
	require.Equal(t, process.ErrNilForkDetectorCreator, err)
	require.Nil(t, sfdf)

	sf, _ := sync.NewShardForkDetectorFactory()
	sfdf, err = sync.NewSovereignForkDetectorFactory(sf)
	require.Nil(t, err)
	require.NotNil(t, sfdf)
}

func TestSovereignForkDetectorFactory_CreateForkDetector(t *testing.T) {
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
}

func TestSovereignForkDetectorFactory_IsInterfaceNil(t *testing.T) {
	sf, _ := sync.NewShardForkDetectorFactory()
	sfdf, _ := sync.NewSovereignForkDetectorFactory(sf)
	require.False(t, sfdf.IsInterfaceNil())
}
