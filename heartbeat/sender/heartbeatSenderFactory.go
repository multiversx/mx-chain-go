package sender

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
)

type argHeartbeatSenderFactory struct {
	argBaseSender
	baseVersionNumber          string
	versionNumber              string
	nodeDisplayName            string
	identity                   string
	peerSubType                core.P2PPeerSubType
	currentBlockProvider       heartbeat.CurrentBlockProvider
	peerTypeProvider           heartbeat.PeerTypeProviderHandler
	managedPeersHolder         heartbeat.ManagedPeersHolder
	shardCoordinator           heartbeat.ShardCoordinator
	nodesCoordinator           heartbeat.NodesCoordinator
	trieSyncStatisticsProvider heartbeat.TrieSyncStatisticsProvider
}

func createHeartbeatSender(args argHeartbeatSenderFactory) (heartbeatSenderHandler, error) {
	isMultikey, err := isMultikeyMode(args.privKey, args.managedPeersHolder, args.nodesCoordinator)
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
			mainMessenger:             args.mainMessenger,
			fullArchiveMessenger:      args.fullArchiveMessenger,
			marshaller:                args.marshaller,
			topic:                     args.topic,
			timeBetweenSends:          args.timeBetweenSends,
			timeBetweenSendsWhenError: args.timeBetweenSendsWhenError,
			thresholdBetweenSends:     args.thresholdBetweenSends,
			redundancyHandler:         args.redundancyHandler,
			privKey:                   args.privKey,
		},
		versionNumber:              args.versionNumber,
		nodeDisplayName:            args.nodeDisplayName,
		identity:                   args.identity,
		peerSubType:                args.peerSubType,
		currentBlockProvider:       args.currentBlockProvider,
		peerTypeProvider:           args.peerTypeProvider,
		trieSyncStatisticsProvider: args.trieSyncStatisticsProvider,
	}

	return newHeartbeatSender(argsSender)
}

func createMultikeyHeartbeatSender(args argHeartbeatSenderFactory) (*multikeyHeartbeatSender, error) {
	argsSender := argMultikeyHeartbeatSender{
		argBaseSender: argBaseSender{
			mainMessenger:             args.mainMessenger,
			fullArchiveMessenger:      args.fullArchiveMessenger,
			marshaller:                args.marshaller,
			topic:                     args.topic,
			timeBetweenSends:          args.timeBetweenSends,
			timeBetweenSendsWhenError: args.timeBetweenSendsWhenError,
			thresholdBetweenSends:     args.thresholdBetweenSends,
			redundancyHandler:         args.redundancyHandler,
			privKey:                   args.privKey,
		},
		peerTypeProvider:           args.peerTypeProvider,
		versionNumber:              args.versionNumber,
		baseVersionNumber:          args.baseVersionNumber,
		nodeDisplayName:            args.nodeDisplayName,
		identity:                   args.identity,
		peerSubType:                args.peerSubType,
		currentBlockProvider:       args.currentBlockProvider,
		managedPeersHolder:         args.managedPeersHolder,
		shardCoordinator:           args.shardCoordinator,
		trieSyncStatisticsProvider: args.trieSyncStatisticsProvider,
	}

	return newMultikeyHeartbeatSender(argsSender)
}

func isMultikeyMode(privKey crypto.PrivateKey, managedPeersHolder heartbeat.ManagedPeersHolder, nodesCoordinator heartbeat.NodesCoordinator) (bool, error) {
	pk := privKey.GeneratePublic()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return false, err
	}

	isMultikey := managedPeersHolder.IsMultiKeyMode()

	_, _, err = nodesCoordinator.GetValidatorWithPublicKey(pkBytes)
	if err == nil && isMultikey {
		return false, fmt.Errorf("%w, isMultikey = %t, isValidator = %v", heartbeat.ErrInvalidConfiguration, isMultikey, err == nil)
	}

	return isMultikey, nil
}
