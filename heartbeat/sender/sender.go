package sender

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
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
}

// sender defines the component which sends authentication and heartbeat messages
type sender struct {
	routineHandler *routineHandler
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
		},
		nodesCoordinator:         args.NodesCoordinator,
		peerSignatureHandler:     args.PeerSignatureHandler,
		privKey:                  args.PrivateKey,
		redundancyHandler:        args.RedundancyHandler,
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
		},
		versionNumber:        args.VersionNumber,
		nodeDisplayName:      args.NodeDisplayName,
		identity:             args.Identity,
		peerSubType:          args.PeerSubType,
		currentBlockProvider: args.CurrentBlockProvider,
	})
	if err != nil {
		return nil, err
	}

	return &sender{
		routineHandler: newRoutineHandler(pas, hbs, pas),
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
		},
		nodesCoordinator:         args.NodesCoordinator,
		peerSignatureHandler:     args.PeerSignatureHandler,
		privKey:                  args.PrivateKey,
		redundancyHandler:        args.RedundancyHandler,
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
		},
		versionNumber:        args.VersionNumber,
		nodeDisplayName:      args.NodeDisplayName,
		identity:             args.Identity,
		peerSubType:          args.PeerSubType,
		currentBlockProvider: args.CurrentBlockProvider,
	}
	return checkHeartbeatSenderArgs(hbsArgs)
}

// Close closes the internal components
func (sender *sender) Close() error {
	sender.routineHandler.closeProcessLoop()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *sender) IsInterfaceNil() bool {
	return sender == nil
}
