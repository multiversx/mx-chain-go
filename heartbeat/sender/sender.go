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
	BaseVersionNumber                           string
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
	KeysHolder                                  heartbeat.KeysHolder
	PeerAuthenticationTimeBetweenChecks         time.Duration
	ShardCoordinator                            heartbeat.ShardCoordinator
}

// sender defines the component which sends authentication and heartbeat messages
type sender struct {
	heartbeatSender heartbeatSenderHandler
	routineHandler  *routineHandler
}

// NewSender creates a new instance of sender
func NewSender(args ArgSender) (*sender, error) {
	err := checkSenderArgs(args)
	if err != nil {
		return nil, err
	}

	pas, err := createPeerAuthenticationSender(argPeerAuthenticationSenderFactory{
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
		keysHolder:               args.KeysHolder,
		timeBetweenChecks:        args.PeerAuthenticationTimeBetweenChecks,
		shardCoordinator:         args.ShardCoordinator,
	})
	if err != nil {
		return nil, err
	}

	hbs, err := createHeartbeatSender(argHeartbeatSenderFactory{
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
		baseVersionNumber:    args.BaseVersionNumber,
		versionNumber:        args.VersionNumber,
		nodeDisplayName:      args.NodeDisplayName,
		identity:             args.Identity,
		peerSubType:          args.PeerSubType,
		currentBlockProvider: args.CurrentBlockProvider,
		peerTypeProvider:     args.PeerTypeProvider,
		keysHolder:           args.KeysHolder,
		shardCoordinator:     args.ShardCoordinator,
		nodesCoordinator:     args.NodesCoordinator,
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
	basePeerAuthSenderArgs := argBaseSender{
		messenger:                 args.Messenger,
		marshaller:                args.Marshaller,
		topic:                     args.PeerAuthenticationTopic,
		timeBetweenSends:          args.PeerAuthenticationTimeBetweenSends,
		timeBetweenSendsWhenError: args.PeerAuthenticationTimeBetweenSendsWhenError,
		thresholdBetweenSends:     args.PeerAuthenticationThresholdBetweenSends,
		privKey:                   args.PrivateKey,
		redundancyHandler:         args.RedundancyHandler,
	}
	pasArgs := argPeerAuthenticationSender{
		argBaseSender:            basePeerAuthSenderArgs,
		nodesCoordinator:         args.NodesCoordinator,
		peerSignatureHandler:     args.PeerSignatureHandler,
		hardforkTrigger:          args.HardforkTrigger,
		hardforkTimeBetweenSends: args.HardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.HardforkTriggerPubKey,
	}
	err := checkPeerAuthenticationSenderArgs(pasArgs)
	if err != nil {
		return err
	}

	mpasArgs := argMultikeyPeerAuthenticationSender{
		argBaseSender:            basePeerAuthSenderArgs,
		nodesCoordinator:         args.NodesCoordinator,
		peerSignatureHandler:     args.PeerSignatureHandler,
		hardforkTrigger:          args.HardforkTrigger,
		hardforkTimeBetweenSends: args.HardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.HardforkTriggerPubKey,
		keysHolder:               args.KeysHolder,
		timeBetweenChecks:        args.PeerAuthenticationTimeBetweenChecks,
		shardCoordinator:         args.ShardCoordinator,
	}
	err = checkMultikeyPeerAuthenticationSenderArgs(mpasArgs)
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
		versionNumber:        args.VersionNumber,
		nodeDisplayName:      args.NodeDisplayName,
		identity:             args.Identity,
		peerSubType:          args.PeerSubType,
		currentBlockProvider: args.CurrentBlockProvider,
		peerTypeProvider:     args.PeerTypeProvider,
	}
	err = checkHeartbeatSenderArgs(hbsArgs)
	if err != nil {
		return err
	}

	mhbsArgs := argMultikeyHeartbeatSender{
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
		peerTypeProvider:     args.PeerTypeProvider,
		versionNumber:        args.VersionNumber,
		baseVersionNumber:    args.BaseVersionNumber,
		nodeDisplayName:      args.NodeDisplayName,
		identity:             args.Identity,
		peerSubType:          args.PeerSubType,
		currentBlockProvider: args.CurrentBlockProvider,
		keysHolder:           args.KeysHolder,
		shardCoordinator:     args.ShardCoordinator,
	}

	return checkMultikeyHeartbeatSenderArgs(mhbsArgs)
}

// Close closes the internal components
func (sender *sender) Close() error {
	sender.routineHandler.closeProcessLoop()

	return nil
}

// GetCurrentNodeType will return the current peer details
func (sender *sender) GetCurrentNodeType() (string, core.P2PPeerSubType, error) {
	return sender.heartbeatSender.GetCurrentNodeType()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *sender) IsInterfaceNil() bool {
	return sender == nil
}
