package heartbeat

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ArgHeartbeatSender represents the arguments for the heartbeat sender
type ArgHeartbeatSender struct {
	PeerMessenger    PeerMessenger
	SingleSigner     crypto.SingleSigner
	PrivKey          crypto.PrivateKey
	Marshalizer      marshal.Marshalizer
	Topic            string
	ShardCoordinator sharding.Coordinator
	PeerTypeProvider PeerTypeProviderHandler
	StatusHandler    core.AppStatusHandler
	VersionNumber    string
	NodeDisplayName  string
	HardforkTrigger  HardforkTrigger
}

// Sender periodically sends heartbeat messages on a pubsub topic
type Sender struct {
	peerMessenger    PeerMessenger
	singleSigner     crypto.SingleSigner
	privKey          crypto.PrivateKey
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	peerTypeProvider PeerTypeProviderHandler
	statusHandler    core.AppStatusHandler
	topic            string
	versionNumber    string
	nodeDisplayName  string
	hardforkTrigger  HardforkTrigger
}

// NewSender will create a new sender instance
func NewSender(arg ArgHeartbeatSender) (*Sender, error) {
	if check.IfNil(arg.PeerMessenger) {
		return nil, ErrNilMessenger
	}
	if check.IfNil(arg.SingleSigner) {
		return nil, ErrNilSingleSigner
	}
	if check.IfNil(arg.PrivKey) {
		return nil, ErrNilPrivateKey
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(arg.ShardCoordinator) {
		return nil, ErrNilShardCoordinator
	}
	if check.IfNil(arg.PeerTypeProvider) {
		return nil, ErrNilPeerTypeProvider
	}
	if check.IfNil(arg.StatusHandler) {
		return nil, ErrNilAppStatusHandler
	}
	if check.IfNil(arg.HardforkTrigger) {
		return nil, ErrNilHardforkTrigger
	}
	err := VerifyHeartbeatProperyLen("application version string", []byte(arg.VersionNumber))
	if err != nil {
		return nil, err
	}

	sender := &Sender{
		peerMessenger:    arg.PeerMessenger,
		singleSigner:     arg.SingleSigner,
		privKey:          arg.PrivKey,
		marshalizer:      arg.Marshalizer,
		topic:            arg.Topic,
		shardCoordinator: arg.ShardCoordinator,
		peerTypeProvider: arg.PeerTypeProvider,
		statusHandler:    arg.StatusHandler,
		versionNumber:    arg.VersionNumber,
		nodeDisplayName:  arg.NodeDisplayName,
		hardforkTrigger:  arg.HardforkTrigger,
	}

	return sender, nil
}

// SendHeartbeat broadcasts a new heartbeat message
func (s *Sender) SendHeartbeat() error {
	hb := &Heartbeat{
		Payload:         []byte(fmt.Sprintf("%v", time.Now())),
		ShardID:         s.shardCoordinator.SelfId(),
		VersionNumber:   s.versionNumber,
		NodeDisplayName: s.nodeDisplayName,
		Pid:             s.peerMessenger.ID().Bytes(),
	}

	triggerMessage, isHardforkTriggered := s.hardforkTrigger.RecordedTriggerMessage()
	if isHardforkTriggered {
		isPayloadRecorder := len(triggerMessage) != 0
		if isPayloadRecorder {
			//beside sending the regular heartbeat message, send also the initial payload hardfork trigger message
			// so that will be spread in an epidemic manner
			s.peerMessenger.Broadcast(s.topic, triggerMessage)
		} else {
			hb.Payload = s.hardforkTrigger.CreateData()
		}
	}

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

	hbBytes, err := s.marshalizer.Marshal(hb)
	if err != nil {
		return err
	}

	hb.Signature, err = s.singleSigner.Sign(s.privKey, hbBytes)
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

func (s *Sender) updateMetrics(hb *Heartbeat) {
	result := s.computePeerList(hb.Pubkey)
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
