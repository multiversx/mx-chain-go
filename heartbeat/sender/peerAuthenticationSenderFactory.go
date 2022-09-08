package sender

import (
	"fmt"
	"time"

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
	shardCoordinator         heartbeat.ShardCoordinator
}

func createPeerAuthenticationSender(args argPeerAuthenticationSenderFactory) (peerAuthenticationSenderHandler, error) {
	pk := args.privKey.GeneratePublic()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return nil, err
	}

	keysMap := args.keysHolder.GetManagedKeysByCurrentNode()
	isMultikeyMode := len(keysMap) > 0
	_, _, err = args.nodesCoordinator.GetValidatorWithPublicKey(pkBytes)
	if err == nil {
		if isMultikeyMode {
			return nil, fmt.Errorf("%w while creating peer authentication, could not determine node's type", heartbeat.ErrInvalidConfiguration)
		}

		return createRegularSender(args)
	}

	if !isMultikeyMode {
		return createRegularSender(args)
	}

	return createMultikeySender(args)
}

func createRegularSender(args argPeerAuthenticationSenderFactory) (*peerAuthenticationSender, error) {
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

func createMultikeySender(args argPeerAuthenticationSenderFactory) (*multikeyPeerAuthenticationSender, error) {
	argsSender := argMultikeyPeerAuthenticationSender(args)
	return newMultikeyPeerAuthenticationSender(argsSender)
}
