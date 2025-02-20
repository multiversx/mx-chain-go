package rating

import (
	"math"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/require"
)

func createSovereignRatingsDataArgs() RatingsDataArg {
	ratingsDataArg := createDymmyRatingsData()
	ratingsDataArg.Config = createDummyRatingsConfig()
	ratingsDataArg.Config.MetaChain = config.MetaChain{}

	return ratingsDataArg
}

func TestNewSovereignRatingsData(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		args := createSovereignRatingsDataArgs()
		srd, err := NewSovereignRatingsData(args)
		require.Nil(t, err)
		require.NotNil(t, srd)
		require.Equal(t, srd.shardRatingsStepData, srd.metaRatingsStepData)
	})

	t.Run("error checking sovereign ratings config", func(t *testing.T) {
		args := createSovereignRatingsDataArgs()
		args.Config.General.MinRating = 0
		srd, err := NewSovereignRatingsData(args)
		require.Equal(t, process.ErrMinRatingSmallerThanOne, err)
		require.Nil(t, srd)
	})

	t.Run("error computing rating step", func(t *testing.T) {
		args := createSovereignRatingsDataArgs()
		args.Config.ShardChain.ProposerDecreaseFactor = math.MinInt32
		srd, err := NewSovereignRatingsData(args)
		require.True(t, strings.Contains(err.Error(), process.ErrOverflow.Error()))
		require.Nil(t, srd)
	})
}
