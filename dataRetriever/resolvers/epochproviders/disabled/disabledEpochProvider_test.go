package disabled

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEpochProvider(t *testing.T) {
	t.Parallel()

	var ep *epochProvider
	require.True(t, ep.IsInterfaceNil())

	ep = NewEpochProvider()
	require.False(t, ep.IsInterfaceNil())
	require.True(t, ep.EpochIsActiveInNetwork(1))
	ep.EpochConfirmed(0, 0)
}
