package epochStartTrigger

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestSovereignEpochStartTriggerFactory_CreateEpochStartTrigger(t *testing.T) {
	t.Parallel()

	f := NewSovereignEpochStartTriggerFactory()
	require.False(t, f.IsInterfaceNil())

	args := createArgs(core.SovereignChainShardId)
	trigger, err := f.CreateEpochStartTrigger(args)
	require.Nil(t, err)
	require.Equal(t, "*metachain.sovereignTrigger", fmt.Sprintf("%T", trigger))
}
