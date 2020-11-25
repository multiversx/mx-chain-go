package process

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	heartbeatData "github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const delayAfterHardforkMessageBroadcast = time.Second * 5

// ArgHeartbeatSender represents the arguments for the heartbeat sender
type ArgHeartbeatSender struct {
	PeerMessenger        heartbeat.P2PMessenger
	PeerSignatureHandler crypto.PeerSignatureHandler
	PrivKey              crypto.PrivateKey
	Marshalizer          marshal.Marshalizer
	Topic                string
	ShardCoordinator     sharding.Coordinator
	PeerTypeProvider     heartbeat.PeerTypeProviderHandler
	PeerSubType          core.P2PPeerSubType
	StatusHandler        core.AppStatusHandler
	VersionNumber        string
	NodeDisplayName      string
	KeyBaseIdentity      string
	HardforkTrigger      heartbeat.HardforkTrigger
	CurrentBlockProvider heartbeat.CurrentBlockProvider
}

// Sender periodically sends heartbeat messages on a pubsub topic
type Sender struct {
	peerMessenger        heartbeat.P2PMessenger
	peerSignatureHandler crypto.PeerSignatureHandler
	privKey              crypto.PrivateKey
	marshalizer          marshal.Marshalizer
	shardCoordinator     sharding.Coordinator
	peerTypeProvider     heartbeat.PeerTypeProviderHandler
	peerSubType          core.P2PPeerSubType
	statusHandler        core.AppStatusHandler
	topic                string
	versionNumber        string
	nodeDisplayName      string
	keyBaseIdentity      string
	hardforkTrigger      heartbeat.HardforkTrigger
	currentBlockProvider heartbeat.CurrentBlockProvider
}

// NewSender will create a new sender instance
func NewSender(arg ArgHeartbeatSender) (*Sender, error) {
	if check.IfNil(arg.PeerMessenger) {
		return nil, heartbeat.ErrNilMessenger
	}
	if check.IfNil(arg.PeerSignatureHandler) {
		return nil, heartbeat.ErrNilPeerSignatureHandler
	}
	if check.IfNil(arg.PrivKey) {
		return nil, heartbeat.ErrNilPrivateKey
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, heartbeat.ErrNilMarshalizer
	}
	if check.IfNil(arg.ShardCoordinator) {
		return nil, heartbeat.ErrNilShardCoordinator
	}
	if check.IfNil(arg.PeerTypeProvider) {
		return nil, heartbeat.ErrNilPeerTypeProvider
	}
	if check.IfNil(arg.StatusHandler) {
		return nil, heartbeat.ErrNilAppStatusHandler
	}
	if check.IfNil(arg.HardforkTrigger) {
		return nil, heartbeat.ErrNilHardforkTrigger
	}
	if check.IfNil(arg.CurrentBlockProvider) {
		return nil, heartbeat.ErrNilCurrentBlockProvider
	}
	err := VerifyHeartbeatPropertyLen("application version string", []byte(arg.VersionNumber))
	if err != nil {
		return nil, err
	}

	sender := &Sender{
		peerMessenger:        arg.PeerMessenger,
		peerSignatureHandler: arg.PeerSignatureHandler,
		privKey:              arg.PrivKey,
		marshalizer:          arg.Marshalizer,
		topic:                arg.Topic,
		shardCoordinator:     arg.ShardCoordinator,
		peerTypeProvider:     arg.PeerTypeProvider,
		peerSubType:          arg.PeerSubType,
		statusHandler:        arg.StatusHandler,
		versionNumber:        arg.VersionNumber,
		nodeDisplayName:      arg.NodeDisplayName,
		keyBaseIdentity:      arg.KeyBaseIdentity,
		hardforkTrigger:      arg.HardforkTrigger,
		currentBlockProvider: arg.CurrentBlockProvider,
	}

	return sender, nil
}

// SendHeartbeat broadcasts a new heartbeat message
func (s *Sender) SendHeartbeat() error {
	nonce := uint64(0)
	crtBlock := s.currentBlockProvider.GetCurrentBlockHeader()
	if !check.IfNil(crtBlock) {
		nonce = crtBlock.GetNonce()
	}

	hb := &heartbeatData.Heartbeat{
		Payload:         []byte(fmt.Sprintf("%v", time.Now())),
		ShardID:         s.shardCoordinator.SelfId(),
		VersionNumber:   s.versionNumber,
		NodeDisplayName: s.nodeDisplayName,
		Identity:        s.keyBaseIdentity,
		Pid:             s.peerMessenger.ID().Bytes(),
		Nonce:           nonce,
		PeerSubType:     uint32(s.peerSubType),
	}

	triggerMessage, isHardforkTriggered := s.hardforkTrigger.RecordedTriggerMessage()
	if isHardforkTriggered {
		isPayloadRecorded := len(triggerMessage) != 0
		if isPayloadRecorded {
			//beside sending the regular heartbeat message, send also the initial payload hardfork trigger message
			// so that will be spread in an epidemic manner
			log.Debug("broadcasting stored hardfork message")
			s.peerMessenger.Broadcast(s.topic, triggerMessage)
			time.Sleep(delayAfterHardforkMessageBroadcast)
		} else {
			hb.Payload = s.hardforkTrigger.CreateData()
		}
	}

	log.Debug("broadcasting message", "is hardfork triggered", isHardforkTriggered)
	var err error
	hb.Pubkey, err = s.privKey.GeneratePublic().ToByteArray()
	if err != nil {
		return err
	}

	s.updateMetrics(hb)

	err = verifyLengths(hb)
	if err != nil {
		log.Warn("verify hb length", "error", err.Error())
		trimLengths(hb)
	}

	hb.Signature, err = s.peerSignatureHandler.GetPeerSignature(s.privKey, hb.Pid)
	if err != nil {
		return err
	}

	buffToSend, err := s.marshalizer.Marshal(hb)
	if err != nil {
		return err
	}

	s.peerMessenger.Broadcast(s.topic, buffToSend)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *Sender) IsInterfaceNil() bool {
	return s == nil
}

func (s *Sender) updateMetrics(hb *heartbeatData.Heartbeat) {
	result := s.computePeerList(hb.Pubkey)

	nodeType := ""
	if result == string(core.ObserverList) {
		nodeType = string(core.NodeTypeObserver)
	} else {
		nodeType = string(core.NodeTypeValidator)
	}

	subType := core.P2PPeerSubType(hb.PeerSubType)

	s.statusHandler.SetStringValue(core.MetricNodeType, nodeType)
	s.statusHandler.SetStringValue(core.MetricPeerType, result)
	s.statusHandler.SetStringValue(core.MetricPeerSubType, subType.String())
}

func (s *Sender) computePeerList(pubkey []byte) string {
	peerType, _, err := s.peerTypeProvider.ComputeForPubKey(pubkey)
	if err != nil {
		log.Warn("sender: compute peer type", "error", err)
		return string(core.ObserverList)
	}

	return string(peerType)
}
