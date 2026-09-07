package disabled

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDisabledEpochRewards(t *testing.T) {
	t.Parallel()

	rewardsGetter := NewDisabledEpochRewards()
	require.NotNil(t, rewardsGetter)
	require.False(t, rewardsGetter.IsInterfaceNil())

	res := rewardsGetter.GetRewardsTxs(nil)
	require.NotNil(t, res)
}
