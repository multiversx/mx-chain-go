package heartbeat

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/random"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/heartbeat/monitor"
	"github.com/multiversx/mx-chain-go/heartbeat/processor"
	"github.com/multiversx/mx-chain-go/heartbeat/sender"
	"github.com/multiversx/mx-chain-go/heartbeat/status"
	"github.com/multiversx/mx-chain-go/p2p"
	processFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/peer"
	"github.com/multiversx/mx-chain-go/update"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("factory")

// ArgHeartbeatV2ComponentsFactory represents the argument for the heartbeat v2 components factory
type ArgHeartbeatV2ComponentsFactory struct {
	Config               config.Config
	Prefs                config.Preferences
	BaseVersion          string
	AppVersion           string
	BootstrapComponents  factory.BootstrapComponentsHolder
	CoreComponents       factory.CoreComponentsHolder
	DataComponents       factory.DataComponentsHolder
	NetworkComponents    factory.NetworkComponentsHolder
	CryptoComponents     factory.CryptoComponentsHolder
	ProcessComponents    factory.ProcessComponentsHolder
	StatusCoreComponents factory.StatusCoreComponentsHolder
}

type heartbeatV2ComponentsFactory struct {
	config               config.Config
	prefs                config.Preferences
	baseVersion          string
	version              string
	bootstrapComponents  factory.BootstrapComponentsHolder
	coreComponents       factory.CoreComponentsHolder
	dataComponents       factory.DataComponentsHolder
	networkComponents    factory.NetworkComponentsHolder
	cryptoComponents     factory.CryptoComponentsHolder
	processComponents    factory.ProcessComponentsHolder
	statusCoreComponents factory.StatusCoreComponentsHolder
}

type heartbeatV2Components struct {
	sender                               update.Closer
	peerAuthRequestsProcessor            update.Closer
	shardSender                          update.Closer
	monitor                              factory.HeartbeatV2Monitor
	statusHandler                        update.Closer
	mainDirectConnectionProcessor        update.Closer
	fullArchiveDirectConnectionProcessor update.Closer
}

// NewHeartbeatV2ComponentsFactory creates a new instance of heartbeatV2ComponentsFactory
func NewHeartbeatV2ComponentsFactory(args ArgHeartbeatV2ComponentsFactory) (*heartbeatV2ComponentsFactory, error) {
	err := checkHeartbeatV2FactoryArgs(args)
	if err != nil {
		return nil, err
	}

	return &heartbeatV2ComponentsFactory{
		config:               args.Config,
		prefs:                args.Prefs,
		baseVersion:          args.BaseVersion,
		version:              args.AppVersion,
		bootstrapComponents:  args.BootstrapComponents,
		coreComponents:       args.CoreComponents,
		dataComponents:       args.DataComponents,
		networkComponents:    args.NetworkComponents,
		cryptoComponents:     args.CryptoComponents,
		processComponents:    args.ProcessComponents,
		statusCoreComponents: args.StatusCoreComponents,
	}, nil
}

func checkHeartbeatV2FactoryArgs(args ArgHeartbeatV2ComponentsFactory) error {
	if check.IfNil(args.BootstrapComponents) {
		return errors.ErrNilBootstrapComponentsHolder
	}
	if check.IfNil(args.CoreComponents) {
		return errors.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.DataComponents) {
		return errors.ErrNilDataComponentsHolder
	}
	if check.IfNil(args.DataComponents.Datapool()) {
		return errors.ErrNilDataPoolsHolder
	}
	if check.IfNil(args.NetworkComponents) {
		return errors.ErrNilNetworkComponentsHolder
	}
	if check.IfNil(args.NetworkComponents.NetworkMessenger()) {
		return errors.ErrNilMessenger
	}
	if check.IfNil(args.CryptoComponents) {
		return errors.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(args.ProcessComponents) {
		return errors.ErrNilProcessComponentsHolder
	}
	if check.IfNil(args.ProcessComponents.EpochStartTrigger()) {
		return errors.ErrNilEpochStartTrigger
	}
	if check.IfNil(args.StatusCoreComponents) {
		return errors.ErrNilStatusCoreComponents
	}

	return nil
}

// Create creates the heartbeatV2 components
func (hcf *heartbeatV2ComponentsFactory) Create() (*heartbeatV2Components, error) {
	err := hcf.createTopicsIfNeeded()
	if err != nil {
		return nil, err
	}

	cfg := hcf.config.HeartbeatV2
	if cfg.HeartbeatExpiryTimespanInSec <= cfg.PeerAuthenticationTimeBetweenSendsInSec {
		return nil, fmt.Errorf("%w, HeartbeatExpiryTimespanInSec must be greater than PeerAuthenticationTimeBetweenSendsInSec", errors.ErrInvalidHeartbeatV2Config)
	}

	peerSubType := core.RegularPeer
	if hcf.prefs.Preferences.FullArchive {
		peerSubType = core.FullHistoryObserver
	}

	shardC := hcf.bootstrapComponents.ShardCoordinator()
	heartbeatTopic := common.HeartbeatV2Topic + shardC.CommunicationIdentifier(shardC.SelfId())

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
		MainMessenger:                               hcf.networkComponents.NetworkMessenger(),
		FullArchiveMessenger:                        hcf.networkComponents.FullArchiveNetworkMessenger(),
		Marshaller:                                  hcf.coreComponents.InternalMarshalizer(),
		PeerAuthenticationTopic:                     common.PeerAuthenticationTopic,
		HeartbeatTopic:                              heartbeatTopic,
		PeerAuthenticationTimeBetweenSends:          time.Second * time.Duration(cfg.PeerAuthenticationTimeBetweenSendsInSec),
		PeerAuthenticationTimeBetweenSendsWhenError: time.Second * time.Duration(cfg.PeerAuthenticationTimeBetweenSendsWhenErrorInSec),
		PeerAuthenticationTimeThresholdBetweenSends: cfg.PeerAuthenticationTimeThresholdBetweenSends,
		HeartbeatTimeBetweenSends:                   time.Second * time.Duration(cfg.HeartbeatTimeBetweenSendsInSec),
		HeartbeatTimeBetweenSendsWhenError:          time.Second * time.Duration(cfg.HeartbeatTimeBetweenSendsWhenErrorInSec),
		HeartbeatTimeThresholdBetweenSends:          cfg.HeartbeatTimeThresholdBetweenSends,
		BaseVersionNumber:                           hcf.baseVersion,
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
		ManagedPeersHolder:                          hcf.cryptoComponents.ManagedPeersHolder(),
		PeerAuthenticationTimeBetweenChecks:         time.Second * time.Duration(cfg.PeerAuthenticationTimeBetweenChecksInSec),
		ShardCoordinator:                            hcf.processComponents.ShardCoordinator(),
	}
	heartbeatV2Sender, err := sender.NewSender(argsSender)
	if err != nil {
		return nil, err
	}

	epochBootstrapParams := hcf.bootstrapComponents.EpochBootstrapParams()
	argsProcessor := processor.ArgPeerAuthenticationRequestsProcessor{
		RequestHandler:          hcf.processComponents.RequestHandler(),
		NodesCoordinator:        hcf.processComponents.NodesCoordinator(),
		PeerAuthenticationPool:  hcf.dataComponents.Datapool().PeerAuthentications(),
		ShardId:                 epochBootstrapParams.SelfShardID(),
		Epoch:                   epochBootstrapParams.Epoch(),
		MinPeersThreshold:       cfg.MinPeersThreshold,
		DelayBetweenRequests:    time.Second * time.Duration(cfg.DelayBetweenPeerAuthenticationRequestsInSec),
		MaxTimeoutForRequests:   time.Second * time.Duration(cfg.PeerAuthenticationMaxTimeoutForRequestsInSec),
		MaxMissingKeysInRequest: cfg.MaxMissingKeysInRequest,
		Randomizer:              &random.ConcurrentSafeIntRandomizer{},
	}
	paRequestsProcessor, err := processor.NewPeerAuthenticationRequestsProcessor(argsProcessor)
	if err != nil {
		return nil, err
	}

	argsPeerShardSender := sender.ArgPeerShardSender{
		MainMessenger:             hcf.networkComponents.NetworkMessenger(),
		FullArchiveMessenger:      hcf.networkComponents.FullArchiveNetworkMessenger(),
		Marshaller:                hcf.coreComponents.InternalMarshalizer(),
		ShardCoordinator:          hcf.bootstrapComponents.ShardCoordinator(),
		TimeBetweenSends:          time.Second * time.Duration(cfg.PeerShardTimeBetweenSendsInSec),
		TimeThresholdBetweenSends: cfg.PeerShardTimeThresholdBetweenSends,
		NodesCoordinator:          hcf.processComponents.NodesCoordinator(),
	}
	shardSender, err := sender.NewPeerShardSender(argsPeerShardSender)
	if err != nil {
		return nil, err
	}

	argsMonitor := monitor.ArgHeartbeatV2Monitor{
		Cache:                         hcf.dataComponents.Datapool().Heartbeats(),
		PubKeyConverter:               hcf.coreComponents.ValidatorPubKeyConverter(),
		Marshaller:                    hcf.coreComponents.InternalMarshalizer(),
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
		AppStatusHandler:                    hcf.statusCoreComponents.AppStatusHandler(),
		TimeBetweenConnectionsMetricsUpdate: time.Second * time.Duration(cfg.TimeBetweenConnectionsMetricsUpdateInSec),
	}
	statusHandler, err := status.NewMetricsUpdater(argsMetricsUpdater)
	if err != nil {
		return nil, err
	}

	argsMainDirectConnectionProcessor := processor.ArgsDirectConnectionProcessor{
		TimeToReadDirectConnections: time.Second * time.Duration(cfg.TimeToReadDirectConnectionsInSec),
		Messenger:                   hcf.networkComponents.NetworkMessenger(),
		PeerShardMapper:             hcf.processComponents.PeerShardMapper(),
		ShardCoordinator:            hcf.processComponents.ShardCoordinator(),
		BaseIntraShardTopic:         common.ConsensusTopic,
		BaseCrossShardTopic:         processFactory.MiniBlocksTopic,
	}
	mainDirectConnectionProcessor, err := processor.NewDirectConnectionProcessor(argsMainDirectConnectionProcessor)
	if err != nil {
		return nil, err
	}

	argsFullArchiveDirectConnectionProcessor := processor.ArgsDirectConnectionProcessor{
		TimeToReadDirectConnections: time.Second * time.Duration(cfg.TimeToReadDirectConnectionsInSec),
		Messenger:                   hcf.networkComponents.FullArchiveNetworkMessenger(),
		PeerShardMapper:             hcf.processComponents.PeerShardMapper(), // TODO[Sorin]: replace this with the full archive psm
		ShardCoordinator:            hcf.processComponents.ShardCoordinator(),
		BaseIntraShardTopic:         common.ConsensusTopic,
		BaseCrossShardTopic:         processFactory.MiniBlocksTopic,
	}
	fullArchiveDirectConnectionProcessor, err := processor.NewDirectConnectionProcessor(argsFullArchiveDirectConnectionProcessor)
	if err != nil {
		return nil, err
	}

	argsMainCrossShardPeerTopicNotifier := monitor.ArgsCrossShardPeerTopicNotifier{
		ShardCoordinator: hcf.processComponents.ShardCoordinator(),
		PeerShardMapper:  hcf.processComponents.PeerShardMapper(),
	}
	mainCrossShardPeerTopicNotifier, err := monitor.NewCrossShardPeerTopicNotifier(argsMainCrossShardPeerTopicNotifier)
	if err != nil {
		return nil, err
	}
	err = hcf.networkComponents.NetworkMessenger().AddPeerTopicNotifier(mainCrossShardPeerTopicNotifier)
	if err != nil {
		return nil, err
	}

	argsFullArchiveCrossShardPeerTopicNotifier := monitor.ArgsCrossShardPeerTopicNotifier{
		ShardCoordinator: hcf.processComponents.ShardCoordinator(),
		PeerShardMapper:  hcf.processComponents.PeerShardMapper(), // TODO[Sorin]: replace this with the full archive psm
	}
	fullArchiveCrossShardPeerTopicNotifier, err := monitor.NewCrossShardPeerTopicNotifier(argsFullArchiveCrossShardPeerTopicNotifier)
	if err != nil {
		return nil, err
	}
	err = hcf.networkComponents.FullArchiveNetworkMessenger().AddPeerTopicNotifier(fullArchiveCrossShardPeerTopicNotifier)
	if err != nil {
		return nil, err
	}

	return &heartbeatV2Components{
		sender:                               heartbeatV2Sender,
		peerAuthRequestsProcessor:            paRequestsProcessor,
		shardSender:                          shardSender,
		monitor:                              heartbeatsMonitor,
		statusHandler:                        statusHandler,
		mainDirectConnectionProcessor:        mainDirectConnectionProcessor,
		fullArchiveDirectConnectionProcessor: fullArchiveDirectConnectionProcessor,
	}, nil
}

func (hcf *heartbeatV2ComponentsFactory) createTopicsIfNeeded() error {
	err := createTopicIfNeededOnMessenger(hcf.networkComponents.NetworkMessenger())
	if err != nil {
		return err
	}

	return createTopicIfNeededOnMessenger(hcf.networkComponents.FullArchiveNetworkMessenger())
}

func createTopicIfNeededOnMessenger(messenger p2p.Messenger) error {
	if !messenger.HasTopic(common.PeerAuthenticationTopic) {
		err := messenger.CreateTopic(common.PeerAuthenticationTopic, true)
		if err != nil {
			return err
		}
	}
	if !messenger.HasTopic(common.HeartbeatV2Topic) {
		err := messenger.CreateTopic(common.HeartbeatV2Topic, true)
		if err != nil {
			return err
		}
	}

	return nil
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

	if !check.IfNil(hc.mainDirectConnectionProcessor) {
		log.LogIfError(hc.mainDirectConnectionProcessor.Close())
	}

	if !check.IfNil(hc.fullArchiveDirectConnectionProcessor) {
		log.LogIfError(hc.fullArchiveDirectConnectionProcessor.Close())
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hcf *heartbeatV2ComponentsFactory) IsInterfaceNil() bool {
	return hcf == nil
}
