package heartbeat

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// Sender periodically sends heartbeat messages on a pubsub topic
type Sender struct {
	peerMessenger    PeerMessenger
	singleSigner     crypto.SingleSigner
	privKey          crypto.PrivateKey
	marshalizer      marshal.Marshalizer
	topic            string
	shardCoordinator sharding.Coordinator
	versionNumber    string
	nodeDisplayName  string
}

// NewSender will create a new sender instance
func NewSender(
	peerMessenger PeerMessenger,
	singleSigner crypto.SingleSigner,
	privKey crypto.PrivateKey,
	marshalizer marshal.Marshalizer,
	topic string,
	shardCoordinator sharding.Coordinator,
	versionNumber string,
	nodeDisplayName string,
) (*Sender, error) {

	if peerMessenger == nil {
		return nil, ErrNilMessenger
	}
	if singleSigner == nil {
		return nil, ErrNilSingleSigner
	}
	if privKey == nil {
		return nil, ErrNilPrivateKey
	}
	if marshalizer == nil {
		return nil, ErrNilMarshalizer
	}
	//TODO add test for shardCoordinator

	sender := &Sender{
		peerMessenger:    peerMessenger,
		singleSigner:     singleSigner,
		privKey:          privKey,
		marshalizer:      marshalizer,
		topic:            topic,
		shardCoordinator: shardCoordinator,
		versionNumber:    versionNumber,
		nodeDisplayName:  nodeDisplayName,
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
	}

	var err error
	hb.Pubkey, err = s.privKey.GeneratePublic().ToByteArray()
	if err != nil {
		return err
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
