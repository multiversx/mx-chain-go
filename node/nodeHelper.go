package node

import (
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	nodeDisabled "github.com/ElrondNetwork/elrond-go/node/disabled"
	"github.com/ElrondNetwork/elrond-go/node/nodeDebugFactory"
	procFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/blackList"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
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
		antiflood.SetTopicsForAll(common.PeerAuthenticationTopic, selfShardHeartbeatV2Topic, common.ConnectionTopic)
		return
	}

	selfShardTxTopic := procFactory.TransactionTopic + core.CommunicationIdentifierBetweenShards(selfID, selfID)
	antiflood.SetTopicsForAll(common.PeerAuthenticationTopic, selfShardHeartbeatV2Topic, common.ConnectionTopic, selfShardTxTopic)
}

// CreateNode is the node factory
func CreateNode(
	config *config.Config,
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
) (*Node, error) {
	prepareOpenTopics(networkComponents.InputAntiFloodHandler(), processComponents.ShardCoordinator())

	peerDenialEvaluator, err := blackList.NewPeerDenialEvaluator(
		networkComponents.PeerBlackListHandler(),
		networkComponents.PubKeyCacher(),
		processComponents.PeerShardMapper(),
	)
	if err != nil {
		return nil, err
	}

	err = networkComponents.NetworkMessenger().SetPeerDenialEvaluator(peerDenialEvaluator)
	if err != nil {
		return nil, err
	}

	genesisTime := time.Unix(coreComponents.GenesisNodesSetup().GetStartTime(), 0)

	consensusGroupSize, err := consensusComponents.ConsensusGroupSize()
	if err != nil {
		return nil, err
	}

	esdtNftStorage, err := builtInFunctions.NewESDTDataStorage(builtInFunctions.ArgsNewESDTDataStorage{
		Accounts:              stateComponents.AccountsAdapterAPI(),
		GlobalSettingsHandler: nodeDisabled.NewDisabledGlobalSettingHandler(),
		Marshalizer:           coreComponents.InternalMarshalizer(),
		EnableEpochsHandler:   coreComponents.EnableEpochsHandler(),
		ShardCoordinator:      processComponents.ShardCoordinator(),
	})
	if err != nil {
		return nil, err
	}

	var nd *Node
	nd, err = NewNode(
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
		WithESDTNFTStorageHandler(esdtNftStorage),
	)
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
		processComponents.ResolversFinder(),
		config.Debug.InterceptorResolver,
	)
	if err != nil {
		return nil, err
	}

	return nd, nil
}
