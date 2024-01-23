package node

import (
	"errors"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/node/nodeDebugFactory"
	"github.com/multiversx/mx-chain-go/p2p"
	procFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/blackList"
	"github.com/multiversx/mx-chain-go/sharding"
)

// prepareOpenTopics will set to the anti flood handler the topics for which
// the node can receive messages from others than validators
func prepareOpenTopics(
	antiflood factory.P2PAntifloodHandler,
	shardCoordinator sharding.Coordinator,
) {
	selfID := shardCoordinator.SelfId()
	selfShardHeartbeatV2Topic := common.HeartbeatV2Topic + core.CommunicationIdentifierBetweenShards(selfID, selfID)
	if selfID == core.MetachainShardId {
		antiflood.SetTopicsForAll(
			common.PeerAuthenticationTopic,
			selfShardHeartbeatV2Topic,
			common.ConnectionTopic)
		return
	}

	selfShardTxTopic := procFactory.TransactionTopic + core.CommunicationIdentifierBetweenShards(selfID, selfID)
	antiflood.SetTopicsForAll(
		common.PeerAuthenticationTopic,
		selfShardHeartbeatV2Topic,
		common.ConnectionTopic,
		selfShardTxTopic)
}

// CreateNode is the node factory
func CreateNode(
	config *config.Config,
	statusCoreComponents factory.StatusCoreComponentsHandler,
	bootstrapComponents factory.BootstrapComponentsHandler,
	coreComponents factory.CoreComponentsHandler,
	cryptoComponents factory.CryptoComponentsHandler,
	dataComponents factory.DataComponentsHandler,
	networkComponents factory.NetworkComponentsHandler,
	processComponents factory.ProcessComponentsHandler,
	stateComponents factory.StateComponentsHandler,
	statusComponents factory.StatusComponentsHandler,
	heartbeatV2Components factory.HeartbeatV2ComponentsHandler,
	consensusComponents factory.ConsensusComponentsHandler,
	bootstrapRoundIndex uint64,
	isInImportMode bool,
	extraOptions ...Option,
) (*Node, error) {
	prepareOpenTopics(networkComponents.InputAntiFloodHandler(), processComponents.ShardCoordinator())

	peerDenialEvaluator, err := createAndAttachPeerDenialEvaluators(networkComponents, processComponents)
	if err != nil {
		return nil, err
	}

	genesisTime := time.Unix(coreComponents.GenesisNodesSetup().GetStartTime(), 0)

	consensusGroupSize, err := consensusComponents.ConsensusGroupSize()
	if err != nil {
		return nil, err
	}

	var nd *Node
	options := []Option{
		WithStatusCoreComponents(statusCoreComponents),
		WithCoreComponents(coreComponents),
		WithCryptoComponents(cryptoComponents),
		WithBootstrapComponents(bootstrapComponents),
		WithStateComponents(stateComponents),
		WithDataComponents(dataComponents),
		WithStatusComponents(statusComponents),
		WithProcessComponents(processComponents),
		WithHeartbeatV2Components(heartbeatV2Components),
		WithConsensusComponents(consensusComponents),
		WithNetworkComponents(networkComponents),
		WithInitialNodesPubKeys(coreComponents.GenesisNodesSetup().InitialNodesPubKeys()),
		WithRoundDuration(coreComponents.GenesisNodesSetup().GetRoundDuration()),
		WithConsensusGroupSize(consensusGroupSize),
		WithGenesisTime(genesisTime),
		WithConsensusType(config.Consensus.Type),
		WithBootstrapRoundIndex(bootstrapRoundIndex),
		WithPeerDenialEvaluator(peerDenialEvaluator),
		WithRequestedItemsHandler(processComponents.RequestedItemsHandler()),
		WithAddressSignatureSize(config.AddressPubkeyConverter.SignatureLength),
		WithValidatorSignatureSize(config.ValidatorPubkeyConverter.SignatureLength),
		WithPublicKeySize(config.ValidatorPubkeyConverter.Length),
		WithNodeStopChannel(coreComponents.ChanStopNodeProcess()),
		WithImportMode(isInImportMode),
		WithESDTNFTStorageHandler(processComponents.ESDTDataStorageHandlerForAPI()),
	}
	options = append(options, extraOptions...)

	nd, err = NewNode(options...)
	if err != nil {
		return nil, errors.New("error creating node: " + err.Error())
	}

	if processComponents.ShardCoordinator().SelfId() < processComponents.ShardCoordinator().NumberOfShards() {
		err = nd.CreateShardedStores()
		if err != nil {
			return nil, err
		}
	}

	err = nodeDebugFactory.CreateInterceptedDebugHandler(
		nd,
		processComponents.InterceptorsContainer(),
		processComponents.ResolversContainer(),
		processComponents.RequestersFinder(),
		config.Debug.InterceptorResolver,
	)
	if err != nil {
		return nil, err
	}

	return nd, nil
}

func createAndAttachPeerDenialEvaluators(
	networkComponents factory.NetworkComponentsHandler,
	processComponents factory.ProcessComponentsHandler,
) (p2p.PeerDenialEvaluator, error) {
	mainPeerDenialEvaluator, err := blackList.NewPeerDenialEvaluator(
		networkComponents.PeerBlackListHandler(),
		networkComponents.PubKeyCacher(),
		processComponents.PeerShardMapper(),
	)
	if err != nil {
		return nil, err
	}

	err = networkComponents.NetworkMessenger().SetPeerDenialEvaluator(mainPeerDenialEvaluator)
	if err != nil {
		return nil, err
	}

	fullArchivePeerDenialEvaluator, err := blackList.NewPeerDenialEvaluator(
		networkComponents.PeerBlackListHandler(),
		networkComponents.PubKeyCacher(),
		processComponents.FullArchivePeerShardMapper(),
	)
	if err != nil {
		return nil, err
	}

	err = networkComponents.FullArchiveNetworkMessenger().SetPeerDenialEvaluator(fullArchivePeerDenialEvaluator)
	if err != nil {
		return nil, err
	}

	return mainPeerDenialEvaluator, nil
}
