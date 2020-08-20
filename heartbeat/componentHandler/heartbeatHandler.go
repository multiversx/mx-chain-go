package componentHandler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/process"
	heartbeatStorage "github.com/ElrondNetwork/elrond-go/heartbeat/storage"
	"github.com/ElrondNetwork/elrond-go/marshal"
	peerProcess "github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/peer"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("heartbeat/componenthandler")

// ArgHeartbeat represents the heartbeat creation argument
type ArgHeartbeat struct {
	HeartbeatConfig          config.HeartbeatConfig
	PrefsConfig              config.PreferencesConfig
	Marshalizer              marshal.Marshalizer
	Messenger                heartbeat.P2PMessenger
	ShardCoordinator         sharding.Coordinator
	NodesCoordinator         sharding.NodesCoordinator
	AppStatusHandler         core.AppStatusHandler
	Storer                   storage.Storer
	ValidatorStatistics      heartbeat.ValidatorStatisticsProcessor
	PeerSignatureHandler     crypto.PeerSignatureHandler
	PrivKey                  crypto.PrivateKey
	HardforkTrigger          heartbeat.HardforkTrigger
	AntifloodHandler         heartbeat.P2PAntifloodHandler
	ValidatorPubkeyConverter core.PubkeyConverter
	EpochStartTrigger        sharding.EpochHandler
	EpochStartRegistration   sharding.EpochStartEventNotifier
	Timer                    heartbeat.Timer
	GenesisTime              time.Time
	VersionNumber            string
	PeerShardMapper          heartbeat.NetworkShardingCollector
	SizeCheckDelta           uint32
	ValidatorsProvider       peerProcess.ValidatorsProvider
	CurrentBlockProvider     heartbeat.CurrentBlockProvider
}

// HeartbeatHandler is the struct used to manage heartbeat subsystem consisting of a heartbeat sender and monitor
// wired on a dedicated p2p topic
type HeartbeatHandler struct {
	monitor          *process.Monitor
	sender           *process.Sender
	arg              ArgHeartbeat
	peerTypeProvider *peer.PeerTypeProvider
	cancelFunc       func()
}

// NewHeartbeatHandler will create a heartbeat handler containing both a monitor and a sender
func NewHeartbeatHandler(arg ArgHeartbeat) (*HeartbeatHandler, error) {
	hbh := &HeartbeatHandler{
		arg: arg,
	}

	err := hbh.create()
	if err != nil {
		return nil, err
	}

	return hbh, nil
}

func (hbh *HeartbeatHandler) create() error {
	arg := hbh.arg

	var ctx context.Context
	ctx, hbh.cancelFunc = context.WithCancel(context.Background())

	err := hbh.checkConfigParams(arg.HeartbeatConfig)
	if err != nil {
		return err
	}

	if check.IfNil(arg.Messenger) {
		return heartbeat.ErrNilMessenger
	}

	if arg.Messenger.HasTopicValidator(core.HeartbeatTopic) {
		return heartbeat.ErrValidatorAlreadySet
	}

	if !arg.Messenger.HasTopic(core.HeartbeatTopic) {
		err = arg.Messenger.CreateTopic(core.HeartbeatTopic, true)
		if err != nil {
			return err
		}
	}
	argPeerTypeProvider := peer.ArgPeerTypeProvider{
		NodesCoordinator:        arg.NodesCoordinator,
		StartEpoch:              arg.EpochStartTrigger.MetaEpoch(),
		EpochStartEventNotifier: arg.EpochStartRegistration,
	}
	peerTypeProvider, err := peer.NewPeerTypeProvider(argPeerTypeProvider)
	if err != nil {
		return err
	}
	hbh.peerTypeProvider = peerTypeProvider
	argSender := process.ArgHeartbeatSender{
		PeerMessenger:        arg.Messenger,
		PeerSignatureHandler: arg.PeerSignatureHandler,
		PrivKey:              arg.PrivKey,
		Marshalizer:          arg.Marshalizer,
		Topic:                core.HeartbeatTopic,
		ShardCoordinator:     arg.ShardCoordinator,
		PeerTypeProvider:     peerTypeProvider,
		StatusHandler:        arg.AppStatusHandler,
		VersionNumber:        arg.VersionNumber,
		NodeDisplayName:      arg.PrefsConfig.NodeDisplayName,
		KeyBaseIdentity:      arg.PrefsConfig.Identity,
		HardforkTrigger:      arg.HardforkTrigger,
		CurrentBlockProvider: arg.CurrentBlockProvider,
	}

	hbh.sender, err = process.NewSender(argSender)
	if err != nil {
		return err
	}

	log.Debug("heartbeat's sender component has been instantiated")

	heartBeatMsgProcessor, err := process.NewMessageProcessor(
		arg.PeerSignatureHandler,
		arg.Marshalizer,
		arg.PeerShardMapper,
	)
	if err != nil {
		return err
	}

	heartbeatStorer, err := heartbeatStorage.NewHeartbeatDbStorer(arg.Storer, arg.Marshalizer)
	if err != nil {
		return err
	}

	timer := &process.RealTimer{}
	netInputMarshalizer := arg.Marshalizer
	if arg.SizeCheckDelta > 0 {
		netInputMarshalizer = marshal.NewSizeCheckUnmarshalizer(arg.Marshalizer, arg.SizeCheckDelta)
	}

	allValidators, _, _ := hbh.getLatestValidators()
	pubKeysMap := make(map[uint32][]string)
	for shardID, valsInShard := range allValidators {
		for _, val := range valsInShard {
			pubKeysMap[shardID] = append(pubKeysMap[shardID], string(val.PublicKey))
		}
	}

	argMonitor := process.ArgHeartbeatMonitor{
		Marshalizer:                        netInputMarshalizer,
		MaxDurationPeerUnresponsive:        time.Second * time.Duration(arg.HeartbeatConfig.DurationToConsiderUnresponsiveInSec),
		PubKeysMap:                         pubKeysMap,
		GenesisTime:                        arg.GenesisTime,
		MessageHandler:                     heartBeatMsgProcessor,
		Storer:                             heartbeatStorer,
		PeerTypeProvider:                   peerTypeProvider,
		Timer:                              timer,
		AntifloodHandler:                   arg.AntifloodHandler,
		HardforkTrigger:                    arg.HardforkTrigger,
		ValidatorPubkeyConverter:           arg.ValidatorPubkeyConverter,
		HeartbeatRefreshIntervalInSec:      arg.HeartbeatConfig.HeartbeatRefreshIntervalInSec,
		HideInactiveValidatorIntervalInSec: arg.HeartbeatConfig.HideInactiveValidatorIntervalInSec,
	}
	hbh.monitor, err = process.NewMonitor(argMonitor)
	if err != nil {
		return err
	}

	log.Debug("heartbeat's monitor component has been instantiated")

	err = hbh.monitor.SetAppStatusHandler(arg.AppStatusHandler)
	if err != nil {
		return err
	}

	err = arg.Messenger.RegisterMessageProcessor(core.HeartbeatTopic, hbh.monitor)
	if err != nil {
		return err
	}

	go hbh.startSendingHeartbeats(ctx)

	return nil
}

func (hbh *HeartbeatHandler) getLatestValidators() (map[uint32][]*state.ValidatorInfo, map[string]*state.ValidatorApiResponse, error) {
	latestHash, err := hbh.arg.ValidatorStatistics.RootHash()
	if err != nil {
		return nil, nil, err
	}

	validators, err := hbh.arg.ValidatorStatistics.GetValidatorInfoForRootHash(latestHash)
	if err != nil {
		return nil, nil, err
	}

	return validators, nil, nil
}

func (hbh *HeartbeatHandler) startSendingHeartbeats(ctx context.Context) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	cfg := hbh.arg.HeartbeatConfig

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
		case <-hbh.arg.HardforkTrigger.NotifyTriggerReceived(): //this will force an immediate broadcast of the trigger
			//message on the network
		}

		err := hbh.sender.SendHeartbeat()
		if err != nil {
			log.Debug("SendHeartbeat", "error", err.Error())
		}
	}
}

func (hbh *HeartbeatHandler) checkConfigParams(config config.HeartbeatConfig) error {
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

// Monitor returns the monitor component
func (hbh *HeartbeatHandler) Monitor() *process.Monitor {
	return hbh.monitor
}

// Sender returns the sender component
func (hbh *HeartbeatHandler) Sender() *process.Sender {
	return hbh.sender
}

// Close will close the endless running go routine
func (hbh *HeartbeatHandler) Close() error {
	hbh.cancelFunc()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hbh *HeartbeatHandler) IsInterfaceNil() bool {
	return hbh == nil
}
