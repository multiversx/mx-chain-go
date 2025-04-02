package common_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/require"
)

func TestGetEpochForConsensus(t *testing.T) {
	t.Parallel()

	providedEpoch := uint32(10)
	proof := &block.HeaderProof{
		HeaderEpoch:    providedEpoch,
		IsStartOfEpoch: false,
	}

	epoch := common.GetEpochForConsensus(proof)
	require.Equal(t, providedEpoch, epoch)

	proof = &block.HeaderProof{
		HeaderEpoch:    providedEpoch,
		IsStartOfEpoch: true,
	}

	epoch = common.GetEpochForConsensus(proof)
	require.Equal(t, providedEpoch-1, epoch)
}
