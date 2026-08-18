package epochProviders

import (
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
	"github.com/multiversx/mx-chain-go/process"
)

var log = logger.GetOrCreate("dataretriever/factory/epochproviders")

// defaultAssumedPeersNumActivePersisters mirrors the NumActivePersisters default shipped to
// regular (pruning) nodes; used when the config predates AssumedPeersNumActivePersisters
const defaultAssumedPeersNumActivePersisters = 3

// CreateCurrentEpochProvider will create an instance of dataRetriever.CurrentNetworkEpochProviderHandler
func CreateCurrentEpochProvider(
	chainParametersHandler process.ChainParametersHandler,
	startTime int64,
	isFullArchive bool,
	enableEpochsHandler common.EnableEpochsHandler,
	assumedPeersNumActivePersisters uint32,
) (dataRetriever.CurrentNetworkEpochProviderHandler, error) {
	if !isFullArchive {
		return disabled.NewEpochProvider(), nil
	}

	if assumedPeersNumActivePersisters == 0 {
		assumedPeersNumActivePersisters = defaultAssumedPeersNumActivePersisters
		log.Warn("AssumedPeersNumActivePersisters is not configured, falling back to the default",
			"default", defaultAssumedPeersNumActivePersisters,
		)
	}

	arg := epochproviders.ArgArithmeticEpochProvider{
		ChainParametersHandler:          chainParametersHandler,
		StartTime:                       startTime,
		EnableEpochsHandler:             enableEpochsHandler,
		AssumedPeersNumActivePersisters: assumedPeersNumActivePersisters,
	}

	return epochproviders.NewArithmeticEpochProvider(arg)
}
