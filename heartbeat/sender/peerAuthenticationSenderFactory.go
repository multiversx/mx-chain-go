package sender

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/covalent-indexer-go/process"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

type argPeerAuthenticationSenderFactory struct {
	argBaseSender
	nodesCoordinator         heartbeat.NodesCoordinator
	peerSignatureHandler     crypto.PeerSignatureHandler
	hardforkTrigger          heartbeat.HardforkTrigger
	hardforkTimeBetweenSends time.Duration
	hardforkTriggerPubKey    []byte
	keysHolder               heartbeat.KeysHolder
	timeBetweenChecks        time.Duration
	shardCoordinator         process.ShardCoordinator
	privKey                  crypto.PrivateKey
	redundancyHandler        heartbeat.NodeRedundancyHandler
}

type peerAuthenticationSenderFactory struct {
	messenger                 heartbeat.P2PMessenger
	marshaller                marshal.Marshalizer
	topic                     string
	timeBetweenSends          time.Duration
	timeBetweenSendsWhenError time.Duration
	thresholdBetweenSends     float64
	nodesCoordinator          heartbeat.NodesCoordinator
	peerSignatureHandler      crypto.PeerSignatureHandler
	hardforkTrigger           heartbeat.HardforkTrigger
	hardforkTimeBetweenSends  time.Duration
	hardforkTriggerPubKey     []byte
	keysHolder                heartbeat.KeysHolder
	timeBetweenChecks         time.Duration
	shardCoordinator          process.ShardCoordinator
	privKey                   crypto.PrivateKey
	redundancyHandler         heartbeat.NodeRedundancyHandler
}

func newPeerAuthenticationSenderFactory(args argPeerAuthenticationSenderFactory) (*peerAuthenticationSenderFactory, error) {
	return &peerAuthenticationSenderFactory{
		messenger:                 args.messenger,
		marshaller:                args.marshaller,
		topic:                     args.topic,
		timeBetweenSends:          args.timeBetweenSends,
		timeBetweenSendsWhenError: args.timeBetweenSendsWhenError,
		thresholdBetweenSends:     args.thresholdBetweenSends,
		nodesCoordinator:          args.nodesCoordinator,
		peerSignatureHandler:      args.peerSignatureHandler,
		hardforkTrigger:           args.hardforkTrigger,
		hardforkTimeBetweenSends:  args.hardforkTimeBetweenSends,
		hardforkTriggerPubKey:     args.hardforkTriggerPubKey,
		keysHolder:                args.keysHolder,
		timeBetweenChecks:         args.timeBetweenChecks,
		shardCoordinator:          args.shardCoordinator,
		privKey:                   args.privKey,
		redundancyHandler:         args.redundancyHandler,
	}, nil
}

func (pasf *peerAuthenticationSenderFactory) create() (peerAuthenticationHandler, error) {
	isMultikey := check.IfNil(pasf.privKey)
	isRegular := check.IfNil(pasf.keysHolder)
	if isMultikey && isRegular {
		return nil, fmt.Errorf("%w for peerAuthenticationSenderFactory, received both validator key and multikey holder", heartbeat.ErrInvalidConfiguration)
	}

	if !isMultikey && !isRegular {
		return nil, fmt.Errorf("%w for peerAuthenticationSenderFactory, could not determine node's type", heartbeat.ErrInvalidConfiguration)
	}

	if isMultikey {
		return pasf.createMultikeyPeerAuthenticationSender()
	}

	return pasf.createPeerAuthenticationSender()
}

func (pasf *peerAuthenticationSenderFactory) createPeerAuthenticationSender() (*peerAuthenticationSender, error) {
	args := argPeerAuthenticationSender{
		argBaseSender: argBaseSender{
			messenger:                 pasf.messenger,
			marshaller:                pasf.marshaller,
			topic:                     pasf.topic,
			timeBetweenSends:          pasf.timeBetweenSends,
			timeBetweenSendsWhenError: pasf.timeBetweenSendsWhenError,
			thresholdBetweenSends:     pasf.thresholdBetweenSends,
		},
		nodesCoordinator:         pasf.nodesCoordinator,
		peerSignatureHandler:     pasf.peerSignatureHandler,
		privKey:                  pasf.privKey,
		redundancyHandler:        pasf.redundancyHandler,
		hardforkTrigger:          pasf.hardforkTrigger,
		hardforkTimeBetweenSends: pasf.hardforkTimeBetweenSends,
		hardforkTriggerPubKey:    pasf.hardforkTriggerPubKey,
	}

	return newPeerAuthenticationSender(args)
}

func (pasf *peerAuthenticationSenderFactory) createMultikeyPeerAuthenticationSender() (*multikeyPeerAuthenticationSender, error) {
	args := argMultikeyPeerAuthenticationSender{
		argBaseSender: argBaseSender{
			messenger:                 pasf.messenger,
			marshaller:                pasf.marshaller,
			topic:                     pasf.topic,
			timeBetweenSends:          pasf.timeBetweenSends,
			timeBetweenSendsWhenError: pasf.timeBetweenSendsWhenError,
			thresholdBetweenSends:     pasf.thresholdBetweenSends,
		},
		nodesCoordinator:         pasf.nodesCoordinator,
		peerSignatureHandler:     pasf.peerSignatureHandler,
		hardforkTrigger:          pasf.hardforkTrigger,
		hardforkTimeBetweenSends: pasf.hardforkTimeBetweenSends,
		hardforkTriggerPubKey:    pasf.hardforkTriggerPubKey,
		keysHolder:               pasf.keysHolder,
		timeBetweenChecks:        pasf.timeBetweenChecks,
		shardCoordinator:         pasf.shardCoordinator,
	}

	return newMultikeyPeerAuthenticationSender(args)
}
