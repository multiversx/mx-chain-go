package sender

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
)

// argPeerAuthenticationSender represents the arguments for the peer authentication sender
type argPeerAuthenticationSender struct {
	argBaseSender
	nodesCoordinator     heartbeat.NodesCoordinator
	peerSignatureHandler crypto.PeerSignatureHandler
}

type peerAuthenticationSender struct {
	commonPeerAuthenticationSender
	redundancy        heartbeat.NodeRedundancyHandler
	privKey           crypto.PrivateKey
	publicKey         crypto.PublicKey
	observerPublicKey crypto.PublicKey
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
			baseSender:           createBaseSender(args.argBaseSender),
			nodesCoordinator:     args.nodesCoordinator,
			peerSignatureHandler: args.peerSignatureHandler,
		},
		redundancy:        redundancyHandler,
		privKey:           args.privKey,
		publicKey:         args.privKey.GeneratePublic(),
		observerPublicKey: redundancyHandler.ObserverPrivateKey().GeneratePublic(),
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

	if !sender.isValidator(pkBytes) {
		duration = sender.timeBetweenSendsWhenError
		return
	}

	duration = sender.computeRandomDuration(sender.timeBetweenSends)
	err = sender.execute()
	if err != nil {
		duration = sender.timeBetweenSendsWhenError
		log.Error("error sending peer authentication message", "error", err, "next send will be in", duration)
		return
	}

	log.Debug("peer authentication message sent", "next send will be in", duration)
}

func (sender *peerAuthenticationSender) execute() error {
	sk, pk := sender.getCurrentPrivateAndPublicKeys()

	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return err
	}

	data, msgTimestamp, err := sender.generateMessageBytes(pkBytes, sk, nil, sender.mainMessenger.ID().Bytes())
	if err != nil {
		return err
	}

	log.Debug("sending peer authentication message",
		"public key", pkBytes, "pid", sender.mainMessenger.ID().Pretty(),
		"timestamp", msgTimestamp)
	sender.mainMessenger.Broadcast(sender.topic, data)

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
