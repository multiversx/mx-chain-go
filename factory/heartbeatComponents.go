package factory

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	heartbeatProcess "github.com/ElrondNetwork/elrond-go/heartbeat/process"
	heartbeatStorage "github.com/ElrondNetwork/elrond-go/heartbeat/storage"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/peer"
)

// HeartbeatComponentsFactoryArgs holds the arguments needed to create a heartbeat components factory
type HeartbeatComponentsFactoryArgs struct {
	Config            config.Config
	Prefs             config.Preferences
	AppVersion        string
	GenesisTime       time.Time
	HardforkTrigger   heartbeat.HardforkTrigger
	CoreComponents    CoreComponentsHolder
	DataComponents    DataComponentsHolder
	NetworkComponents NetworkComponentsHolder
	CryptoComponents  CryptoComponentsHolder
	ProcessComponents ProcessComponentsHolder
}

type heartbeatComponentsFactory struct {
	config            config.Config
	prefs             config.Preferences
	version           string
	GenesisTime       time.Time
	hardforkTrigger   heartbeat.HardforkTrigger
	coreComponents    CoreComponentsHolder
	dataComponents    DataComponentsHolder
	networkComponents NetworkComponentsHolder
	cryptoComponents  CryptoComponentsHolder
	processComponents ProcessComponentsHolder
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

	if check.IfNil(args.HardforkTrigger) {
		return nil, heartbeat.ErrNilHardforkTrigger
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

	return &heartbeatComponentsFactory{
		config:            args.Config,
		prefs:             args.Prefs,
		version:           args.AppVersion,
		GenesisTime:       args.GenesisTime,
		hardforkTrigger:   args.HardforkTrigger,
		coreComponents:    args.CoreComponents,
		dataComponents:    args.DataComponents,
		networkComponents: args.NetworkComponents,
		cryptoComponents:  args.CryptoComponents,
		processComponents: args.ProcessComponents,
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

	if !hcf.networkComponents.NetworkMessenger().HasTopic(core.HeartbeatTopic) {
		err = hcf.networkComponents.NetworkMessenger().CreateTopic(core.HeartbeatTopic, true)
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
	if hcf.config.StoragePruning.FullArchive {
		peerSubType = core.FullHistoryObserver
	}

	argSender := heartbeatProcess.ArgHeartbeatSender{
		PeerMessenger:        hcf.networkComponents.NetworkMessenger(),
		PeerSignatureHandler: hcf.cryptoComponents.PeerSignatureHandler(),
		PrivKey:              hcf.cryptoComponents.PrivateKey(),
		Marshalizer:          hcf.coreComponents.InternalMarshalizer(),
		Topic:                core.HeartbeatTopic,
		ShardCoordinator:     hcf.processComponents.ShardCoordinator(),
		PeerTypeProvider:     peerTypeProvider,
		StatusHandler:        hcf.coreComponents.StatusHandler(),
		VersionNumber:        hcf.version,
		NodeDisplayName:      hcf.prefs.Preferences.NodeDisplayName,
		KeyBaseIdentity:      hcf.prefs.Preferences.Identity,
		HardforkTrigger:      hcf.hardforkTrigger,
		CurrentBlockProvider: hcf.dataComponents.Blockchain(),
		PeerSubType:          peerSubType,
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
		HardforkTrigger:                    hcf.hardforkTrigger,
		ValidatorPubkeyConverter:           hcf.coreComponents.ValidatorPubKeyConverter(),
		HeartbeatRefreshIntervalInSec:      hcf.config.Heartbeat.HeartbeatRefreshIntervalInSec,
		HideInactiveValidatorIntervalInSec: hcf.config.Heartbeat.HideInactiveValidatorIntervalInSec,
		AppStatusHandler:                   hcf.coreComponents.StatusHandler(),
	}
	hbc.monitor, err = heartbeatProcess.NewMonitor(argMonitor)
	if err != nil {
		return nil, err
	}

	log.Debug("heartbeat's monitor component has been instantiated")

	err = hcf.networkComponents.NetworkMessenger().RegisterMessageProcessor(
		core.HeartbeatTopic, core.DefaultInterceptorsIdentifier, hbc.monitor,
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

	for {
		randomNanos := r.Int63n(diffNanos)
		timeToWait := time.Second*time.Duration(cfg.MinTimeToWaitBetweenBroadcastsInSec) + time.Duration(randomNanos)

		select {
		case <-ctx.Done():
			log.Debug("heartbeat's go routine is stopping...")
			return
		case <-time.After(timeToWait):
		case <-hcf.hardforkTrigger.NotifyTriggerReceived():
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
