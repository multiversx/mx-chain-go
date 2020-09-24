package epochproviders_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/epochproviders"
	"github.com/stretchr/testify/require"
)

func TestNewCurrentNetworkEpochProvider(t *testing.T) {
	t.Parallel()

	cnep := epochproviders.NewCurrentNetworkEpochProvider(3)
	require.False(t, check.IfNil(cnep))
}

func TestCurrentNetworkEpochProvider_CurrentEpoch(t *testing.T) {
	t.Parallel()

	cnep := epochproviders.NewCurrentNetworkEpochProvider(3)

	expEpoch := uint32(37)
	cnep.SetCurrentEpoch(expEpoch)
	require.Equal(t, cnep.CurrentEpoch(), expEpoch)
}

func TestCurrentNetworkEpochProvider_EpochIsActiveInNetwork(t *testing.T) {
	t.Parallel()

	numActPersisters := 3
	cnep := epochproviders.NewCurrentNetworkEpochProvider(numActPersisters)

	tests := []struct {
		networkEpoch uint32
		nodeEpoch    uint32
		output       bool
		description  string
	}{
		{
			networkEpoch: 1,
			nodeEpoch:    0,
			output:       true,
			description:  "0 in [0, 1]",
		},
		{
			networkEpoch: 0,
			nodeEpoch:    0,
			output:       true,
			description:  "0 in [0, 0]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    0,
			output:       false,
			description:  "0 not in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    1,
			output:       false,
			description:  "1 not in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    2,
			output:       false,
			description:  "2 not in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    3,
			output:       true,
			description:  "3 in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    4,
			output:       true,
			description:  "4 in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    5,
			output:       true,
			description:  "5 in (2, 5]",
		},
		{
			networkEpoch: 5,
			nodeEpoch:    6,
			output:       false,
			description:  "6 not in (3, 5]",
		},
	}

	for _, tt := range tests {
		cnep.SetCurrentEpoch(tt.networkEpoch)
		testOk := tt.output == cnep.EpochIsActiveInNetwork(tt.nodeEpoch)
		if !testOk {
			require.Failf(t, "%s for epoch param %d and network epoch %d",
				tt.description,
				tt.nodeEpoch,
				tt.networkEpoch,
			)
		}
	}
}
