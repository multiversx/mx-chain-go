package epochProviders

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/epochproviders"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/epochproviders/disabled"
)

// CreateCurrentEpochProvider will create an instance of dataRetriever.CurrentNetworkEpochProviderHandler
func CreateCurrentEpochProvider(
	generalConfigs config.Config,
	roundTimeInMilliseconds uint64,
	startTime int64,
) (dataRetriever.CurrentNetworkEpochProviderHandler, error) {
	if !generalConfigs.StoragePruning.FullArchive {
		return disabled.NewEpochProvider(), nil
	}

	arg := epochproviders.ArgArithmeticEpochProvider{
		RoundsPerEpoch:          uint32(generalConfigs.EpochStartConfig.RoundsPerEpoch),
		RoundTimeInMilliseconds: roundTimeInMilliseconds,
		StartTime:               startTime,
	}

	return epochproviders.NewArithmeticEpochProvider(arg)
}
