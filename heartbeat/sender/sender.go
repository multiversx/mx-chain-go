package sender

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/heartbeat/sender/disabled"
)

// ArgSender represents the arguments for the sender
type ArgSender struct {
	Messenger                                   heartbeat.P2PMessenger
	Marshaller                                  marshal.Marshalizer
	PeerAuthenticationTopic                     string
	HeartbeatTopic                              string
	PeerAuthenticationTimeBetweenSends          time.Duration
	PeerAuthenticationTimeBetweenSendsWhenError time.Duration
	PeerAuthenticationThresholdBetweenSends     float64
	HeartbeatTimeBetweenSends                   time.Duration
	HeartbeatTimeBetweenSendsWhenError          time.Duration
	HeartbeatThresholdBetweenSends              float64
	VersionNumber                               string
	NodeDisplayName                             string
	Identity                                    string
	PeerSubType                                 core.P2PPeerSubType
	CurrentBlockProvider                        heartbeat.CurrentBlockProvider
	PeerSignatureHandler                        crypto.PeerSignatureHandler
	PrivateKey                                  crypto.PrivateKey
	RedundancyHandler                           heartbeat.NodeRedundancyHandler
	NodesCoordinator                            heartbeat.NodesCoordinator
	HardforkTrigger                             heartbeat.HardforkTrigger
	HardforkTimeBetweenSends                    time.Duration
	HardforkTriggerPubKey                       []byte
	PeerTypeProvider                            heartbeat.PeerTypeProviderHandler
}

// sender defines the component which sends authentication and heartbeat messages
type sender struct {
	heartbeatSender *heartbeatSender
	routineHandler  *routineHandler
}

// NewSender creates a new instance of sender
func NewSender(args ArgSender) (*sender, error) {
	err := checkSenderArgs(args)
	if err != nil {
		return nil, err
	}

	pas, err := newPeerAuthenticationSender(argPeerAuthenticationSender{
		argBaseSender: argBaseSender{
			messenger:                 args.Messenger,
			marshaller:                args.Marshaller,
			topic:                     args.PeerAuthenticationTopic,
			timeBetweenSends:          args.PeerAuthenticationTimeBetweenSends,
			timeBetweenSendsWhenError: args.PeerAuthenticationTimeBetweenSendsWhenError,
			thresholdBetweenSends:     args.PeerAuthenticationThresholdBetweenSends,
			privKey:                   args.PrivateKey,
			redundancyHandler:         args.RedundancyHandler,
		},
		nodesCoordinator:         args.NodesCoordinator,
		peerSignatureHandler:     args.PeerSignatureHandler,
		hardforkTrigger:          args.HardforkTrigger,
		hardforkTimeBetweenSends: args.HardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.HardforkTriggerPubKey,
	})
	if err != nil {
		return nil, err
	}

	hbs, err := newHeartbeatSender(argHeartbeatSender{
		argBaseSender: argBaseSender{
			messenger:                 args.Messenger,
			marshaller:                args.Marshaller,
			topic:                     args.HeartbeatTopic,
			timeBetweenSends:          args.HeartbeatTimeBetweenSends,
			timeBetweenSendsWhenError: args.HeartbeatTimeBetweenSendsWhenError,
			thresholdBetweenSends:     args.HeartbeatThresholdBetweenSends,
			privKey:                   args.PrivateKey,
			redundancyHandler:         args.RedundancyHandler,
		},
		versionNumber:              args.VersionNumber,
		nodeDisplayName:            args.NodeDisplayName,
		identity:                   args.Identity,
		peerSubType:                args.PeerSubType,
		currentBlockProvider:       args.CurrentBlockProvider,
		peerTypeProvider:           args.PeerTypeProvider,
		trieSyncStatisticsProvider: disabled.NewTrieSyncStatisticsProvider(),
	})
	if err != nil {
		return nil, err
	}

	return &sender{
		heartbeatSender: hbs,
		routineHandler:  newRoutineHandler(pas, hbs, pas),
	}, nil
}

func checkSenderArgs(args ArgSender) error {
	pasArg := argPeerAuthenticationSender{
		argBaseSender: argBaseSender{
			messenger:                 args.Messenger,
			marshaller:                args.Marshaller,
			topic:                     args.PeerAuthenticationTopic,
			timeBetweenSends:          args.PeerAuthenticationTimeBetweenSends,
			timeBetweenSendsWhenError: args.PeerAuthenticationTimeBetweenSendsWhenError,
			thresholdBetweenSends:     args.PeerAuthenticationThresholdBetweenSends,
			privKey:                   args.PrivateKey,
			redundancyHandler:         args.RedundancyHandler,
		},
		nodesCoordinator:         args.NodesCoordinator,
		peerSignatureHandler:     args.PeerSignatureHandler,
		hardforkTrigger:          args.HardforkTrigger,
		hardforkTimeBetweenSends: args.HardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.HardforkTriggerPubKey,
	}
	err := checkPeerAuthenticationSenderArgs(pasArg)
	if err != nil {
		return err
	}

	hbsArgs := argHeartbeatSender{
		argBaseSender: argBaseSender{
			messenger:                 args.Messenger,
			marshaller:                args.Marshaller,
			topic:                     args.HeartbeatTopic,
			timeBetweenSends:          args.HeartbeatTimeBetweenSends,
			timeBetweenSendsWhenError: args.HeartbeatTimeBetweenSendsWhenError,
			thresholdBetweenSends:     args.HeartbeatThresholdBetweenSends,
			privKey:                   args.PrivateKey,
			redundancyHandler:         args.RedundancyHandler,
		},
		versionNumber:              args.VersionNumber,
		nodeDisplayName:            args.NodeDisplayName,
		identity:                   args.Identity,
		peerSubType:                args.PeerSubType,
		currentBlockProvider:       args.CurrentBlockProvider,
		peerTypeProvider:           args.PeerTypeProvider,
		trieSyncStatisticsProvider: disabled.NewTrieSyncStatisticsProvider(),
	}
	return checkHeartbeatSenderArgs(hbsArgs)
}

// Close closes the internal components
func (sender *sender) Close() error {
	sender.routineHandler.closeProcessLoop()

	return nil
}

// GetSenderInfo will return the current sender info
func (sender *sender) GetSenderInfo() (string, core.P2PPeerSubType, error) {
	return sender.heartbeatSender.getSenderInfo()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *sender) IsInterfaceNil() bool {
	return sender == nil
}
