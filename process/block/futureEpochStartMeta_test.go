package block

import (
	"testing"

	coreBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestCheckFutureEpochStartMeta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		candidateShardEpoch uint32
		metaEpoch           uint32
		isEpochStart        bool
		expectedError       error
	}{
		{
			name:                "ordinary future meta header is accepted",
			candidateShardEpoch: 7,
			metaEpoch:           8,
		},
		{
			name:                "older epoch start meta header is accepted",
			candidateShardEpoch: 7,
			metaEpoch:           6,
			isEpochStart:        true,
		},
		{
			name:                "same epoch start meta header is accepted",
			candidateShardEpoch: 7,
			metaEpoch:           7,
			isEpochStart:        true,
		},
		{
			name:                "future epoch start meta header is rejected",
			candidateShardEpoch: 7,
			metaEpoch:           8,
			isEpochStart:        true,
			expectedError:       errFutureEpochStartMetaHeader,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			metaHeader := &coreBlock.MetaBlock{
				Epoch: test.metaEpoch,
			}
			if test.isEpochStart {
				metaHeader.EpochStart = coreBlock.EpochStart{
					LastFinalizedHeaders: []coreBlock.EpochStartShardData{{}},
				}
			}

			err := checkFutureEpochStartMeta(test.candidateShardEpoch, metaHeader)
			require.ErrorIs(t, err, test.expectedError)
		})
	}
}
