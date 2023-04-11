package sender

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/heartbeat"
)

const delayedBroadcast = 200 * time.Millisecond

// argMultikeyPeerAuthenticationSender represents the arguments for the peer authentication sender
type argMultikeyPeerAuthenticationSender struct {
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

type multikeyPeerAuthenticationSender struct {
	commonPeerAuthenticationSender
	hardforkTimeBetweenSends time.Duration
	managedPeersHolder       heartbeat.ManagedPeersHolder
	timeBetweenChecks        time.Duration
	shardCoordinator         heartbeat.ShardCoordinator
	getCurrentTimeHandler    func() time.Time
}

// newMultikeyPeerAuthenticationSender will create a new instance of type multikeyPeerAuthenticationSender
func newMultikeyPeerAuthenticationSender(args argMultikeyPeerAuthenticationSender) (*multikeyPeerAuthenticationSender, error) {
	err := checkMultikeyPeerAuthenticationSenderArgs(args)
	if err != nil {
		return nil, err
	}

	senderInstance := &multikeyPeerAuthenticationSender{
		commonPeerAuthenticationSender: commonPeerAuthenticationSender{
			baseSender:            createBaseSender(args.argBaseSender),
			nodesCoordinator:      args.nodesCoordinator,
			peerSignatureHandler:  args.peerSignatureHandler,
			hardforkTrigger:       args.hardforkTrigger,
			hardforkTriggerPubKey: args.hardforkTriggerPubKey,
		},
		hardforkTimeBetweenSends: args.hardforkTimeBetweenSends,
		managedPeersHolder:       args.managedPeersHolder,
		timeBetweenChecks:        args.timeBetweenChecks,
		shardCoordinator:         args.shardCoordinator,
		getCurrentTimeHandler:    getCurrentTime,
	}

	return senderInstance, nil
}

func getCurrentTime() time.Time {
	return time.Now()
}

func checkMultikeyPeerAuthenticationSenderArgs(args argMultikeyPeerAuthenticationSender) error {
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
	if check.IfNil(args.managedPeersHolder) {
		return heartbeat.ErrNilManagedPeersHolder
	}
	if args.timeBetweenChecks < minTimeBetweenSends {
		return fmt.Errorf("%w for timeBetweenChecks", heartbeat.ErrInvalidTimeDuration)
	}
	if check.IfNil(args.shardCoordinator) {
		return heartbeat.ErrNilShardCoordinator
	}

	return nil
}

// Execute will handle the execution of a cycle in which the peer authentication message will be sent
func (sender *multikeyPeerAuthenticationSender) Execute() {
	currentTimeAsUnix := sender.getCurrentTimeHandler().Unix()
	managedKeys := sender.managedPeersHolder.GetManagedKeysByCurrentNode()
	for pk, sk := range managedKeys {
		err := sender.process(pk, sk, currentTimeAsUnix)
		if err != nil {
			nextTimeToCheck, errNextPeerAuth := sender.managedPeersHolder.GetNextPeerAuthenticationTime([]byte(pk))
			if errNextPeerAuth != nil {
				log.Error("could not get next peer authentication time for pk", "pk", pk, "process error", err, "GetNextPeerAuthenticationTime error", errNextPeerAuth)
				return
			}

			log.Error("error sending peer authentication message", "bls pk", pk,
				"next send is scheduled on", nextTimeToCheck, "error", err)
		}
	}

	sender.CreateNewTimer(sender.timeBetweenChecks)
}

func (sender *multikeyPeerAuthenticationSender) process(pk string, sk crypto.PrivateKey, currentTimeAsUnix int64) error {
	pkBytes := []byte(pk)
	if !sender.processIfShouldSend(pkBytes, currentTimeAsUnix) {
		return nil
	}

	currentTimeStamp := time.Unix(currentTimeAsUnix, 0)

	data, isHardforkTriggered, _, err := sender.prepareMessage([]byte(pk), sk)
	if err != nil {
		sender.managedPeersHolder.SetNextPeerAuthenticationTime(pkBytes, currentTimeStamp.Add(sender.timeBetweenSendsWhenError))
		return err
	}
	if isHardforkTriggered {
		nextTimeStamp := currentTimeStamp.Add(sender.computeRandomDuration(sender.hardforkTimeBetweenSends))
		sender.managedPeersHolder.SetNextPeerAuthenticationTime(pkBytes, nextTimeStamp)
	} else {
		nextTimeStamp := currentTimeStamp.Add(sender.computeRandomDuration(sender.timeBetweenSends))
		sender.managedPeersHolder.SetNextPeerAuthenticationTime(pkBytes, nextTimeStamp)
		sender.managedPeersHolder.SetValidatorState(pkBytes, true)
	}

	sender.sendData(pkBytes, data, isHardforkTriggered)

	return nil
}

func (sender *multikeyPeerAuthenticationSender) processIfShouldSend(pkBytes []byte, currentTimeAsUnix int64) bool {
	if !sender.managedPeersHolder.IsKeyManagedByCurrentNode(pkBytes) {
		return false
	}
	isValidatorNow, shardID := sender.getIsValidatorStatusAndShardID(pkBytes)
	isHardforkSource := sender.isHardforkSource(pkBytes)
	oldIsValidator := sender.managedPeersHolder.IsKeyValidator(pkBytes)
	sender.managedPeersHolder.SetValidatorState(pkBytes, isValidatorNow)

	if !isValidatorNow && !isHardforkSource {
		return false
	}
	if shardID != sender.shardCoordinator.SelfId() {
		return false
	}

	nextTimeToCheck, err := sender.managedPeersHolder.GetNextPeerAuthenticationTime(pkBytes)
	if err != nil {
		return false
	}

	timeToCheck := nextTimeToCheck.Unix() < currentTimeAsUnix
	if timeToCheck {
		return true
	}
	if !oldIsValidator && isValidatorNow {
		return true
	}

	return false
}

func (sender *multikeyPeerAuthenticationSender) prepareMessage(pkBytes []byte, privateKey crypto.PrivateKey) ([]byte, bool, int64, error) {
	p2pSkBytes, pid, err := sender.managedPeersHolder.GetP2PIdentity(pkBytes)
	if err != nil {
		return nil, false, 0, err
	}

	return sender.generateMessageBytes(pkBytes, privateKey, p2pSkBytes, pid.Bytes())
}

func (sender *multikeyPeerAuthenticationSender) sendData(pkBytes []byte, data []byte, isHardforkTriggered bool) {
	// extra delay as to avoid sending a lot of messages in the same time
	time.Sleep(delayedBroadcast)

	p2pSk, pid, err := sender.managedPeersHolder.GetP2PIdentity(pkBytes)
	if err != nil {
		log.Error("could not get identity for pk", "pk", hex.EncodeToString(pkBytes), "error", err)
		return
	}
	sender.messenger.BroadcastUsingPrivateKey(sender.topic, data, pid, p2pSk)

	nextTimeToCheck, err := sender.managedPeersHolder.GetNextPeerAuthenticationTime(pkBytes)
	if err != nil {
		log.Error("could not get next peer authentication time for pk", "pk", hex.EncodeToString(pkBytes), "error", err)
		return
	}

	log.Debug("peer authentication message sent",
		"bls pk", pkBytes,
		"pid", pid.Pretty(),
		"is hardfork triggered", isHardforkTriggered,
		"next send is scheduled on", nextTimeToCheck)
}

// ShouldTriggerHardfork signals when hardfork message should be sent
func (sender *multikeyPeerAuthenticationSender) ShouldTriggerHardfork() <-chan struct{} {
	return sender.hardforkTrigger.NotifyTriggerReceivedV2()
}

func (sender *multikeyPeerAuthenticationSender) getIsValidatorStatusAndShardID(pkBytes []byte) (bool, uint32) {
	_, shardID, err := sender.nodesCoordinator.GetValidatorWithPublicKey(pkBytes)
	return err == nil, shardID
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *multikeyPeerAuthenticationSender) IsInterfaceNil() bool {
	return sender == nil
}
