package bootstrap

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/stretchr/testify/require"
)

func TestNewEpochStartSovereignSyncer(t *testing.T) {
	t.Parallel()

	args := getEpochStartSyncerArgs()
	sovSyncer, err := newEpochStartSovereignSyncer(args)
	require.Nil(t, err)
	require.Equal(t, "*bootstrap.epochStartSovereignSyncer", fmt.Sprintf("%T", sovSyncer))
	require.Equal(t, "*bootstrap.sovereignTopicProvider", fmt.Sprintf("%T", sovSyncer.epochStartTopicProviderHandler))
	require.Equal(t, fmt.Sprintf("%s_%d", factory.ShardBlocksTopic, core.SovereignChainShardId), sovSyncer.epochStartTopicProviderHandler.getTopic())
}
