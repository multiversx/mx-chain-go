package sender

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/heartbeat"
)

const maxSizeInBytes = 128

// argHeartbeatSender represents the arguments for the heartbeat sender
type argHeartbeatSender struct {
	argBaseSender
	versionNumber              string
	nodeDisplayName            string
	identity                   string
	peerSubType                core.P2PPeerSubType
	currentBlockProvider       heartbeat.CurrentBlockProvider
	peerTypeProvider           heartbeat.PeerTypeProviderHandler
	trieSyncStatisticsProvider heartbeat.TrieSyncStatisticsProvider
}

type heartbeatSender struct {
	commonHeartbeatSender
	trieSyncStatisticsProvider heartbeat.TrieSyncStatisticsProvider
}

// newHeartbeatSender creates a new instance of type heartbeatSender
func newHeartbeatSender(args argHeartbeatSender) (*heartbeatSender, error) {
	err := checkHeartbeatSenderArgs(args)
	if err != nil {
		return nil, err
	}

	return &heartbeatSender{
		commonHeartbeatSender: commonHeartbeatSender{
			baseSender:           createBaseSender(args.argBaseSender),
			currentBlockProvider: args.currentBlockProvider,
			peerTypeProvider:     args.peerTypeProvider,
			versionNumber:        args.versionNumber,
			nodeDisplayName:      args.nodeDisplayName,
			identity:             args.identity,
			peerSubType:          args.peerSubType,
		},
		trieSyncStatisticsProvider: args.trieSyncStatisticsProvider,
	}, nil
}

func checkHeartbeatSenderArgs(args argHeartbeatSender) error {
	err := checkBaseSenderArgs(args.argBaseSender)
	if err != nil {
		return err
	}
	if len(args.versionNumber) > maxSizeInBytes {
		return fmt.Errorf("%w for versionNumber, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, args.versionNumber, len(args.versionNumber), maxSizeInBytes)
	}
	if len(args.nodeDisplayName) > maxSizeInBytes {
		return fmt.Errorf("%w for nodeDisplayName, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, args.nodeDisplayName, len(args.nodeDisplayName), maxSizeInBytes)
	}
	if len(args.identity) > maxSizeInBytes {
		return fmt.Errorf("%w for identity, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, args.identity, len(args.identity), maxSizeInBytes)
	}
	if check.IfNil(args.currentBlockProvider) {
		return heartbeat.ErrNilCurrentBlockProvider
	}
	if check.IfNil(args.peerTypeProvider) {
		return heartbeat.ErrNilPeerTypeProvider
	}
	if check.IfNil(args.trieSyncStatisticsProvider) {
		return heartbeat.ErrNilTrieSyncStatisticsProvider
	}

	return nil
}

// Execute will handle the execution of a cycle in which the heartbeat message will be sent
func (sender *heartbeatSender) Execute() {
	duration := sender.computeRandomDuration(sender.timeBetweenSends)
	err := sender.execute()
	if err != nil {
		duration = sender.timeBetweenSendsWhenError
		log.Error("error sending heartbeat message", "error", err, "next send will be in", duration)
	} else {
		log.Debug("heartbeat message sent", "next send will be in", duration)
	}

	sender.CreateNewTimer(duration)
}

func (sender *heartbeatSender) execute() error {
	_, pk := sender.getCurrentPrivateAndPublicKeys()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return err
	}

	trieNodesReceived := uint64(sender.trieSyncStatisticsProvider.NumProcessed())
	msgBytes, err := sender.generateMessageBytes(
		sender.versionNumber,
		sender.nodeDisplayName,
		sender.identity,
		uint32(sender.peerSubType),
		pkBytes,
		trieNodesReceived,
	)
	if err != nil {
		return err
	}

	sender.mainMessenger.Broadcast(sender.topic, msgBytes)
	sender.fullArchiveMessenger.Broadcast(sender.topic, msgBytes)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *heartbeatSender) IsInterfaceNil() bool {
	return sender == nil
}
