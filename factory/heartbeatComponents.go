package factory

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	heartbeatProcess "github.com/ElrondNetwork/elrond-go/heartbeat/process"
	heartbeatStorage "github.com/ElrondNetwork/elrond-go/heartbeat/storage"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/state"
)

// HeartbeatComponentsFactoryArgs holds the arguments needed to create a heartbeat components factory
type HeartbeatComponentsFactoryArgs struct {
	Config                config.Config
	Prefs                 config.Preferences
	AppVersion            string
	GenesisTime           time.Time
	RedundancyHandler     heartbeat.NodeRedundancyHandler
	CoreComponents        CoreComponentsHolder
	DataComponents        DataComponentsHolder
	NetworkComponents     NetworkComponentsHolder
	CryptoComponents      CryptoComponentsHolder
	ProcessComponents     ProcessComponentsHolder
	HeartbeatDisableEpoch uint32
}

type heartbeatComponentsFactory struct {
	config                config.Config
	prefs                 config.Preferences
	version               string
	GenesisTime           time.Time
	redundancyHandler     heartbeat.NodeRedundancyHandler
	coreComponents        CoreComponentsHolder
	dataComponents        DataComponentsHolder
	networkComponents     NetworkComponentsHolder
	cryptoComponents      CryptoComponentsHolder
	processComponents     ProcessComponentsHolder
	heartbeatDisableEpoch uint32
}

type heartbeatComponents struct {
	messageHandler heartbeat.MessageHandler
	monitor        HeartbeatMonitor
	sender         HeartbeatSender
	storer         HeartbeatStorer
	cancelFunc     context.CancelFunc
}

// NewHeartbeatComponentsFactory creates the heartbeat components factory
func NewHeartbeatComponentsFactory(args HeartbeatComponentsFactoryArgs) (*heartbeatComponentsFactory, error) {

	if check.IfNil(args.RedundancyHandler) {
		return nil, heartbeat.ErrNilRedundancyHandler
	}
	if check.IfNil(args.CoreComponents) {
		return nil, errors.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.DataComponents) {
		return nil, errors.ErrNilDataComponentsHolder
	}
	if check.IfNil(args.NetworkComponents) {
		return nil, errors.ErrNilNetworkComponentsHolder
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, errors.ErrNilCryptoComponentsHolder
	}
	if check.IfNil(args.ProcessComponents) {
		return nil, errors.ErrNilProcessComponentsHolder
	}
	hardforkTrigger := args.ProcessComponents.HardforkTrigger()
	if check.IfNil(hardforkTrigger) {
		return nil, heartbeat.ErrNilHardforkTrigger
	}

	return &heartbeatComponentsFactory{
		config:                args.Config,
		prefs:                 args.Prefs,
		version:               args.AppVersion,
		GenesisTime:           args.GenesisTime,
		redundancyHandler:     args.RedundancyHandler,
		coreComponents:        args.CoreComponents,
		dataComponents:        args.DataComponents,
		networkComponents:     args.NetworkComponents,
		cryptoComponents:      args.CryptoComponents,
		processComponents:     args.ProcessComponents,
		heartbeatDisableEpoch: args.HeartbeatDisableEpoch,
	}, nil
}

// Create creates the heartbeat components
func (hcf *heartbeatComponentsFactory) Create() (*heartbeatComponents, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	hbc := &heartbeatComponents{
		cancelFunc: cancelFunc,
	}

	err := checkConfigParams(hcf.config.Heartbeat)
	if err != nil {
		return nil, err
	}
	if check.IfNil(hcf.networkComponents) {
		return nil, errors.ErrNilNetworkComponentsHolder
	}
	if check.IfNil(hcf.networkComponents.NetworkMessenger()) {
		return nil, errors.ErrNilMessenger
	}
	if !hcf.networkComponents.NetworkMessenger().HasTopic(common.HeartbeatTopic) {
		err = hcf.networkComponents.NetworkMessenger().CreateTopic(common.HeartbeatTopic, true)
		if err != nil {
			return nil, err
		}
	}
	argPeerTypeProvider := peer.ArgPeerTypeProvider{
		NodesCoordinator:        hcf.processComponents.NodesCoordinator(),
		StartEpoch:              hcf.processComponents.EpochStartTrigger().MetaEpoch(),
		EpochStartEventNotifier: hcf.processComponents.EpochStartNotifier(),
	}
	peerTypeProvider, err := peer.NewPeerTypeProvider(argPeerTypeProvider)
	if err != nil {
		return nil, err
	}

	peerSubType := core.RegularPeer
	if hcf.prefs.Preferences.FullArchive {
		peerSubType = core.FullHistoryObserver
	}

	hardforkTrigger := hcf.processComponents.HardforkTrigger()

	argSender := heartbeatProcess.ArgHeartbeatSender{
		PeerSubType:           peerSubType,
		PeerMessenger:         hcf.networkComponents.NetworkMessenger(),
		PeerSignatureHandler:  hcf.cryptoComponents.PeerSignatureHandler(),
		PrivKey:               hcf.cryptoComponents.PrivateKey(),
		Marshalizer:           hcf.coreComponents.InternalMarshalizer(),
		Topic:                 common.HeartbeatTopic,
		ShardCoordinator:      hcf.processComponents.ShardCoordinator(),
		PeerTypeProvider:      peerTypeProvider,
		StatusHandler:         hcf.coreComponents.StatusHandler(),
		VersionNumber:         hcf.version,
		NodeDisplayName:       hcf.prefs.Preferences.NodeDisplayName,
		KeyBaseIdentity:       hcf.prefs.Preferences.Identity,
		HardforkTrigger:       hardforkTrigger,
		CurrentBlockProvider:  hcf.dataComponents.Blockchain(),
		RedundancyHandler:     hcf.redundancyHandler,
		EpochNotifier:         hcf.coreComponents.EpochNotifier(),
		HeartbeatDisableEpoch: hcf.heartbeatDisableEpoch,
	}

	hbc.sender, err = heartbeatProcess.NewSender(argSender)
	if err != nil {
		return nil, err
	}

	log.Debug("heartbeat's sender component has been instantiated")

	hbc.messageHandler, err = heartbeatProcess.NewMessageProcessor(
		hcf.cryptoComponents.PeerSignatureHandler(),
		hcf.coreComponents.InternalMarshalizer(),
		hcf.processComponents.PeerShardMapper(),
	)
	if err != nil {
		return nil, err
	}
	storer := hcf.dataComponents.StorageService().GetStorer(dataRetriever.HeartbeatUnit)
	marshalizer := hcf.coreComponents.InternalMarshalizer()
	heartbeatStorer, err := heartbeatStorage.NewHeartbeatDbStorer(storer, marshalizer)
	if err != nil {
		return nil, err
	}

	hbc.storer = heartbeatStorer

	timer := &heartbeatProcess.RealTimer{}
	if hcf.config.Marshalizer.SizeCheckDelta > 0 {
		marshalizer = marshal.NewSizeCheckUnmarshalizer(marshalizer, hcf.config.Marshalizer.SizeCheckDelta)
	}

	allValidators, _, _ := hcf.getLatestValidators()
	pubKeysMap := make(map[uint32][]string)
	for shardID, valsInShard := range allValidators {
		for _, val := range valsInShard {
			pubKeysMap[shardID] = append(pubKeysMap[shardID], string(val.PublicKey))
		}
	}

	unresponsivePeerDuration := time.Second * time.Duration(hcf.config.Heartbeat.DurationToConsiderUnresponsiveInSec)
	argMonitor := heartbeatProcess.ArgHeartbeatMonitor{
		Marshalizer:                        marshalizer,
		MaxDurationPeerUnresponsive:        unresponsivePeerDuration,
		PubKeysMap:                         pubKeysMap,
		GenesisTime:                        hcf.GenesisTime,
		MessageHandler:                     hbc.messageHandler,
		Storer:                             heartbeatStorer,
		PeerTypeProvider:                   peerTypeProvider,
		Timer:                              timer,
		AntifloodHandler:                   hcf.networkComponents.InputAntiFloodHandler(),
		HardforkTrigger:                    hardforkTrigger,
		ValidatorPubkeyConverter:           hcf.coreComponents.ValidatorPubKeyConverter(),
		HeartbeatRefreshIntervalInSec:      hcf.config.Heartbeat.HeartbeatRefreshIntervalInSec,
		HideInactiveValidatorIntervalInSec: hcf.config.Heartbeat.HideInactiveValidatorIntervalInSec,
		AppStatusHandler:                   hcf.coreComponents.StatusHandler(),
		EpochNotifier:                      hcf.coreComponents.EpochNotifier(),
		HeartbeatDisableEpoch:              hcf.heartbeatDisableEpoch,
	}
	hbc.monitor, err = heartbeatProcess.NewMonitor(argMonitor)
	if err != nil {
		return nil, err
	}

	log.Debug("heartbeat's monitor component has been instantiated")

	err = hcf.networkComponents.NetworkMessenger().RegisterMessageProcessor(
		common.HeartbeatTopic,
		common.DefaultInterceptorsIdentifier,
		hbc.monitor,
	)
	if err != nil {
		return nil, err
	}

	go hcf.startSendingHeartbeats(ctx, hbc.sender, hbc.monitor)

	return hbc, nil
}

func (hcf *heartbeatComponentsFactory) getLatestValidators() (map[uint32][]*state.ValidatorInfo, map[string]*state.ValidatorApiResponse, error) {
	latestHash, err := hcf.processComponents.ValidatorsStatistics().RootHash()
	if err != nil {
		return nil, nil, err
	}

	validators, err := hcf.processComponents.ValidatorsStatistics().GetValidatorInfoForRootHash(latestHash)
	if err != nil {
		return nil, nil, err
	}

	return validators, nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hcf *heartbeatComponentsFactory) IsInterfaceNil() bool {
	return hcf == nil
}

func (hcf *heartbeatComponentsFactory) startSendingHeartbeats(ctx context.Context, sender HeartbeatSender, monitor HeartbeatMonitor) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	cfg := hcf.config.Heartbeat

	log.Debug("heartbeat's endless sending go routine started")

	diffSeconds := cfg.MaxTimeToWaitBetweenBroadcastsInSec - cfg.MinTimeToWaitBetweenBroadcastsInSec
	diffNanos := int64(diffSeconds) * time.Second.Nanoseconds()

	hardforkTrigger := hcf.processComponents.HardforkTrigger()
	for {
		randomNanos := r.Int63n(diffNanos)
		timeToWait := time.Second*time.Duration(cfg.MinTimeToWaitBetweenBroadcastsInSec) + time.Duration(randomNanos)

		select {
		case <-ctx.Done():
			log.Debug("heartbeat's go routine is stopping...")
			return
		case <-time.After(timeToWait):
		case <-hardforkTrigger.NotifyTriggerReceived():
			//this will force an immediate broadcast of the trigger
			//message on the network
			log.Debug("hardfork message prepared for heartbeat sending")
		}

		err := sender.SendHeartbeat()
		if err != nil {
			log.Debug("SendHeartbeat", "error", err.Error())
		}
		monitor.Cleanup()
	}
}

// Close closes the heartbeat components
func (hc *heartbeatComponents) Close() error {
	hc.cancelFunc()
	log.Debug("calling close on heartbeat system")

	if !check.IfNil(hc.monitor) {
		log.LogIfError(hc.monitor.Close())
	}

	return nil
}

func checkConfigParams(config config.HeartbeatConfig) error {
	if config.DurationToConsiderUnresponsiveInSec < 1 {
		return heartbeat.ErrInvalidDurationToConsiderUnresponsiveInSec
	}
	if config.MaxTimeToWaitBetweenBroadcastsInSec < 1 {
		return heartbeat.ErrNegativeMaxTimeToWaitBetweenBroadcastsInSec
	}
	if config.MinTimeToWaitBetweenBroadcastsInSec < 1 {
		return heartbeat.ErrNegativeMinTimeToWaitBetweenBroadcastsInSec
	}
	if config.MaxTimeToWaitBetweenBroadcastsInSec <= config.MinTimeToWaitBetweenBroadcastsInSec {
		return fmt.Errorf("%w for MaxTimeToWaitBetweenBroadcastsInSec", heartbeat.ErrWrongValues)
	}
	if config.DurationToConsiderUnresponsiveInSec <= config.MaxTimeToWaitBetweenBroadcastsInSec {
		return fmt.Errorf("%w for DurationToConsiderUnresponsiveInSec", heartbeat.ErrWrongValues)
	}

	return nil
}
