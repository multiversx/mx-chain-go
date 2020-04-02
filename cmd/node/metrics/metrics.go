package metrics

import (
	"errors"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/appStatusPolling"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const millisecondsInSecond = 1000

// InitMetrics will init metrics for status handler
func InitMetrics(
	appStatusHandler core.AppStatusHandler,
	pubKey crypto.PublicKey,
	nodeType core.NodeType,
	shardCoordinator sharding.Coordinator,
	nodesConfig *sharding.NodesSetup,
	version string,
	economicsConfig *config.EconomicsConfig,
	roundsPerEpoch int64,
) {
	shardId := uint64(shardCoordinator.SelfId())
	roundDuration := nodesConfig.RoundDuration
	isSyncing := uint64(1)
	initUint := uint64(0)
	initString := ""

	appStatusHandler.SetStringValue(core.MetricPublicKeyBlockSign, factory.GetPkEncoded(pubKey))
	appStatusHandler.SetUInt64Value(core.MetricShardId, shardId)
	appStatusHandler.SetStringValue(core.MetricNodeType, string(nodeType))
	appStatusHandler.SetUInt64Value(core.MetricRoundTime, roundDuration/millisecondsInSecond)
	appStatusHandler.SetStringValue(core.MetricAppVersion, version)
	appStatusHandler.SetUInt64Value(core.MetricRoundsPerEpoch, uint64(roundsPerEpoch))
	appStatusHandler.SetUInt64Value(core.MetricCountConsensus, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountLeader, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, initUint)
	appStatusHandler.SetStringValue(core.MetricConsensusState, initString)
	appStatusHandler.SetStringValue(core.MetricConsensusRoundState, initString)
	appStatusHandler.SetStringValue(core.MetricCrossCheckBlockHeight, "0")
	appStatusHandler.SetUInt64Value(core.MetricIsSyncing, isSyncing)
	appStatusHandler.SetStringValue(core.MetricCurrentBlockHash, initString)
	appStatusHandler.SetUInt64Value(core.MetricNumProcessedTxs, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCurrentRoundTimestamp, initUint)
	appStatusHandler.SetUInt64Value(core.MetricHeaderSize, initUint)
	appStatusHandler.SetUInt64Value(core.MetricMiniBlocksSize, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersFromPool, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumShardHeadersProcessed, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNumTimesInForkChoice, initUint)
	appStatusHandler.SetUInt64Value(core.MetricHighestFinalBlockInShard, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountConsensusAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(core.MetricRoundAtEpochStart, initUint)
	appStatusHandler.SetUInt64Value(core.MetricRoundsPassedInCurrentEpoch, initUint)
	appStatusHandler.SetStringValue(core.MetricLeaderPercentage, fmt.Sprintf("%f", economicsConfig.RewardsSettings.LeaderPercentage))
	appStatusHandler.SetStringValue(core.MetricDenominationCoefficient, economicsConfig.RewardsSettings.DenominationCoefficientForView)
	appStatusHandler.SetUInt64Value(core.MetricNumConnectedPeers, initUint)
	appStatusHandler.SetStringValue(core.MetricNumConnectedPeersClassification, initString)

	appStatusHandler.SetStringValue(core.MetricP2PNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(core.MetricP2PPeerInfo, initString)
	appStatusHandler.SetStringValue(core.MetricP2PIntraShardValidators, initString)
	appStatusHandler.SetStringValue(core.MetricP2PIntraShardObservers, initString)
	appStatusHandler.SetStringValue(core.MetricP2PCrossShardValidators, initString)
	appStatusHandler.SetStringValue(core.MetricP2PCrossShardObservers, initString)
	appStatusHandler.SetStringValue(core.MetricP2PUnknownPeers, initString)

	var consensusGroupSize uint32
	switch {
	case shardCoordinator.SelfId() < shardCoordinator.NumberOfShards():
		consensusGroupSize = nodesConfig.ConsensusGroupSize
	case shardCoordinator.SelfId() == core.MetachainShardId:
		consensusGroupSize = nodesConfig.MetaChainConsensusGroupSize
	default:
		consensusGroupSize = 0
	}

	validatorsNodes, _ := nodesConfig.InitialNodesInfo()
	numValidators := len(validatorsNodes[shardCoordinator.SelfId()])

	appStatusHandler.SetUInt64Value(core.MetricNumValidators, uint64(numValidators))
	appStatusHandler.SetUInt64Value(core.MetricConsensusGroupSize, uint64(consensusGroupSize))
}

// SaveUint64Metric will save a uint64 metric in status handler
func SaveUint64Metric(ash core.AppStatusHandler, key string, value uint64) {
	ash.SetUInt64Value(key, value)
}

// SaveStringMetric will save a string metric in status handler
func SaveStringMetric(ash core.AppStatusHandler, key, value string) {
	ash.SetStringValue(key, value)
}

// StartStatusPolling will start save information in status handler about network
func StartStatusPolling(
	ash core.AppStatusHandler,
	pollingInterval time.Duration,
	networkComponents *factory.Network,
	processComponents *factory.Process,
) error {

	if ash == nil {
		return errors.New("nil AppStatusHandler")
	}

	appStatusPollingHandler, err := appStatusPolling.NewAppStatusPolling(ash, pollingInterval)
	if err != nil {
		return errors.New("cannot init AppStatusPolling")
	}

	err = registerPollConnectedPeers(appStatusPollingHandler, networkComponents)
	if err != nil {
		return err
	}

	err = registerPollProbableHighestNonce(appStatusPollingHandler, processComponents)
	if err != nil {
		return err
	}

	appStatusPollingHandler.Poll()

	return nil
}

func registerPollConnectedPeers(
	appStatusPollingHandler *appStatusPolling.AppStatusPolling,
	networkComponents *factory.Network,
) error {

	p2pMetricsHandlerFunc := func(appStatusHandler core.AppStatusHandler) {
		computeNumConnectedPeers(appStatusHandler, networkComponents)
		computeConnectedPeers(appStatusHandler, networkComponents)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(p2pMetricsHandlerFunc)
	if err != nil {
		return errors.New("cannot register handler func for num of connected peers")
	}

	return nil
}

func computeNumConnectedPeers(
	appStatusHandler core.AppStatusHandler,
	networkComponents *factory.Network,
) {
	numOfConnectedPeers := uint64(len(networkComponents.NetMessenger.ConnectedAddresses()))
	appStatusHandler.SetUInt64Value(core.MetricNumConnectedPeers, numOfConnectedPeers)
}

func computeConnectedPeers(
	appStatusHandler core.AppStatusHandler,
	networkComponents *factory.Network,
) {
	peersInfo := networkComponents.NetMessenger.GetConnectedPeersInfo()

	peerClassification := fmt.Sprintf("intraVal:%d,crossVal:%d,intraObs:%d,crossObs:%d,unknown:%d,",
		len(peersInfo.IntraShardValidators),
		len(peersInfo.CrossShardValidators),
		len(peersInfo.IntraShardObservers),
		len(peersInfo.CrossShardObservers),
		len(peersInfo.UnknownPeers),
	)
	appStatusHandler.SetStringValue(core.MetricNumConnectedPeersClassification, peerClassification)
	appStatusHandler.SetStringValue(core.MetricP2PNumConnectedPeersClassification, peerClassification)

	setP2pConnectedPeersMetrics(appStatusHandler, peersInfo)
	setCurrentP2pNodeAddresses(appStatusHandler, networkComponents)
}

func setP2pConnectedPeersMetrics(appStatusHandler core.AppStatusHandler, info *p2p.ConnectedPeersInfo) {
	appStatusHandler.SetStringValue(core.MetricP2PUnknownPeers, sliceToString(info.UnknownPeers))
	appStatusHandler.SetStringValue(core.MetricP2PIntraShardValidators, sliceToString(info.IntraShardValidators))
	appStatusHandler.SetStringValue(core.MetricP2PIntraShardObservers, sliceToString(info.IntraShardObservers))
	appStatusHandler.SetStringValue(core.MetricP2PCrossShardValidators, sliceToString(info.CrossShardValidators))
	appStatusHandler.SetStringValue(core.MetricP2PCrossShardObservers, sliceToString(info.CrossShardObservers))
}

func sliceToString(input []string) string {
	output := ""
	for _, str := range input {
		output += str + ","
	}

	return output
}

func setCurrentP2pNodeAddresses(
	appStatusHandler core.AppStatusHandler,
	networkComponents *factory.Network,
) {
	appStatusHandler.SetStringValue(core.MetricP2PPeerInfo, sliceToString(networkComponents.NetMessenger.Addresses()))
}

func registerPollProbableHighestNonce(
	appStatusPollingHandler *appStatusPolling.AppStatusPolling,
	processComponents *factory.Process,
) error {

	probableHighestNonceHandlerFunc := func(appStatusHandler core.AppStatusHandler) {
		probableHigherNonce := processComponents.ForkDetector.ProbableHighestNonce()
		appStatusHandler.SetUInt64Value(core.MetricProbableHighestNonce, probableHigherNonce)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(probableHighestNonceHandlerFunc)
	if err != nil {
		return errors.New("cannot register handler func for forkdetector's probable higher nonce")
	}

	return nil
}
