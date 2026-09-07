package metachain

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func Test_GetEpochToUseEpochStartData(t *testing.T) {
	t.Parallel()

	t.Run("should work correctly for header v3", func(t *testing.T) {
		t.Parallel()

		epoch := GetEpochToUseEpochStartData(&block.HeaderV3{
			Epoch: 1,
		})
		require.Equal(t, uint32(2), epoch)
	})

	t.Run("should work correctly for header v2", func(t *testing.T) {
		t.Parallel()

		epoch := GetEpochToUseEpochStartData(&block.HeaderV2{
			Header: &block.Header{
				Epoch: 1,
			},
		})
		require.Equal(t, uint32(1), epoch)
	})

	t.Run("should work correctly for header v1", func(t *testing.T) {
		t.Parallel()

		epoch := GetEpochToUseEpochStartData(&block.Header{
			Epoch: 1,
		})
		require.Equal(t, uint32(1), epoch)
	})
}
