package sender

import (
	"fmt"

	"github.com/ElrondNetwork/covalent-indexer-go/process"
	"github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

type argHeartbeatSenderFactory struct {
	argBaseSender
	baseVersionNumber    string
	versionNumber        string
	nodeDisplayName      string
	identity             string
	peerSubType          core.P2PPeerSubType
	currentBlockProvider heartbeat.CurrentBlockProvider
	peerTypeProvider     heartbeat.PeerTypeProviderHandler
	keysHolder           heartbeat.KeysHolder
	shardCoordinator     process.ShardCoordinator
	nodesCoordinator     heartbeat.NodesCoordinator
}

func createHeartbeatSender(args argHeartbeatSenderFactory) (senderHandler, error) {
	isMultikey, err := isMultikeyMode(args.privKey, args.keysHolder, args.nodesCoordinator)
	if err != nil {
		return nil, fmt.Errorf("%w while creating heartbeat sender", err)
	}

	if isMultikey {
		return createMultikeyHeartbeatSender(args)
	}

	return createRegularHeartbeatSender(args)
}

func createRegularHeartbeatSender(args argHeartbeatSenderFactory) (*heartbeatSender, error) {
	argsSender := argHeartbeatSender{
		argBaseSender: argBaseSender{
			messenger:                 args.messenger,
			marshaller:                args.marshaller,
			topic:                     args.topic,
			timeBetweenSends:          args.timeBetweenSends,
			timeBetweenSendsWhenError: args.timeBetweenSendsWhenError,
			thresholdBetweenSends:     args.thresholdBetweenSends,
			redundancyHandler:         args.redundancyHandler,
			privKey:                   args.privKey,
		},
		versionNumber:        args.versionNumber,
		nodeDisplayName:      args.nodeDisplayName,
		identity:             args.identity,
		peerSubType:          args.peerSubType,
		currentBlockProvider: args.currentBlockProvider,
	}

	return newHeartbeatSender(argsSender)
}

func createMultikeyHeartbeatSender(args argHeartbeatSenderFactory) (*multikeyHeartbeatSender, error) {
	argsSender := argMultikeyHeartbeatSender{
		argBaseSender: argBaseSender{
			messenger:                 args.messenger,
			marshaller:                args.marshaller,
			topic:                     args.topic,
			timeBetweenSends:          args.timeBetweenSends,
			timeBetweenSendsWhenError: args.timeBetweenSendsWhenError,
			thresholdBetweenSends:     args.thresholdBetweenSends,
			redundancyHandler:         args.redundancyHandler,
			privKey:                   args.privKey,
		},
		peerTypeProvider:     args.peerTypeProvider,
		versionNumber:        args.versionNumber,
		baseVersionNumber:    args.baseVersionNumber,
		nodeDisplayName:      args.nodeDisplayName,
		identity:             args.identity,
		peerSubType:          args.peerSubType,
		currentBlockProvider: args.currentBlockProvider,
		keysHolder:           args.keysHolder,
		shardCoordinator:     args.shardCoordinator,
	}

	return newMultikeyHeartbeatSender(argsSender)
}

func isMultikeyMode(privKey crypto.PrivateKey, keysHolder heartbeat.KeysHolder, nodesCoordinator heartbeat.NodesCoordinator) (bool, error) {
	pk := privKey.GeneratePublic()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return false, err
	}

	keysMap := keysHolder.GetManagedKeysByCurrentNode()
	isMultikey := len(keysMap) > 0

	_, _, err = nodesCoordinator.GetValidatorWithPublicKey(pkBytes)
	if err == nil && isMultikey {
		return false, heartbeat.ErrInvalidConfiguration
	}

	return isMultikey, nil
}
