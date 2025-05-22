package epochProviders

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
	"github.com/multiversx/mx-chain-go/process"
)

// CreateCurrentEpochProvider will create an instance of dataRetriever.CurrentNetworkEpochProviderHandler
func CreateCurrentEpochProvider(
	chainParametersHandler process.ChainParametersHandler,
	startTime int64,
	isFullArchive bool,
) (dataRetriever.CurrentNetworkEpochProviderHandler, error) {
	if !isFullArchive {
		return disabled.NewEpochProvider(), nil
	}

	arg := epochproviders.ArgArithmeticEpochProvider{
		ChainParametersHandler: chainParametersHandler,
		StartTime:              startTime,
	}

	return epochproviders.NewArithmeticEpochProvider(arg)
}
