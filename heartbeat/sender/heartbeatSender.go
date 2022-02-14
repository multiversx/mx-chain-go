package sender

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

// ArgHeartbeatSender represents the arguments for the heartbeat sender
type ArgHeartbeatSender struct {
	ArgBaseSender
	VersionNumber        string
	NodeDisplayName      string
	Identity             string
	PeerSubType          core.P2PPeerSubType
	CurrentBlockProvider heartbeat.CurrentBlockProvider
}

type heartbeatSender struct {
	baseSender
	versionNumber        string
	nodeDisplayName      string
	identity             string
	peerSubType          core.P2PPeerSubType
	currentBlockProvider heartbeat.CurrentBlockProvider
}

// NewHeartbeatSender creates a new instance of type heartbeatSender
func NewHeartbeatSender(args ArgHeartbeatSender) (*heartbeatSender, error) {
	err := checkHeartbeatSenderArg(args)
	if err != nil {
		return nil, err
	}

	return &heartbeatSender{
		baseSender:           createBaseSender(args.ArgBaseSender),
		versionNumber:        args.VersionNumber,
		nodeDisplayName:      args.NodeDisplayName,
		identity:             args.Identity,
		currentBlockProvider: args.CurrentBlockProvider,
		peerSubType:          args.PeerSubType,
	}, nil
}

func checkHeartbeatSenderArg(args ArgHeartbeatSender) error {
	err := checkBaseSenderArgs(args.ArgBaseSender)
	if err != nil {
		return err
	}
	if len(args.VersionNumber) == 0 {
		return heartbeat.ErrEmptyVersionNumber
	}
	if len(args.NodeDisplayName) == 0 {
		return heartbeat.ErrEmptyNodeDisplayName
	}
	if len(args.Identity) == 0 {
		return heartbeat.ErrEmptyIdentity
	}
	if check.IfNil(args.CurrentBlockProvider) {
		return heartbeat.ErrNilCurrentBlockProvider
	}

	return nil
}

// Execute will handle the execution of a cycle in which the heartbeat message will be sent
func (sender *heartbeatSender) Execute() {
	duration := sender.timeBetweenSends
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

	msg := heartbeat.HeartbeatV2{
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
