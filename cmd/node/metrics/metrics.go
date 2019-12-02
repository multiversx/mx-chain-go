package metrics

import (
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/constants"

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
	nodeType constants.NodeType,
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

	appStatusHandler.SetStringValue(constants.MetricPublicKeyBlockSign, factory.GetPkEncoded(pubKey))
	appStatusHandler.SetUInt64Value(constants.MetricShardId, shardId)
	appStatusHandler.SetStringValue(constants.MetricNodeType, string(nodeType))
	appStatusHandler.SetUInt64Value(constants.MetricRoundTime, roundDuration/millisecondsInSecond)
	appStatusHandler.SetStringValue(constants.MetricAppVersion, version)
	appStatusHandler.SetUInt64Value(constants.MetricCountConsensus, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricCountLeader, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricCountAcceptedBlocks, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricNumTxInBlock, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricNumMiniBlocks, initUint)
	appStatusHandler.SetStringValue(constants.MetricConsensusState, initString)
	appStatusHandler.SetStringValue(constants.MetricConsensusRoundState, initString)
	appStatusHandler.SetStringValue(constants.MetricCrossCheckBlockHeight, "0")
	appStatusHandler.SetUInt64Value(constants.MetricIsSyncing, isSyncing)
	appStatusHandler.SetStringValue(constants.MetricCurrentBlockHash, initString)
	appStatusHandler.SetUInt64Value(constants.MetricNumProcessedTxs, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricCurrentRoundTimestamp, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricHeaderSize, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricMiniBlocksSize, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricNumShardHeadersFromPool, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricNumShardHeadersProcessed, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricNumTimesInForkChoice, initUint)
	appStatusHandler.SetStringValue(constants.MetricPublicKeyTxSign, initString)
	appStatusHandler.SetUInt64Value(constants.MetricHighestFinalBlockInShard, initUint)
	appStatusHandler.SetUInt64Value(constants.MetricCountConsensusAcceptedBlocks, initUint)
	appStatusHandler.SetStringValue(constants.MetricRewardsValue, economicsConfig.RewardsSettings.RewardsValue)
	appStatusHandler.SetStringValue(constants.MetricLeaderPercentage, fmt.Sprintf("%f", economicsConfig.RewardsSettings.LeaderPercentage))
	appStatusHandler.SetStringValue(constants.MetricCommunityPercentage, fmt.Sprintf("%f", economicsConfig.RewardsSettings.CommunityPercentage))
	appStatusHandler.SetStringValue(constants.MetricDenominationCoefficient, economicsConfig.RewardsSettings.DenominationCoefficientForView)

	var consensusGroupSize uint32
	switch {
	case shardCoordinator.SelfId() < shardCoordinator.NumberOfShards():
		consensusGroupSize = nodesConfig.ConsensusGroupSize
	case shardCoordinator.SelfId() == constants.MetachainShardId:
		consensusGroupSize = nodesConfig.MetaChainConsensusGroupSize
	default:
		consensusGroupSize = 0
	}

	validatorsNodes, _ := nodesConfig.InitialNodesInfo()
	numValidators := len(validatorsNodes[shardCoordinator.SelfId()])

	appStatusHandler.SetUInt64Value(constants.MetricNumValidators, uint64(numValidators))
	appStatusHandler.SetUInt64Value(constants.MetricConsensusGroupSize, uint64(consensusGroupSize))
}

// SaveCurrentNodeNameAndPubKey will save metric in status handler with nodeName and transaction sign public key
func SaveCurrentNodeNameAndPubKey(ash core.AppStatusHandler, txSignPk string, nodeName string) {
	ash.SetStringValue(constants.MetricPublicKeyTxSign, txSignPk)
	ash.SetStringValue(constants.MetricNodeDisplayName, nodeName)
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
		appStatusHandler.SetUInt64Value(constants.MetricNumConnectedPeers, numOfConnectedPeers)
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
		appStatusHandler.SetUInt64Value(constants.MetricProbableHighestNonce, probableHigherNonce)
	}

	err := appStatusPollingHandler.RegisterPollingFunc(probableHighestNonceHandlerFunc)
	if err != nil {
		return errors.New("cannot register handler func for forkdetector's probable higher nonce")
	}

	return nil
}
