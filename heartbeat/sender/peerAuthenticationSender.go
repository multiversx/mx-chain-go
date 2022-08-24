package sender

import (
	"bytes"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

// argPeerAuthenticationSender represents the arguments for the peer authentication sender
type argPeerAuthenticationSender struct {
	argBaseSender
	nodesCoordinator         heartbeat.NodesCoordinator
	peerSignatureHandler     crypto.PeerSignatureHandler
	privKey                  crypto.PrivateKey
	redundancyHandler        heartbeat.NodeRedundancyHandler
	hardforkTrigger          heartbeat.HardforkTrigger
	hardforkTimeBetweenSends time.Duration
	hardforkTriggerPubKey    []byte
}

type peerAuthenticationSender struct {
	baseSender
	nodesCoordinator         heartbeat.NodesCoordinator
	peerSignatureHandler     crypto.PeerSignatureHandler
	redundancy               heartbeat.NodeRedundancyHandler
	privKey                  crypto.PrivateKey
	publicKey                crypto.PublicKey
	observerPublicKey        crypto.PublicKey
	hardforkTrigger          heartbeat.HardforkTrigger
	hardforkTimeBetweenSends time.Duration
	hardforkTriggerPubKey    []byte
}

// newPeerAuthenticationSender will create a new instance of type peerAuthenticationSender
func newPeerAuthenticationSender(args argPeerAuthenticationSender) (*peerAuthenticationSender, error) {
	err := checkPeerAuthenticationSenderArgs(args)
	if err != nil {
		return nil, err
	}

	redundancyHandler := args.redundancyHandler
	sender := &peerAuthenticationSender{
		baseSender:               createBaseSender(args.argBaseSender),
		nodesCoordinator:         args.nodesCoordinator,
		peerSignatureHandler:     args.peerSignatureHandler,
		redundancy:               redundancyHandler,
		privKey:                  args.privKey,
		publicKey:                args.privKey.GeneratePublic(),
		observerPublicKey:        redundancyHandler.ObserverPrivateKey().GeneratePublic(),
		hardforkTrigger:          args.hardforkTrigger,
		hardforkTimeBetweenSends: args.hardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.hardforkTriggerPubKey,
	}

	return sender, nil
}

func checkPeerAuthenticationSenderArgs(args argPeerAuthenticationSender) error {
	err := checkBaseSenderArgs(args.argBaseSender)
	if err != nil {
		return err
	}
	if check.IfNil(args.nodesCoordinator) {
		return heartbeat.ErrNilNodesCoordinator
	}
	if check.IfNil(args.peerSignatureHandler) {
		return heartbeat.ErrNilPeerSignatureHandler
	}
	if check.IfNil(args.privKey) {
		return heartbeat.ErrNilPrivateKey
	}
	if check.IfNil(args.redundancyHandler) {
		return heartbeat.ErrNilRedundancyHandler
	}
	if check.IfNil(args.hardforkTrigger) {
		return heartbeat.ErrNilHardforkTrigger
	}
	if args.hardforkTimeBetweenSends < minTimeBetweenSends {
		return fmt.Errorf("%w for hardforkTimeBetweenSends", heartbeat.ErrInvalidTimeDuration)
	}
	if len(args.hardforkTriggerPubKey) == 0 {
		return fmt.Errorf("%w hardfork trigger public key bytes length is 0", heartbeat.ErrInvalidValue)
	}

	return nil
}

// Execute will handle the execution of a cycle in which the peer authentication message will be sent
func (sender *peerAuthenticationSender) Execute() {
	var duration time.Duration
	defer func() {
		sender.CreateNewTimer(duration)
	}()

	_, pk := sender.getCurrentPrivateAndPublicKeys()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		duration = sender.timeBetweenSendsWhenError
		return
	}

	if !sender.isValidator(pkBytes) && !sender.isHardforkSource(pkBytes) {
		duration = sender.timeBetweenSendsWhenError
		return
	}

	duration = sender.computeRandomDuration(sender.timeBetweenSends)
	err, isHardforkTriggered := sender.execute()
	if err != nil {
		duration = sender.timeBetweenSendsWhenError
		log.Error("error sending peer authentication message", "error", err, "is hardfork triggered", isHardforkTriggered, "next send will be in", duration)
		return
	}

	if isHardforkTriggered {
		duration = sender.computeRandomDuration(sender.hardforkTimeBetweenSends)
	}

	log.Debug("peer authentication message sent", "is hardfork triggered", isHardforkTriggered, "next send will be in", duration)
}

func (sender *peerAuthenticationSender) execute() (error, bool) {
	sk, pk := sender.getCurrentPrivateAndPublicKeys()

	msg := &heartbeat.PeerAuthentication{
		Pid: sender.messenger.ID().Bytes(),
	}

	hardforkPayload, isTriggered := sender.getHardforkPayload()
	payload := &heartbeat.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: string(hardforkPayload),
	}
	payloadBytes, err := sender.marshaller.Marshal(payload)
	if err != nil {
		return err, isTriggered
	}
	msg.Payload = payloadBytes
	msg.PayloadSignature, err = sender.messenger.Sign(payloadBytes)
	if err != nil {
		return err, isTriggered
	}

	msg.Pubkey, err = pk.ToByteArray()
	if err != nil {
		return err, isTriggered
	}

	msg.Signature, err = sender.peerSignatureHandler.GetPeerSignature(sk, msg.Pid)
	if err != nil {
		return err, isTriggered
	}

	msgBytes, err := sender.marshaller.Marshal(msg)
	if err != nil {
		return err, isTriggered
	}

	b := &batch.Batch{
		Data: make([][]byte, 1),
	}
	b.Data[0] = msgBytes
	data, err := sender.marshaller.Marshal(b)
	if err != nil {
		return err, isTriggered
	}

	sender.messenger.Broadcast(sender.topic, data)

	return nil, isTriggered
}

// ShouldTriggerHardfork signals when hardfork message should be sent
func (sender *peerAuthenticationSender) ShouldTriggerHardfork() <-chan struct{} {
	return sender.hardforkTrigger.NotifyTriggerReceivedV2()
}

func (sender *peerAuthenticationSender) getCurrentPrivateAndPublicKeys() (crypto.PrivateKey, crypto.PublicKey) {
	shouldUseOriginalKeys := !sender.redundancy.IsRedundancyNode() || (sender.redundancy.IsRedundancyNode() && !sender.redundancy.IsMainMachineActive())
	if shouldUseOriginalKeys {
		return sender.privKey, sender.publicKey
	}

	return sender.redundancy.ObserverPrivateKey(), sender.observerPublicKey
}

func (sender *peerAuthenticationSender) isValidator(pkBytes []byte) bool {
	_, _, err := sender.nodesCoordinator.GetValidatorWithPublicKey(pkBytes)
	return err == nil
}

func (sender *peerAuthenticationSender) isHardforkSource(pkBytes []byte) bool {
	return bytes.Equal(pkBytes, sender.hardforkTriggerPubKey)
}

func (sender *peerAuthenticationSender) getHardforkPayload() ([]byte, bool) {
	payload := make([]byte, 0)
	_, isTriggered := sender.hardforkTrigger.RecordedTriggerMessage()
	if isTriggered {
		payload = sender.hardforkTrigger.CreateData()
	}

	return payload, isTriggered
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *peerAuthenticationSender) IsInterfaceNil() bool {
	return sender == nil
}
