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
	MainMessenger                               heartbeat.P2PMessenger
	FullArchiveMessenger                        heartbeat.P2PMessenger
	Marshaller                                  marshal.Marshalizer
	PeerAuthenticationTopic                     string
	HeartbeatTopic                              string
	PeerAuthenticationTimeBetweenSends          time.Duration
	PeerAuthenticationTimeBetweenSendsWhenError time.Duration
	PeerAuthenticationTimeThresholdBetweenSends float64
	HeartbeatTimeBetweenSends                   time.Duration
	HeartbeatTimeBetweenSendsWhenError          time.Duration
	HeartbeatTimeThresholdBetweenSends          float64
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
	ManagedPeersHolder                          heartbeat.ManagedPeersHolder
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
			mainMessenger:             args.MainMessenger,
			fullArchiveMessenger:      args.FullArchiveMessenger,
			marshaller:                args.Marshaller,
			topic:                     args.PeerAuthenticationTopic,
			timeBetweenSends:          args.PeerAuthenticationTimeBetweenSends,
			timeBetweenSendsWhenError: args.PeerAuthenticationTimeBetweenSendsWhenError,
			thresholdBetweenSends:     args.PeerAuthenticationTimeThresholdBetweenSends,
			privKey:                   args.PrivateKey,
			redundancyHandler:         args.RedundancyHandler,
		},
		nodesCoordinator:         args.NodesCoordinator,
		peerSignatureHandler:     args.PeerSignatureHandler,
		hardforkTrigger:          args.HardforkTrigger,
		hardforkTimeBetweenSends: args.HardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.HardforkTriggerPubKey,
		managedPeersHolder:       args.ManagedPeersHolder,
		timeBetweenChecks:        args.PeerAuthenticationTimeBetweenChecks,
		shardCoordinator:         args.ShardCoordinator,
	})
	if err != nil {
		return nil, err
	}

	hbs, err := createHeartbeatSender(argHeartbeatSenderFactory{
		argBaseSender: argBaseSender{
			mainMessenger:             args.MainMessenger,
			fullArchiveMessenger:      args.FullArchiveMessenger,
			marshaller:                args.Marshaller,
			topic:                     args.HeartbeatTopic,
			timeBetweenSends:          args.HeartbeatTimeBetweenSends,
			timeBetweenSendsWhenError: args.HeartbeatTimeBetweenSendsWhenError,
			thresholdBetweenSends:     args.HeartbeatTimeThresholdBetweenSends,
			privKey:                   args.PrivateKey,
			redundancyHandler:         args.RedundancyHandler,
		},
		baseVersionNumber:          args.BaseVersionNumber,
		versionNumber:              args.VersionNumber,
		nodeDisplayName:            args.NodeDisplayName,
		identity:                   args.Identity,
		peerSubType:                args.PeerSubType,
		currentBlockProvider:       args.CurrentBlockProvider,
		peerTypeProvider:           args.PeerTypeProvider,
		managedPeersHolder:         args.ManagedPeersHolder,
		shardCoordinator:           args.ShardCoordinator,
		nodesCoordinator:           args.NodesCoordinator,
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
	basePeerAuthSenderArgs := argBaseSender{
		mainMessenger:             args.MainMessenger,
		fullArchiveMessenger:      args.FullArchiveMessenger,
		marshaller:                args.Marshaller,
		topic:                     args.PeerAuthenticationTopic,
		timeBetweenSends:          args.PeerAuthenticationTimeBetweenSends,
		timeBetweenSendsWhenError: args.PeerAuthenticationTimeBetweenSendsWhenError,
		thresholdBetweenSends:     args.PeerAuthenticationTimeThresholdBetweenSends,
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
		managedPeersHolder:       args.ManagedPeersHolder,
		timeBetweenChecks:        args.PeerAuthenticationTimeBetweenChecks,
		shardCoordinator:         args.ShardCoordinator,
	}
	err = checkMultikeyPeerAuthenticationSenderArgs(mpasArgs)
	if err != nil {
		return err
	}

	hbsArgs := argHeartbeatSender{
		argBaseSender: argBaseSender{
			mainMessenger:             args.MainMessenger,
			fullArchiveMessenger:      args.FullArchiveMessenger,
			marshaller:                args.Marshaller,
			topic:                     args.HeartbeatTopic,
			timeBetweenSends:          args.HeartbeatTimeBetweenSends,
			timeBetweenSendsWhenError: args.HeartbeatTimeBetweenSendsWhenError,
			thresholdBetweenSends:     args.HeartbeatTimeThresholdBetweenSends,
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
	err = checkHeartbeatSenderArgs(hbsArgs)
	if err != nil {
		return err
	}

	mhbsArgs := argMultikeyHeartbeatSender{
		argBaseSender: argBaseSender{
			mainMessenger:             args.MainMessenger,
			fullArchiveMessenger:      args.FullArchiveMessenger,
			marshaller:                args.Marshaller,
			topic:                     args.HeartbeatTopic,
			timeBetweenSends:          args.HeartbeatTimeBetweenSends,
			timeBetweenSendsWhenError: args.HeartbeatTimeBetweenSendsWhenError,
			thresholdBetweenSends:     args.HeartbeatTimeThresholdBetweenSends,
			privKey:                   args.PrivateKey,
			redundancyHandler:         args.RedundancyHandler,
		},
		peerTypeProvider:           args.PeerTypeProvider,
		versionNumber:              args.VersionNumber,
		baseVersionNumber:          args.BaseVersionNumber,
		nodeDisplayName:            args.NodeDisplayName,
		identity:                   args.Identity,
		peerSubType:                args.PeerSubType,
		currentBlockProvider:       args.CurrentBlockProvider,
		managedPeersHolder:         args.ManagedPeersHolder,
		shardCoordinator:           args.ShardCoordinator,
		trieSyncStatisticsProvider: disabled.NewTrieSyncStatisticsProvider(),
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
