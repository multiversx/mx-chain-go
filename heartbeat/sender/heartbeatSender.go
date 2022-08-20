package sender

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

const maxSizeInBytes = 128

// argHeartbeatSender represents the arguments for the heartbeat sender
type argHeartbeatSender struct {
	argBaseSender
	versionNumber        string
	nodeDisplayName      string
	identity             string
	peerSubType          core.P2PPeerSubType
	currentBlockProvider heartbeat.CurrentBlockProvider
}

type heartbeatSender struct {
	baseSender
	versionNumber        string
	nodeDisplayName      string
	identity             string
	peerSubType          core.P2PPeerSubType
	currentBlockProvider heartbeat.CurrentBlockProvider
}

// newHeartbeatSender creates a new instance of type heartbeatSender
func newHeartbeatSender(args argHeartbeatSender) (*heartbeatSender, error) {
	err := checkHeartbeatSenderArgs(args)
	if err != nil {
		return nil, err
	}

	return &heartbeatSender{
		baseSender:           createBaseSender(args.argBaseSender),
		versionNumber:        args.versionNumber,
		nodeDisplayName:      args.nodeDisplayName,
		identity:             args.identity,
		currentBlockProvider: args.currentBlockProvider,
		peerSubType:          args.peerSubType,
	}, nil
}

func checkHeartbeatSenderArgs(args argHeartbeatSender) error {
	err := checkBaseSenderArgs(args.argBaseSender)
	if err != nil {
		return err
	}
	if len(args.versionNumber) > maxSizeInBytes {
		return heartbeat.ErrPropertyTooLong
	}
	if check.IfNil(args.currentBlockProvider) {
		return heartbeat.ErrNilCurrentBlockProvider
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
	payload := &heartbeat.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: "", // sent through peer authentication message
	}
	payloadBytes, err := sender.marshaller.Marshal(payload)
	if err != nil {
		return err
	}

	nonce := uint64(0)
	currentBlock := sender.currentBlockProvider.GetCurrentBlockHeader()
	if currentBlock != nil {
		nonce = currentBlock.GetNonce()
	}

	msg := &heartbeat.HeartbeatV2{
		Payload:         payloadBytes,
		VersionNumber:   sender.versionNumber,
		NodeDisplayName: sender.nodeDisplayName,
		Identity:        sender.identity,
		Nonce:           nonce,
		PeerSubType:     uint32(sender.peerSubType),
	}

	msgBytes, err := sender.marshaller.Marshal(msg)
	if err != nil {
		return err
	}

	sender.messenger.Broadcast(sender.topic, msgBytes)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *heartbeatSender) IsInterfaceNil() bool {
	return sender == nil
}
