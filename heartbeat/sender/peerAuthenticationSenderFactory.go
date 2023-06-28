package sender

import (
	"fmt"
	"time"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
)

type argPeerAuthenticationSenderFactory struct {
	argBaseSender
	nodesCoordinator         heartbeat.NodesCoordinator
	peerSignatureHandler     crypto.PeerSignatureHandler
	hardforkTrigger          heartbeat.HardforkTrigger
	hardforkTimeBetweenSends time.Duration
	hardforkTriggerPubKey    []byte
	managedPeersHolder       heartbeat.ManagedPeersHolder
	timeBetweenChecks        time.Duration
	shardCoordinator         heartbeat.ShardCoordinator
}

func createPeerAuthenticationSender(args argPeerAuthenticationSenderFactory) (peerAuthenticationSenderHandler, error) {
	isMultikey, err := isMultikeyMode(args.privKey, args.managedPeersHolder, args.nodesCoordinator)
	if err != nil {
		return nil, fmt.Errorf("%w while creating peer authentication sender", err)
	}

	if isMultikey {
		return createMultikeyPeerAuthenticationSender(args)
	}

	return createRegularPeerAuthenticationSender(args)
}

func createRegularPeerAuthenticationSender(args argPeerAuthenticationSenderFactory) (*peerAuthenticationSender, error) {
	argsSender := argPeerAuthenticationSender{
		argBaseSender:            args.argBaseSender,
		nodesCoordinator:         args.nodesCoordinator,
		peerSignatureHandler:     args.peerSignatureHandler,
		hardforkTrigger:          args.hardforkTrigger,
		hardforkTimeBetweenSends: args.hardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.hardforkTriggerPubKey,
	}

	return newPeerAuthenticationSender(argsSender)
}

func createMultikeyPeerAuthenticationSender(args argPeerAuthenticationSenderFactory) (*multikeyPeerAuthenticationSender, error) {
	argsSender := argMultikeyPeerAuthenticationSender(args)
	return newMultikeyPeerAuthenticationSender(argsSender)
}
