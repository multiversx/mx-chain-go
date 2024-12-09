package syncer_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignValidatorAccountsSyncer(t *testing.T) {
	t.Parallel()

	args := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: getDefaultBaseAccSyncerArgs(),
	}

	sovSyncer, err := syncer.NewSovereignValidatorAccountsSyncer(args)
	require.Nil(t, err)
	require.NotNil(t, sovSyncer)
	require.Equal(t, core.SovereignChainShardId, sovSyncer.GetShardID())
}
