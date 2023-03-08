package sender

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
)

// argPeerAuthenticationSender represents the arguments for the peer authentication sender
type argPeerAuthenticationSender struct {
	argBaseSender
	nodesCoordinator         heartbeat.NodesCoordinator
	peerSignatureHandler     crypto.PeerSignatureHandler
	hardforkTrigger          heartbeat.HardforkTrigger
	hardforkTimeBetweenSends time.Duration
	hardforkTriggerPubKey    []byte
}

type peerAuthenticationSender struct {
	commonPeerAuthenticationSender
	redundancy               heartbeat.NodeRedundancyHandler
	privKey                  crypto.PrivateKey
	publicKey                crypto.PublicKey
	observerPublicKey        crypto.PublicKey
	hardforkTimeBetweenSends time.Duration
}

// newPeerAuthenticationSender will create a new instance of type peerAuthenticationSender
func newPeerAuthenticationSender(args argPeerAuthenticationSender) (*peerAuthenticationSender, error) {
	err := checkPeerAuthenticationSenderArgs(args)
	if err != nil {
		return nil, err
	}

	redundancyHandler := args.redundancyHandler
	senderInstance := &peerAuthenticationSender{
		commonPeerAuthenticationSender: commonPeerAuthenticationSender{
			baseSender:            createBaseSender(args.argBaseSender),
			nodesCoordinator:      args.nodesCoordinator,
			peerSignatureHandler:  args.peerSignatureHandler,
			hardforkTrigger:       args.hardforkTrigger,
			hardforkTriggerPubKey: args.hardforkTriggerPubKey,
		},
		redundancy:               redundancyHandler,
		privKey:                  args.privKey,
		publicKey:                args.privKey.GeneratePublic(),
		observerPublicKey:        redundancyHandler.ObserverPrivateKey().GeneratePublic(),
		hardforkTimeBetweenSends: args.hardforkTimeBetweenSends,
	}

	return senderInstance, nil
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

	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return err, false
	}

	data, isTriggered, msgTimestamp, err := sender.generateMessageBytes(pkBytes, sk, nil, sender.messenger.ID().Bytes())
	if err != nil {
		return err, isTriggered
	}

	log.Debug("sending peer authentication message",
		"public key", pkBytes, "pid", sender.messenger.ID().Pretty(),
		"timestamp", msgTimestamp)
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

// IsInterfaceNil returns true if there is no value under the interface
func (sender *peerAuthenticationSender) IsInterfaceNil() bool {
	return sender == nil
}
