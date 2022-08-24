package sender

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/covalent-indexer-go/process"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/keysManagement"
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
	keysHolder               keysManagement.KeysHolder
	timeBetweenChecks        time.Duration
	shardCoordinator         process.ShardCoordinator
}

type multikeyPeerAuthenticationSender struct {
	baseSender
	nodesCoordinator         heartbeat.NodesCoordinator
	peerSignatureHandler     crypto.PeerSignatureHandler
	hardforkTrigger          heartbeat.HardforkTrigger
	hardforkTimeBetweenSends time.Duration
	hardforkTriggerPubKey    []byte
	keysHolder               keysManagement.KeysHolder
	timeBetweenChecks        time.Duration
	shardCoordinator         process.ShardCoordinator
	getCurrentTimeHandler    func() time.Time
}

// newMultikeyPeerAuthenticationSender will create a new instance of type multikeyPeerAuthenticationSender
func newMultikeyPeerAuthenticationSender(args argMultikeyPeerAuthenticationSender) (*multikeyPeerAuthenticationSender, error) {
	err := checkMultikeyPeerAuthenticationSenderArgs(args)
	if err != nil {
		return nil, err
	}

	senderInstance := &multikeyPeerAuthenticationSender{
		baseSender:               createBaseSender(args.argBaseSender),
		nodesCoordinator:         args.nodesCoordinator,
		peerSignatureHandler:     args.peerSignatureHandler,
		hardforkTrigger:          args.hardforkTrigger,
		hardforkTimeBetweenSends: args.hardforkTimeBetweenSends,
		hardforkTriggerPubKey:    args.hardforkTriggerPubKey,
		keysHolder:               args.keysHolder,
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
	if check.IfNil(args.keysHolder) {
		return heartbeat.ErrNilKeysHolder
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
	managedKeys := sender.keysHolder.GetManagedKeysByCurrentNode()
	index := 0
	for pk, sk := range managedKeys {
		err := sender.process(index, pk, sk, currentTimeAsUnix)
		if err != nil {
			nextTimeToCheck, errNextPeerAuth := sender.keysHolder.GetNextPeerAuthenticationTime([]byte(pk))
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

func (sender *multikeyPeerAuthenticationSender) process(index int, pk string, sk crypto.PrivateKey, currentTimeAsUnix int64) error {
	pkBytes := []byte(pk)
	if !sender.processIfShouldSend(pkBytes, currentTimeAsUnix) {
		return nil
	}

	currentTimeStamp := time.Unix(currentTimeAsUnix, 0)

	data, isHardforkTriggered, err := sender.prepareMessage([]byte(pk), sk)
	if err != nil {
		setTimeErr := sender.keysHolder.SetNextPeerAuthenticationTime(pkBytes, currentTimeStamp.Add(sender.timeBetweenSendsWhenError))
		if setTimeErr != nil {
			return fmt.Errorf("%w while seting next peer authentication time, after prepare message error %s", setTimeErr, err.Error())
		}
		return err
	}
	if isHardforkTriggered {
		nextTimeStamp := currentTimeStamp.Add(sender.computeRandomDuration(sender.hardforkTimeBetweenSends))
		setTimeErr := sender.keysHolder.SetNextPeerAuthenticationTime(pkBytes, nextTimeStamp)
		if setTimeErr != nil {
			return fmt.Errorf("%w while seting next peer authentication time, hardfork triggered", setTimeErr)
		}
	} else {
		nextTimeStamp := currentTimeStamp.Add(sender.computeRandomDuration(sender.timeBetweenSends))
		setTimeErr := sender.keysHolder.SetNextPeerAuthenticationTime(pkBytes, nextTimeStamp)
		if setTimeErr != nil {
			return fmt.Errorf("%w while seting next peer authentication time", setTimeErr)
		}

		setValidatorErr := sender.keysHolder.SetValidatorState(pkBytes, true)
		if setValidatorErr != nil {
			return fmt.Errorf("%w while seting validator state", setValidatorErr)
		}
	}

	sender.sendData(index, pkBytes, data, isHardforkTriggered)

	return nil
}

func (sender *multikeyPeerAuthenticationSender) processIfShouldSend(pkBytes []byte, currentTimeAsUnix int64) bool {
	if !sender.keysHolder.IsKeyManagedByCurrentNode(pkBytes) {
		return false
	}
	isValidatorNow, shardID := sender.getIsValidatorStatusAndShardID(pkBytes)
	isHardforkSource := sender.isHardforkSource(pkBytes)
	oldIsValidator, err := sender.keysHolder.IsKeyValidator(pkBytes)
	if err != nil {
		return false
	}

	err = sender.keysHolder.SetValidatorState(pkBytes, isValidatorNow)
	if err != nil {
		return false
	}

	if !isValidatorNow && !isHardforkSource {
		return false
	}
	if shardID != sender.shardCoordinator.SelfId() {
		return false
	}

	nextTimeToCheck, err := sender.keysHolder.GetNextPeerAuthenticationTime(pkBytes)
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

func (sender *multikeyPeerAuthenticationSender) prepareMessage(pkBytes []byte, privateKey crypto.PrivateKey) ([]byte, bool, error) {
	p2pSkBytes, pid, err := sender.keysHolder.GetP2PIdentity(pkBytes)
	if err != nil {
		return nil, false, err
	}

	msg := &heartbeat.PeerAuthentication{
		Pid:    pid.Bytes(),
		Pubkey: pkBytes,
	}

	hardforkPayload, isTriggered := sender.getHardforkPayload()
	payload := &heartbeat.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: string(hardforkPayload),
	}
	payloadBytes, err := sender.marshaller.Marshal(payload)
	if err != nil {
		return nil, isTriggered, err
	}
	msg.Payload = payloadBytes
	msg.PayloadSignature, err = sender.messenger.SignWithPrivateKey(p2pSkBytes, payloadBytes)
	if err != nil {
		return nil, isTriggered, err
	}

	msg.Signature, err = sender.peerSignatureHandler.GetPeerSignature(privateKey, msg.Pid)
	if err != nil {
		return nil, isTriggered, err
	}

	msgBytes, err := sender.marshaller.Marshal(msg)
	if err != nil {
		return nil, isTriggered, err
	}

	b := &batch.Batch{
		Data: make([][]byte, 1),
	}
	b.Data[0] = msgBytes
	data, err := sender.marshaller.Marshal(b)
	if err != nil {
		return nil, isTriggered, err
	}

	return data, isTriggered, nil
}

func (sender *multikeyPeerAuthenticationSender) sendData(index int, pkBytes []byte, data []byte, isHardforkTriggered bool) {
	go func() {
		// extra delay as to avoid sending a lot of messages in the same time
		time.Sleep(time.Duration(index) * delayedBroadcast)

		p2pSk, pid, err := sender.keysHolder.GetP2PIdentity(pkBytes)
		if err != nil {
			log.Error("could not get identity for pk", "pk", hex.EncodeToString(pkBytes))
			return
		}
		sender.messenger.BroadcastWithSk(sender.topic, data, pid, p2pSk)

		nextTimeToCheck, err := sender.keysHolder.GetNextPeerAuthenticationTime(pkBytes)
		if err != nil {
			log.Error("could not get next peer authentication time for pk", "pk", hex.EncodeToString(pkBytes))
			return
		}

		log.Debug("peer authentication message sent",
			"bls pk", pkBytes,
			"pid", pid.Pretty(),
			"is hardfork triggered", isHardforkTriggered,
			"next send is scheduled on", nextTimeToCheck)
	}()
}

// ShouldTriggerHardfork signals when hardfork message should be sent
func (sender *multikeyPeerAuthenticationSender) ShouldTriggerHardfork() <-chan struct{} {
	return sender.hardforkTrigger.NotifyTriggerReceivedV2()
}

func (sender *multikeyPeerAuthenticationSender) getIsValidatorStatusAndShardID(pkBytes []byte) (bool, uint32) {
	_, shardID, err := sender.nodesCoordinator.GetValidatorWithPublicKey(pkBytes)
	return err == nil, shardID
}

func (sender *multikeyPeerAuthenticationSender) isHardforkSource(pkBytes []byte) bool {
	return bytes.Equal(pkBytes, sender.hardforkTriggerPubKey)
}

func (sender *multikeyPeerAuthenticationSender) getHardforkPayload() ([]byte, bool) {
	payload := make([]byte, 0)
	_, isTriggered := sender.hardforkTrigger.RecordedTriggerMessage()
	if isTriggered {
		payload = sender.hardforkTrigger.CreateData()
	}

	return payload, isTriggered
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *multikeyPeerAuthenticationSender) IsInterfaceNil() bool {
	return sender == nil
}
