package metrics

import (
	"errors"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/appStatusPolling"
	"github.com/ElrondNetwork/elrond-go/core/check"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const millisecondsInSecond = 1000

// InitMetrics will init metrics for status handler
func InitMetrics(
	appStatusHandler core.AppStatusHandler,
	pubkeyStr string,
	nodeType core.NodeType,
	shardCoordinator sharding.Coordinator,
	nodesConfig *sharding.NodesSetup,
	version string,
	economicsConfig *config.EconomicsConfig,
	roundsPerEpoch int64,
) error {
	if check.IfNil(appStatusHandler) {
		return fmt.Errorf("nil AppStatusHandler when initializing metrics")
	}
	if check.IfNil(shardCoordinator) {
		return fmt.Errorf("nil shard coordinator when initializing metrics")
	}
	if nodesConfig == nil {
		return fmt.Errorf("nil nodes config when initializing metrics")
	}
	if economicsConfig == nil {
		return fmt.Errorf("nil economics config when initializing metrics")
	}

	shardId := uint64(shardCoordinator.SelfId())
	numOfShards := uint64(shardCoordinator.NumberOfShards())
	roundDuration := nodesConfig.RoundDuration
	isSyncing := uint64(1)
	initUint := uint64(0)
	initString := ""

	appStatusHandler.SetStringValue(core.MetricPublicKeyBlockSign, pubkeyStr)
	appStatusHandler.SetUInt64Value(core.MetricShardId, shardId)
	appStatusHandler.SetUInt64Value(core.MetricNumShardsWithoutMetacahin, numOfShards)
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
	appStatusHandler.SetUInt64Value(core.MetricNonceAtEpochStart, initUint)
	appStatusHandler.SetUInt64Value(core.MetricRoundsPassedInCurrentEpoch, initUint)
	appStatusHandler.SetUInt64Value(core.MetricNoncesPassedInCurrentEpoch, initUint)
	appStatusHandler.SetStringValue(core.MetricLeaderPercentage, fmt.Sprintf("%f", economicsConfig.RewardsSettings.LeaderPercentage))
	appStatusHandler.SetStringValue(core.MetricDenominationCoefficient, economicsConfig.RewardsSettings.DenominationCoefficientForView)
	appStatusHandler.SetUInt64Value(core.MetricNumConnectedPeers, initUint)
	appStatusHandler.SetStringValue(core.MetricNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(core.MetricLatestTagSoftwareVersion, initString)

	appStatusHandler.SetStringValue(core.MetricP2PNumConnectedPeersClassification, initString)
	appStatusHandler.SetStringValue(core.MetricP2PPeerInfo, initString)
	appStatusHandler.SetStringValue(core.MetricP2PIntraShardValidators, initString)
	appStatusHandler.SetStringValue(core.MetricP2PIntraShardObservers, initString)
	appStatusHandler.SetStringValue(core.MetricP2PCrossShardValidators, initString)
	appStatusHandler.SetStringValue(core.MetricP2PCrossShardObservers, initString)
	appStatusHandler.SetStringValue(core.MetricP2PUnknownPeers, initString)
	appStatusHandler.SetUInt64Value(core.MetricShardConsensusGroupSize, uint64(nodesConfig.ConsensusGroupSize))
	appStatusHandler.SetUInt64Value(core.MetricMetaConsensusGroupSize, uint64(nodesConfig.MetaChainConsensusGroupSize))
	appStatusHandler.SetUInt64Value(core.MetricNumNodesPerShard, uint64(nodesConfig.MinNodesPerShard))
	appStatusHandler.SetUInt64Value(core.MetricNumMetachainNodes, uint64(nodesConfig.MetaChainMinNodes))
	appStatusHandler.SetUInt64Value(core.MetricStartTime, uint64(nodesConfig.StartTime))
	appStatusHandler.SetUInt64Value(core.MetricRoundDuration, nodesConfig.RoundDuration)

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

	return nil
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
	networkComponents mainFactory.NetworkComponentsHolder,
	processComponents mainFactory.ProcessComponentsHolder,
	shardCoordinator sharding.Coordinator,
) error {
	if ash == nil {
		return errors.New("nil AppStatusHandler")
	}
	if check.IfNil(networkComponents) {
		return errors.New("nil networkComponents")
	}
	if processComponents == nil {
		return errors.New("nil processComponents")
	}
	if check.IfNil(shardCoordinator) {
		return errors.New("nil shard coordinator")
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

	err = registerShardsInformation(appStatusPollingHandler, shardCoordinator)
	if err != nil {
		return err
	}

	appStatusPollingHandler.Poll()

	return nil
}

func registerPollConnectedPeers(
	appStatusPollingHandler *appStatusPolling.AppStatusPolling,
	networkComponents mainFactory.NetworkComponentsHolder,
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

func registerShardsInformation(
	appStatusPollingHandler *appStatusPolling.AppStatusPolling,
	coordinator sharding.Coordinator,
) error {

	computeShardsInfo := func(appStatusHandler core.AppStatusHandler) {
		shardId := uint64(coordinator.SelfId())
		numOfShards := uint64(coordinator.NumberOfShards())

		appStatusHandler.SetUInt64Value(core.MetricShardId, shardId)
		appStatusHandler.SetUInt64Value(core.MetricNumShardsWithoutMetacahin, numOfShards)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(computeShardsInfo)
	if err != nil {
		return fmt.Errorf("%w, cannot register handler func for shards information", err)
	}

	return nil
}

func computeNumConnectedPeers(
	appStatusHandler core.AppStatusHandler,
	networkComponents mainFactory.NetworkComponentsHolder,
) {
	numOfConnectedPeers := uint64(len(networkComponents.NetworkMessenger().ConnectedAddresses()))
	appStatusHandler.SetUInt64Value(core.MetricNumConnectedPeers, numOfConnectedPeers)
}

func computeConnectedPeers(
	appStatusHandler core.AppStatusHandler,
	networkComponents mainFactory.NetworkComponentsHolder,
) {
	peersInfo := networkComponents.NetworkMessenger().GetConnectedPeersInfo()

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
	networkComponents mainFactory.NetworkComponentsHolder,
) {
	appStatusHandler.SetStringValue(core.MetricP2PPeerInfo, sliceToString(networkComponents.NetworkMessenger().Addresses()))
}

func registerPollProbableHighestNonce(
	appStatusPollingHandler *appStatusPolling.AppStatusPolling,
	processComponents mainFactory.ProcessComponentsHolder,
) error {

	probableHighestNonceHandlerFunc := func(appStatusHandler core.AppStatusHandler) {
		probableHigherNonce := processComponents.ForkDetector().ProbableHighestNonce()
		appStatusHandler.SetUInt64Value(core.MetricProbableHighestNonce, probableHigherNonce)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(probableHighestNonceHandlerFunc)
	if err != nil {
		return errors.New("cannot register handler func for forkdetector's probable higher nonce")
	}

	return nil
}
