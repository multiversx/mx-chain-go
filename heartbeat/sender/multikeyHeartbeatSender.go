package sender

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/covalent-indexer-go/process"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

// argMultikeyHeartbeatSender represents the arguments for the heartbeat sender
type argMultikeyHeartbeatSender struct {
	argBaseSender
	peerTypeProvider     heartbeat.PeerTypeProviderHandler
	versionNumber        string
	baseVersionNumber    string
	nodeDisplayName      string
	identity             string
	peerSubType          core.P2PPeerSubType
	currentBlockProvider heartbeat.CurrentBlockProvider
	keysHolder           heartbeat.KeysHolder
	shardCoordinator     process.ShardCoordinator
}

type multikeyHeartbeatSender struct {
	commonHeartbeatSender
	peerTypeProvider  heartbeat.PeerTypeProviderHandler
	versionNumber     string
	baseVersionNumber string
	nodeDisplayName   string
	identity          string
	peerSubType       core.P2PPeerSubType
	keysHolder        heartbeat.KeysHolder
	shardCoordinator  process.ShardCoordinator
}

// newMultikeyHeartbeatSender creates a new instance of type multikeyHeartbeatSender
func newMultikeyHeartbeatSender(args argMultikeyHeartbeatSender) (*multikeyHeartbeatSender, error) {
	err := checkMultikeyHeartbeatSenderArgs(args)
	if err != nil {
		return nil, err
	}

	return &multikeyHeartbeatSender{
		commonHeartbeatSender: commonHeartbeatSender{
			baseSender:           createBaseSender(args.argBaseSender),
			currentBlockProvider: args.currentBlockProvider,
		},
		versionNumber:     args.versionNumber,
		baseVersionNumber: args.baseVersionNumber,
		nodeDisplayName:   args.nodeDisplayName,
		identity:          args.identity,
		peerSubType:       args.peerSubType,
		peerTypeProvider:  args.peerTypeProvider,
		keysHolder:        args.keysHolder,
		shardCoordinator:  args.shardCoordinator,
	}, nil
}

func checkMultikeyHeartbeatSenderArgs(args argMultikeyHeartbeatSender) error {
	err := checkBaseSenderArgs(args.argBaseSender)
	if err != nil {
		return err
	}
	if check.IfNil(args.peerTypeProvider) {
		return heartbeat.ErrNilPeerTypeProvider
	}
	if len(args.versionNumber) > maxSizeInBytes {
		return fmt.Errorf("%w for versionNumber, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, args.versionNumber, len(args.versionNumber), maxSizeInBytes)
	}
	if len(args.nodeDisplayName) > maxSizeInBytes {
		return fmt.Errorf("%w for nodeDisplayName, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, args.nodeDisplayName, len(args.nodeDisplayName), maxSizeInBytes)
	}
	if len(args.identity) > maxSizeInBytes {
		return fmt.Errorf("%w for identity, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, args.identity, len(args.identity), maxSizeInBytes)
	}
	if check.IfNil(args.currentBlockProvider) {
		return heartbeat.ErrNilCurrentBlockProvider
	}
	if check.IfNil(args.keysHolder) {
		return heartbeat.ErrNilKeysHolder
	}
	if check.IfNil(args.shardCoordinator) {
		return heartbeat.ErrNilShardCoordinator
	}

	return nil
}

// Execute will handle the execution of a cycle in which the heartbeat message will be sent
func (sender *multikeyHeartbeatSender) Execute() {
	duration := sender.computeRandomDuration(sender.timeBetweenSends)
	err := sender.execute()
	if err != nil {
		duration = sender.timeBetweenSendsWhenError
		log.Error("error sending heartbeat messages", "error", err, "next send will be in", duration)
	} else {
		log.Debug("heartbeat messages sent", "next send will be in", duration)
	}

	sender.CreateNewTimer(duration)
}

func (sender *multikeyHeartbeatSender) execute() error {
	_, pk := sender.getCurrentPrivateAndPublicKeys()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return err
	}

	buff, err := sender.generateMessageBytes(
		sender.versionNumber,
		sender.nodeDisplayName,
		sender.identity,
		uint32(sender.peerSubType),
		pkBytes,
	)
	if err != nil {
		return err
	}

	sender.messenger.Broadcast(sender.topic, buff)

	return sender.sendMultiKeysInfo()
}

func (sender *multikeyHeartbeatSender) sendMultiKeysInfo() error {
	managedKeys := sender.keysHolder.GetManagedKeysByCurrentNode()
	for pk := range managedKeys {
		pkBytes := []byte(pk)
		shouldSend := sender.processIfShouldSend(pkBytes)
		if !shouldSend {
			continue
		}

		err := sender.sendMessageForKey(pkBytes)
		if err != nil {
			log.Warn("could not broadcast for pk", "pk", pkBytes, "error", err)
		}
	}

	return nil
}

func (sender *multikeyHeartbeatSender) sendMessageForKey(pkBytes []byte) error {
	time.Sleep(delayedBroadcast)

	name, identity, err := sender.keysHolder.GetNameAndIdentity(pkBytes)
	if err != nil {
		return err
	}

	machineID, err := sender.keysHolder.GetMachineID(pkBytes)
	if err != nil {
		return err
	}
	versionNumber := fmt.Sprintf("%s/%s", sender.baseVersionNumber, machineID)

	buff, err := sender.generateMessageBytes(
		versionNumber,
		name,
		identity,
		uint32(core.RegularPeer), // force all in one peers to be of type regular peers
		pkBytes,
	)
	if err != nil {
		return err
	}

	p2pSk, pid, err := sender.keysHolder.GetP2PIdentity(pkBytes)
	if err != nil {
		return err
	}

	sender.messenger.BroadcastUsingPrivateKey(sender.topic, buff, pid, p2pSk)

	return nil
}

func (sender *multikeyHeartbeatSender) processIfShouldSend(pk []byte) bool {
	if !sender.keysHolder.IsKeyManagedByCurrentNode(pk) {
		return false
	}
	_, shardID, err := sender.peerTypeProvider.ComputeForPubKey(pk)
	if err != nil {
		log.Debug("processIfShouldSend.ComputeForPubKey", "error", err)
		return false
	}

	if shardID != sender.shardCoordinator.SelfId() {
		log.Debug("processIfShouldSend: shard id does not match",
			"pk", pk,
			"self shard", sender.shardCoordinator.SelfId(),
			"pk shard", shardID)
		return false
	}

	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *multikeyHeartbeatSender) IsInterfaceNil() bool {
	return sender == nil
}
