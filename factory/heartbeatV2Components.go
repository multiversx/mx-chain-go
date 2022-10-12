package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/random"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/heartbeat/monitor"
	"github.com/ElrondNetwork/elrond-go/heartbeat/processor"
	"github.com/ElrondNetwork/elrond-go/heartbeat/sender"
	"github.com/ElrondNetwork/elrond-go/heartbeat/status"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/update"
)

// ArgHeartbeatV2ComponentsFactory represents the argument for the heartbeat v2 components factory
type ArgHeartbeatV2ComponentsFactory struct {
	Config                  config.Config
	Prefs                   config.Preferences
	AppVersion              string
	BoostrapComponents      BootstrapComponentsHolder
	CoreComponents          CoreComponentsHolder
	DataComponents          DataComponentsHolder
	NetworkComponents       NetworkComponentsHolder
	CryptoComponents        CryptoComponentsHolder
	ProcessComponents       ProcessComponentsHolder
	HeartbeatV1DisableEpoch uint32
}

type heartbeatV2ComponentsFactory struct {
	config                  config.Config
	prefs                   config.Preferences
	version                 string
	boostrapComponents      BootstrapComponentsHolder
	coreComponents          CoreComponentsHolder
	dataComponents          DataComponentsHolder
	networkComponents       NetworkComponentsHolder
	cryptoComponents        CryptoComponentsHolder
	processComponents       ProcessComponentsHolder
	heartbeatV1DisableEpoch uint32
}

type heartbeatV2Components struct {
	sender                    update.Closer
	peerAuthRequestsProcessor update.Closer
	shardSender               update.Closer
	monitor                   HeartbeatV2Monitor
	statusHandler             update.Closer
}

// NewHeartbeatV2ComponentsFactory creates a new instance of heartbeatV2ComponentsFactory
func NewHeartbeatV2ComponentsFactory(args ArgHeartbeatV2ComponentsFactory) (*heartbeatV2ComponentsFactory, error) {
	err := checkHeartbeatV2FactoryArgs(args)
	if err != nil {
		return nil, err
	}

	return &heartbeatV2ComponentsFactory{
		config:                  args.Config,
		prefs:                   args.Prefs,
		version:                 args.AppVersion,
		boostrapComponents:      args.BoostrapComponents,
		coreComponents:          args.CoreComponents,
		dataComponents:          args.DataComponents,
		networkComponents:       args.NetworkComponents,
		cryptoComponents:        args.CryptoComponents,
		processComponents:       args.ProcessComponents,
		heartbeatV1DisableEpoch: args.HeartbeatV1DisableEpoch,
	}, nil
}

func checkHeartbeatV2FactoryArgs(args ArgHeartbeatV2ComponentsFactory) error {
	if check.IfNil(args.BoostrapComponents) {
		return errors.ErrNilBootstrapComponentsHolder
	}
	if check.IfNil(args.CoreComponents) {
		return errors.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.DataComponents) {
		return errors.ErrNilDataComponentsHolder
	}
	if check.IfNil(args.NetworkComponents) {
		return errors.ErrNilNetworkComponentsHolder
	}
	if check.IfNil(args.CryptoComponents) {
		return errors.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(args.ProcessComponents) {
		return errors.ErrNilProcessComponentsHolder
	}
	hardforkTrigger := args.ProcessComponents.HardforkTrigger()
	if check.IfNil(hardforkTrigger) {
		return errors.ErrNilHardforkTrigger
	}

	return nil
}

// Create creates the heartbeatV2 components
func (hcf *heartbeatV2ComponentsFactory) Create() (*heartbeatV2Components, error) {
	if !hcf.networkComponents.NetworkMessenger().HasTopic(common.PeerAuthenticationTopic) {
		err := hcf.networkComponents.NetworkMessenger().CreateTopic(common.PeerAuthenticationTopic, true)
		if err != nil {
			return nil, err
		}
	}
	if !hcf.networkComponents.NetworkMessenger().HasTopic(common.HeartbeatV2Topic) {
		err := hcf.networkComponents.NetworkMessenger().CreateTopic(common.HeartbeatV2Topic, true)
		if err != nil {
			return nil, err
		}
	}

	peerSubType := core.RegularPeer
	if hcf.prefs.Preferences.FullArchive {
		peerSubType = core.FullHistoryObserver
	}

	shardC := hcf.boostrapComponents.ShardCoordinator()
	heartbeatTopic := common.HeartbeatV2Topic + shardC.CommunicationIdentifier(shardC.SelfId())

	cfg := hcf.config.HeartbeatV2

	argPeerTypeProvider := peer.ArgPeerTypeProvider{
		NodesCoordinator:        hcf.processComponents.NodesCoordinator(),
		StartEpoch:              hcf.processComponents.EpochStartTrigger().MetaEpoch(),
		EpochStartEventNotifier: hcf.processComponents.EpochStartNotifier(),
	}
	peerTypeProvider, err := peer.NewPeerTypeProvider(argPeerTypeProvider)
	if err != nil {
		return nil, err
	}

	argsSender := sender.ArgSender{
		Messenger:                          hcf.networkComponents.NetworkMessenger(),
		Marshaller:                         hcf.coreComponents.InternalMarshalizer(),
		PeerAuthenticationTopic:            common.PeerAuthenticationTopic,
		HeartbeatTopic:                     heartbeatTopic,
		PeerAuthenticationTimeBetweenSends: time.Second * time.Duration(cfg.PeerAuthenticationTimeBetweenSendsInSec),
		PeerAuthenticationTimeBetweenSendsWhenError: time.Second * time.Duration(cfg.PeerAuthenticationTimeBetweenSendsWhenErrorInSec),
		PeerAuthenticationThresholdBetweenSends:     cfg.PeerAuthenticationThresholdBetweenSends,
		HeartbeatTimeBetweenSends:                   time.Second * time.Duration(cfg.HeartbeatTimeBetweenSendsInSec),
		HeartbeatTimeBetweenSendsWhenError:          time.Second * time.Duration(cfg.HeartbeatTimeBetweenSendsWhenErrorInSec),
		HeartbeatThresholdBetweenSends:              cfg.HeartbeatThresholdBetweenSends,
		VersionNumber:                               hcf.version,
		NodeDisplayName:                             hcf.prefs.Preferences.NodeDisplayName,
		Identity:                                    hcf.prefs.Preferences.Identity,
		PeerSubType:                                 peerSubType,
		CurrentBlockProvider:                        hcf.dataComponents.Blockchain(),
		PeerSignatureHandler:                        hcf.cryptoComponents.PeerSignatureHandler(),
		PrivateKey:                                  hcf.cryptoComponents.PrivateKey(),
		RedundancyHandler:                           hcf.processComponents.NodeRedundancyHandler(),
		NodesCoordinator:                            hcf.processComponents.NodesCoordinator(),
		HardforkTrigger:                             hcf.processComponents.HardforkTrigger(),
		HardforkTimeBetweenSends:                    time.Second * time.Duration(cfg.HardforkTimeBetweenSendsInSec),
		HardforkTriggerPubKey:                       hcf.coreComponents.HardforkTriggerPubKey(),
		PeerTypeProvider:                            peerTypeProvider,
	}
	heartbeatV2Sender, err := sender.NewSender(argsSender)
	if err != nil {
		return nil, err
	}

	epochBootstrapParams := hcf.boostrapComponents.EpochBootstrapParams()
	argsProcessor := processor.ArgPeerAuthenticationRequestsProcessor{
		RequestHandler:          hcf.processComponents.RequestHandler(),
		NodesCoordinator:        hcf.processComponents.NodesCoordinator(),
		PeerAuthenticationPool:  hcf.dataComponents.Datapool().PeerAuthentications(),
		ShardId:                 epochBootstrapParams.SelfShardID(),
		Epoch:                   epochBootstrapParams.Epoch(),
		MinPeersThreshold:       cfg.MinPeersThreshold,
		DelayBetweenRequests:    time.Second * time.Duration(cfg.DelayBetweenRequestsInSec),
		MaxTimeout:              time.Second * time.Duration(cfg.MaxTimeoutInSec),
		MaxMissingKeysInRequest: cfg.MaxMissingKeysInRequest,
		Randomizer:              &random.ConcurrentSafeIntRandomizer{},
	}
	paRequestsProcessor, err := processor.NewPeerAuthenticationRequestsProcessor(argsProcessor)
	if err != nil {
		return nil, err
	}

	argsPeerShardSender := sender.ArgPeerShardSender{
		Messenger:             hcf.networkComponents.NetworkMessenger(),
		Marshaller:            hcf.coreComponents.InternalMarshalizer(),
		ShardCoordinator:      hcf.boostrapComponents.ShardCoordinator(),
		TimeBetweenSends:      time.Second * time.Duration(cfg.PeerShardTimeBetweenSendsInSec),
		ThresholdBetweenSends: cfg.PeerShardThresholdBetweenSends,
		NodesCoordinator:      hcf.processComponents.NodesCoordinator(),
	}
	shardSender, err := sender.NewPeerShardSender(argsPeerShardSender)
	if err != nil {
		return nil, err
	}

	argsMonitor := monitor.ArgHeartbeatV2Monitor{
		Cache:                         hcf.dataComponents.Datapool().Heartbeats(),
		PubKeyConverter:               hcf.coreComponents.ValidatorPubKeyConverter(),
		Marshaller:                    hcf.coreComponents.InternalMarshalizer(),
		PeerShardMapper:               hcf.processComponents.PeerShardMapper(),
		MaxDurationPeerUnresponsive:   time.Second * time.Duration(cfg.MaxDurationPeerUnresponsiveInSec),
		HideInactiveValidatorInterval: time.Second * time.Duration(cfg.HideInactiveValidatorIntervalInSec),
		ShardId:                       epochBootstrapParams.SelfShardID(),
		PeerTypeProvider:              peerTypeProvider,
	}
	heartbeatsMonitor, err := monitor.NewHeartbeatV2Monitor(argsMonitor)
	if err != nil {
		return nil, err
	}

	argsMetricsUpdater := status.ArgsMetricsUpdater{
		PeerAuthenticationCacher:            hcf.dataComponents.Datapool().PeerAuthentications(),
		HeartbeatMonitor:                    heartbeatsMonitor,
		HeartbeatSenderInfoProvider:         heartbeatV2Sender,
		AppStatusHandler:                    hcf.coreComponents.StatusHandler(),
		TimeBetweenConnectionsMetricsUpdate: time.Second * time.Duration(hcf.config.HeartbeatV2.TimeBetweenConnectionsMetricsUpdateInSec),
		EpochNotifier:                       hcf.coreComponents.EpochNotifier(),
		HeartbeatV1DisableEpoch:             hcf.heartbeatV1DisableEpoch,
	}
	statusHandler, err := status.NewMetricsUpdater(argsMetricsUpdater)
	if err != nil {
		return nil, err
	}

	argsCrossShardPeerTopicNotifier := monitor.ArgsCrossShardPeerTopicNotifier{
		ShardCoordinator: hcf.processComponents.ShardCoordinator(),
		PeerShardMapper:  hcf.processComponents.PeerShardMapper(),
	}
	crossShardPeerTopicNotifier, err := monitor.NewCrossShardPeerTopicNotifier(argsCrossShardPeerTopicNotifier)
	if err != nil {
		return nil, err
	}
	err = hcf.networkComponents.NetworkMessenger().AddPeerTopicNotifier(crossShardPeerTopicNotifier)
	if err != nil {
		return nil, err
	}

	return &heartbeatV2Components{
		sender:                    heartbeatV2Sender,
		peerAuthRequestsProcessor: paRequestsProcessor,
		shardSender:               shardSender,
		monitor:                   heartbeatsMonitor,
		statusHandler:             statusHandler,
	}, nil
}

// Close closes the heartbeat components
func (hc *heartbeatV2Components) Close() error {
	log.Debug("calling close on heartbeatV2 components")

	if !check.IfNil(hc.sender) {
		log.LogIfError(hc.sender.Close())
	}

	if !check.IfNil(hc.peerAuthRequestsProcessor) {
		log.LogIfError(hc.peerAuthRequestsProcessor.Close())
	}

	if !check.IfNil(hc.shardSender) {
		log.LogIfError(hc.shardSender.Close())
	}

	if !check.IfNil(hc.statusHandler) {
		log.LogIfError(hc.statusHandler.Close())
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hcf *heartbeatV2ComponentsFactory) IsInterfaceNil() bool {
	return hcf == nil
}
