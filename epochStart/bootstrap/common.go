package bootstrap

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ShouldSyncWithTheNetwork returns true if a peer is not synced with the latest epoch (especially used when a peer
// wants to join the network after the genesis)
func ShouldSyncWithTheNetwork(
	startTime time.Time,
	epochFoundInStorage bool,
	nodesConfig *sharding.NodesSetup,
	config *config.Config,
) bool {
	isCurrentTimeBeforeGenesis := time.Now().Sub(startTime) < 0
	timeInFirstEpochAtMinRoundsPerEpoch := startTime.Add(time.Duration(nodesConfig.RoundDuration *
		uint64(config.EpochStartConfig.MinRoundsBetweenEpochs)))
	isEpochZero := time.Now().Sub(timeInFirstEpochAtMinRoundsPerEpoch) < 0
	shouldSyncWithTheNetwork := !isCurrentTimeBeforeGenesis && !isEpochZero && !epochFoundInStorage

	return shouldSyncWithTheNetwork
}
