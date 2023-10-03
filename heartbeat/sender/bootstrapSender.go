package sender

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/heartbeat/sender/disabled"
)

// ArgBootstrapSender represents the arguments for the bootstrap bootstrapSender
type ArgBootstrapSender struct {
	MainMessenger                      heartbeat.P2PMessenger
	FullArchiveMessenger               heartbeat.P2PMessenger
	Marshaller                         marshal.Marshalizer
	HeartbeatTopic                     string
	HeartbeatTimeBetweenSends          time.Duration
	HeartbeatTimeBetweenSendsWhenError time.Duration
	HeartbeatTimeThresholdBetweenSends float64
	VersionNumber                      string
	NodeDisplayName                    string
	Identity                           string
	PeerSubType                        core.P2PPeerSubType
	CurrentBlockProvider               heartbeat.CurrentBlockProvider
	PrivateKey                         crypto.PrivateKey
	RedundancyHandler                  heartbeat.NodeRedundancyHandler
	PeerTypeProvider                   heartbeat.PeerTypeProviderHandler
	TrieSyncStatisticsProvider         heartbeat.TrieSyncStatisticsProvider
}

// bootstrapSender defines the component which sends heartbeat messages during bootstrap
type bootstrapSender struct {
	heartbeatSender *heartbeatSender
	routineHandler  *routineHandler
}

// NewBootstrapSender creates a new instance of bootstrapSender
func NewBootstrapSender(args ArgBootstrapSender) (*bootstrapSender, error) {
	hbs, err := newHeartbeatSender(argHeartbeatSender{
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
		trieSyncStatisticsProvider: args.TrieSyncStatisticsProvider,
	})
	if err != nil {
		return nil, err
	}

	return &bootstrapSender{
		heartbeatSender: hbs,
		routineHandler:  newRoutineHandler(disabled.NewDisabledSenderHandler(), hbs, disabled.NewDisabledHardforkHandler()),
	}, nil
}

// Close closes the internal components
func (sender *bootstrapSender) Close() error {
	sender.routineHandler.closeProcessLoop()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *bootstrapSender) IsInterfaceNil() bool {
	return sender == nil
}
