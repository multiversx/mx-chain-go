package sender

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

const minTimeBetweenSends = time.Second

// ArgPeerAuthenticationSender represents the arguments for the peer authentication sender
type ArgPeerAuthenticationSender struct {
	Messenger                 heartbeat.P2PMessenger
	PeerSignatureHandler      crypto.PeerSignatureHandler
	PrivKey                   crypto.PrivateKey
	Marshaller                marshal.Marshalizer
	Topic                     string
	RedundancyHandler         heartbeat.NodeRedundancyHandler
	TimeBetweenSends          time.Duration
	TimeBetweenSendsWhenError time.Duration
}

type peerAuthenticationSender struct {
	timerHandler
	messenger                 heartbeat.P2PMessenger
	peerSignatureHandler      crypto.PeerSignatureHandler
	redundancy                heartbeat.NodeRedundancyHandler
	privKey                   crypto.PrivateKey
	publicKey                 crypto.PublicKey
	observerPublicKey         crypto.PublicKey
	marshaller                marshal.Marshalizer
	topic                     string
	timeBetweenSends          time.Duration
	timeBetweenSendsWhenError time.Duration
}

// NewPeerAuthenticationSender will create a new instance of type peerAuthenticationSender
func NewPeerAuthenticationSender(args ArgPeerAuthenticationSender) (*peerAuthenticationSender, error) {
	err := checkPeerAuthenticationSenderArgs(args)
	if err != nil {
		return nil, err
	}

	redundancyHandler := args.RedundancyHandler
	sender := &peerAuthenticationSender{
		timerHandler: &timerWrapper{
			timer: time.NewTimer(args.TimeBetweenSends),
		},
		messenger:                 args.Messenger,
		peerSignatureHandler:      args.PeerSignatureHandler,
		redundancy:                redundancyHandler,
		privKey:                   args.PrivKey,
		publicKey:                 args.PrivKey.GeneratePublic(),
		observerPublicKey:         redundancyHandler.ObserverPrivateKey().GeneratePublic(),
		marshaller:                args.Marshaller,
		topic:                     args.Topic,
		timeBetweenSends:          args.TimeBetweenSends,
		timeBetweenSendsWhenError: args.TimeBetweenSendsWhenError,
	}

	return sender, nil
}

func checkPeerAuthenticationSenderArgs(args ArgPeerAuthenticationSender) error {
	if check.IfNil(args.Messenger) {
		return heartbeat.ErrNilMessenger
	}
	if check.IfNil(args.PeerSignatureHandler) {
		return heartbeat.ErrNilPeerSignatureHandler
	}
	if check.IfNil(args.PrivKey) {
		return heartbeat.ErrNilPrivateKey
	}
	if check.IfNil(args.Marshaller) {
		return heartbeat.ErrNilMarshaller
	}
	if len(args.Topic) == 0 {
		return heartbeat.ErrEmptySendTopic
	}
	if check.IfNil(args.RedundancyHandler) {
		return heartbeat.ErrNilRedundancyHandler
	}
	if args.TimeBetweenSends < minTimeBetweenSends {
		return fmt.Errorf("%w for TimeBetweenSends", heartbeat.ErrInvalidTimeDuration)
	}
	if args.TimeBetweenSendsWhenError < minTimeBetweenSends {
		return fmt.Errorf("%w for TimeBetweenSendsWhenError", heartbeat.ErrInvalidTimeDuration)
	}

	return nil
}

// Execute will handle the execution of a cycle in which the peer authentication message will be sent
func (sender *peerAuthenticationSender) Execute() {
	duration := sender.timeBetweenSends
	err := sender.execute()
	if err != nil {
		duration = sender.timeBetweenSendsWhenError
		log.Error("error sending peer authentication message", "error", err, "next send will be in", duration)
	} else {
		log.Debug("peer authentication message sent", "next send will be in", duration)
	}

	sender.CreateNewTimer(duration)
}

func (sender *peerAuthenticationSender) execute() error {
	sk, pk := sender.getCurrentPrivateAndPublicKeys()

	msg := &heartbeat.PeerAuthentication{
		Pid: sender.messenger.ID().Bytes(),
	}
	payload := &heartbeat.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: "", // TODO add the hardfork message, if required
	}
	payloadBytes, err := sender.marshaller.Marshal(payload)
	if err != nil {
		return err
	}
	msg.Payload = payloadBytes
	msg.PayloadSignature, err = sender.messenger.Sign(payloadBytes)
	if err != nil {
		return err
	}

	msg.Pubkey, err = pk.ToByteArray()
	if err != nil {
		return err
	}

	msg.Signature, err = sender.peerSignatureHandler.GetPeerSignature(sk, msg.Pid)
	if err != nil {
		return err
	}

	msgBytes, err := sender.marshaller.Marshal(msg)
	if err != nil {
		return err
	}

	sender.messenger.Broadcast(sender.topic, msgBytes)

	return nil
}

func (sender *peerAuthenticationSender) getCurrentPrivateAndPublicKeys() (crypto.PrivateKey, crypto.PublicKey) {
	shouldUseOriginalKeys := !sender.redundancy.IsRedundancyNode() || (sender.redundancy.IsRedundancyNode() && !sender.redundancy.IsMainMachineActive())
	if shouldUseOriginalKeys {
		return sender.privKey, sender.publicKey
	}

	return sender.redundancy.ObserverPrivateKey(), sender.observerPublicKey
}
