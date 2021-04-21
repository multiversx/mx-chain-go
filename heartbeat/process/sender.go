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
	StatusHandler        core.AppStatusHandler
	VersionNumber        string
	NodeDisplayName      string
	KeyBaseIdentity      string
	HardforkTrigger      heartbeat.HardforkTrigger
	CurrentBlockProvider heartbeat.CurrentBlockProvider
	RedundancyHandler    heartbeat.NodeRedundancyHandler
}

// Sender periodically sends heartbeat messages on a pubsub topic
type Sender struct {
	peerMessenger        heartbeat.P2PMessenger
	peerSignatureHandler crypto.PeerSignatureHandler
	privKey              crypto.PrivateKey
	publicKey            crypto.PublicKey
	observerPublicKey    crypto.PublicKey
	marshalizer          marshal.Marshalizer
	shardCoordinator     sharding.Coordinator
	peerTypeProvider     heartbeat.PeerTypeProviderHandler
	statusHandler        core.AppStatusHandler
	topic                string
	versionNumber        string
	nodeDisplayName      string
	keyBaseIdentity      string
	hardforkTrigger      heartbeat.HardforkTrigger
	currentBlockProvider heartbeat.CurrentBlockProvider
	redundancy           heartbeat.NodeRedundancyHandler
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
		return nil, fmt.Errorf("%w for arg.PrivKey", heartbeat.ErrNilPrivateKey)
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
	if check.IfNil(arg.RedundancyHandler) {
		return nil, heartbeat.ErrNilRedundancyHandler
	}
	err := VerifyHeartbeatProperyLen("application version string", []byte(arg.VersionNumber))
	if err != nil {
		return nil, err
	}

	observerPrivateKey := arg.RedundancyHandler.ObserverPrivateKey()
	if check.IfNil(observerPrivateKey) {
		return nil, fmt.Errorf("%w for arg.RedundancyHandler.ObserverPrivateKey()", heartbeat.ErrNilPrivateKey)
	}

	sender := &Sender{
		peerMessenger:        arg.PeerMessenger,
		peerSignatureHandler: arg.PeerSignatureHandler,
		privKey:              arg.PrivKey,
		publicKey:            arg.PrivKey.GeneratePublic(),
		observerPublicKey:    observerPrivateKey.GeneratePublic(),
		marshalizer:          arg.Marshalizer,
		topic:                arg.Topic,
		shardCoordinator:     arg.ShardCoordinator,
		peerTypeProvider:     arg.PeerTypeProvider,
		statusHandler:        arg.StatusHandler,
		versionNumber:        arg.VersionNumber,
		nodeDisplayName:      arg.NodeDisplayName,
		keyBaseIdentity:      arg.KeyBaseIdentity,
		hardforkTrigger:      arg.HardforkTrigger,
		currentBlockProvider: arg.CurrentBlockProvider,
		redundancy:           arg.RedundancyHandler,
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

	err := s.finalizeMessageConstruction(hb)
	if err != nil {
		return err
	}

	log.Debug("broadcasting message heartbeat message",
		"is hardfork triggered", isHardforkTriggered,
		"hex public key", hb.Pubkey,
	)

	buffToSend, err := s.marshalizer.Marshal(hb)
	if err != nil {
		return err
	}

	s.peerMessenger.Broadcast(s.topic, buffToSend)

	return nil
}

func (s *Sender) finalizeMessageConstruction(hb *heartbeatData.Heartbeat) error {
	sk, pk := s.getCurrentPrivateAndPublicKeys()

	var err error
	hb.Pubkey, err = pk.ToByteArray()
	if err != nil {
		return err
	}

	s.updateMetrics(hb)

	err = verifyLengths(hb)
	if err != nil {
		log.Warn("verify hb length", "error", err.Error())
		trimLengths(hb)
	}

	hb.Signature, err = s.peerSignatureHandler.GetPeerSignature(sk, hb.Pid)

	return err
}

func (s *Sender) getCurrentPrivateAndPublicKeys() (crypto.PrivateKey, crypto.PublicKey) {
	shouldUseOriginalKeys := !s.redundancy.IsRedundancyNode() || (s.redundancy.IsRedundancyNode() && !s.redundancy.IsMainMachineActive())
	if shouldUseOriginalKeys {
		return s.privKey, s.publicKey
	}

	return s.redundancy.ObserverPrivateKey(), s.observerPublicKey
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

	s.statusHandler.SetStringValue(core.MetricNodeType, nodeType)
	s.statusHandler.SetStringValue(core.MetricPeerType, result)
}

func (s *Sender) computePeerList(pubkey []byte) string {
	peerType, _, err := s.peerTypeProvider.ComputeForPubKey(pubkey)
	if err != nil {
		log.Warn("sender: compute peer type", "error", err)
		return string(core.ObserverList)
	}

	return string(peerType)
}
