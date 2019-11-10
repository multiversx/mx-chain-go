package metrics

import (
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/appStatusPolling"
	"github.com/ElrondNetwork/elrond-go/crypto"
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
	economicsConfig *config.ConfigEconomics,
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
	appStatusHandler.SetStringValue(core.MetricPublicKeyTxSign, initString)
	appStatusHandler.SetUInt64Value(core.MetricHighestFinalBlockInShard, initUint)
	appStatusHandler.SetUInt64Value(core.MetricCountConsensusAcceptedBlocks, initUint)
	appStatusHandler.SetStringValue(core.MetricRewardsValue, economicsConfig.RewardsSettings.RewardsValue)
	appStatusHandler.SetStringValue(core.MetricLeaderPercentage, fmt.Sprintf("%f", economicsConfig.RewardsSettings.LeaderPercentage))
	appStatusHandler.SetStringValue(core.MetricCommunityPercentage, fmt.Sprintf("%f", economicsConfig.RewardsSettings.CommunityPercentage))
	appStatusHandler.SetStringValue(core.MetricDenominationCoefficient, economicsConfig.RewardsSettings.DenominationCoefficientForView)

	var consensusGroupSize uint32
	switch {
	case shardCoordinator.SelfId() < shardCoordinator.NumberOfShards():
		consensusGroupSize = nodesConfig.ConsensusGroupSize
	case shardCoordinator.SelfId() == sharding.MetachainShardId:
		consensusGroupSize = nodesConfig.MetaChainConsensusGroupSize
	default:
		consensusGroupSize = 0
	}

	validatorsNodes := nodesConfig.InitialNodesInfo()
	numValidators := len(validatorsNodes[shardCoordinator.SelfId()])

	appStatusHandler.SetUInt64Value(core.MetricNumValidators, uint64(numValidators))
	appStatusHandler.SetUInt64Value(core.MetricConsensusGroupSize, uint64(consensusGroupSize))
}

// SaveCurrentNodeNameAndPubKey will save metric in status handler with nodeName and transaction sign public key
func SaveCurrentNodeNameAndPubKey(ash core.AppStatusHandler, txSignPk string, nodeName string) {
	ash.SetStringValue(core.MetricPublicKeyTxSign, txSignPk)
	ash.SetStringValue(core.MetricNodeDisplayName, nodeName)
}

// StartStatusPolling will start save information in status handler about network
func StartStatusPolling(
	ash core.AppStatusHandler,
	pollingInterval int,
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

	numOfConnectedPeersHandlerFunc := func(appStatusHandler core.AppStatusHandler) {
		numOfConnectedPeers := uint64(len(networkComponents.NetMessenger.ConnectedAddresses()))
		appStatusHandler.SetUInt64Value(core.MetricNumConnectedPeers, numOfConnectedPeers)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(numOfConnectedPeersHandlerFunc)
	if err != nil {
		return errors.New("cannot register handler func for num of connected peers")
	}

	return nil
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
