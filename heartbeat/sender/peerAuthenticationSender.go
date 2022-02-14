package sender

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

// ArgPeerAuthenticationSender represents the arguments for the peer authentication sender
type ArgPeerAuthenticationSender struct {
	ArgBaseSender
	PeerSignatureHandler crypto.PeerSignatureHandler
	PrivKey              crypto.PrivateKey
	RedundancyHandler    heartbeat.NodeRedundancyHandler
}

type peerAuthenticationSender struct {
	baseSender
	peerSignatureHandler crypto.PeerSignatureHandler
	redundancy           heartbeat.NodeRedundancyHandler
	privKey              crypto.PrivateKey
	publicKey            crypto.PublicKey
	observerPublicKey    crypto.PublicKey
}

// NewPeerAuthenticationSender will create a new instance of type peerAuthenticationSender
func NewPeerAuthenticationSender(args ArgPeerAuthenticationSender) (*peerAuthenticationSender, error) {
	err := checkPeerAuthenticationSenderArgs(args)
	if err != nil {
		return nil, err
	}

	redundancyHandler := args.RedundancyHandler
	sender := &peerAuthenticationSender{
		baseSender:           createBaseSender(args.ArgBaseSender),
		peerSignatureHandler: args.PeerSignatureHandler,
		redundancy:           redundancyHandler,
		privKey:              args.PrivKey,
		publicKey:            args.PrivKey.GeneratePublic(),
		observerPublicKey:    redundancyHandler.ObserverPrivateKey().GeneratePublic(),
	}

	return sender, nil
}

func checkPeerAuthenticationSenderArgs(args ArgPeerAuthenticationSender) error {
	err := checkBaseSenderArgs(args.ArgBaseSender)
	if err != nil {
		return err
	}
	if check.IfNil(args.PeerSignatureHandler) {
		return heartbeat.ErrNilPeerSignatureHandler
	}
	if check.IfNil(args.PrivKey) {
		return heartbeat.ErrNilPrivateKey
	}
	if check.IfNil(args.RedundancyHandler) {
		return heartbeat.ErrNilRedundancyHandler
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

// IsInterfaceNil returns true if there is no value under the interface
func (sender *peerAuthenticationSender) IsInterfaceNil() bool {
	return sender == nil
}
