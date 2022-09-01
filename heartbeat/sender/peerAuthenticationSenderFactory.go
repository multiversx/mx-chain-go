package sender

import (
	"time"

	"github.com/ElrondNetwork/covalent-indexer-go/process"
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

func createPeerAuthenticationSender(args argPeerAuthenticationSenderFactory) (peerAuthenticationSenderHandler, error) {
	pk := args.privKey.GeneratePublic()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return nil, err
	}

	_, _, err = args.nodesCoordinator.GetValidatorWithPublicKey(pkBytes)
	if err == nil {
		return createRegularSender(args)
	}

	keysMap := args.keysHolder.GetManagedKeysByCurrentNode()
	if len(keysMap) == 0 {
		return createRegularSender(args)
	}

	return createMultikeySender(args)
}

func createRegularSender(args argPeerAuthenticationSenderFactory) (*peerAuthenticationSender, error) {
	argsSender := argPeerAuthenticationSender{
		argBaseSender:            args.argBaseSender,
		nodesCoordinator:         args.nodesCoordinator,
		peerSignatureHandler:     args.peerSignatureHandler,
		privKey:                  args.privKey,
		redundancyHandler:        args.redundancyHandler,
		hardforkTrigger:          args.hardforkTrigger,
		hardforkTimeBetweenSends: args.hardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.hardforkTriggerPubKey,
	}

	return newPeerAuthenticationSender(argsSender)
}

func createMultikeySender(args argPeerAuthenticationSenderFactory) (*multikeyPeerAuthenticationSender, error) {
	argsSender := argMultikeyPeerAuthenticationSender{
		argBaseSender:            args.argBaseSender,
		nodesCoordinator:         args.nodesCoordinator,
		peerSignatureHandler:     args.peerSignatureHandler,
		hardforkTrigger:          args.hardforkTrigger,
		hardforkTimeBetweenSends: args.hardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.hardforkTriggerPubKey,
		keysHolder:               args.keysHolder,
		timeBetweenChecks:        args.timeBetweenChecks,
		shardCoordinator:         args.shardCoordinator,
	}

	return newMultikeyPeerAuthenticationSender(argsSender)
}
