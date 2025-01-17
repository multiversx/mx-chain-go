package sync_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewShardForkDetectorFactory(t *testing.T) {
	t.Parallel()

	sfdf := sync.NewShardForkDetectorFactory()
	require.NotNil(t, sfdf)
	require.Implements(t, new(sync.ForkDetectorCreator), sfdf)
}

func TestShardForkDetectorFactory_CreateForkDetector(t *testing.T) {
	t.Parallel()

	sfdf := sync.NewShardForkDetectorFactory()
	args := sync.ForkDetectorFactoryArgs{
		RoundHandler:    &testscommon.RoundHandlerMock{},
		HeaderBlackList: &testscommon.TimeCacheStub{},
		BlockTracker:    &testscommon.BlockTrackerStub{},
		GenesisTime:     0,
	}

	forkDetector, err := sfdf.CreateForkDetector(args)
	require.Nil(t, err)
	require.NotNil(t, forkDetector)
	require.Implements(t, new(process.ForkDetector), forkDetector)
}

func TestShardForkDetectorFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	sfdf := sync.NewShardForkDetectorFactory()
	require.False(t, sfdf.IsInterfaceNil())
}
