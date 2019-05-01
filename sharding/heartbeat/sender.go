package heartbeat

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

// Sender periodically sends hearbeat messages on a pubsub topic
type Sender struct {
	p2pMessenger P2PMessenger
	singleSigner crypto.SingleSigner
	privKey      crypto.PrivateKey
	marshalizer  marshal.Marshalizer
	topic        string
}

// NewSender will create a new sender instance
func NewSender(
	p2pMessenger P2PMessenger,
	singleSigner crypto.SingleSigner,
	privKey crypto.PrivateKey,
	marshalizer marshal.Marshalizer,
	topic string,
) (*Sender, error) {

	if p2pMessenger == nil {
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

	sender := &Sender{
		p2pMessenger: p2pMessenger,
		singleSigner: singleSigner,
		privKey:      privKey,
		marshalizer:  marshalizer,
		topic:        topic,
	}

	return sender, nil
}

// SendHeartbeat broadcasts a new heartbeat message
func (s *Sender) SendHeartbeat() error {

	hb := &Heartbeat{
		Payload: []byte(fmt.Sprintf("%v", time.Now())),
	}

	var err error
	hb.Pubkey, err = s.privKey.GeneratePublic().ToByteArray()
	if err != nil {
		return err
	}

	hb.Signature, err = s.singleSigner.Sign(s.privKey, hb.Payload)
	if err != nil {
		return err
	}

	buffToSend, err := s.marshalizer.Marshal(hb)
	if err != nil {
		return err
	}

	s.p2pMessenger.Broadcast(s.topic, buffToSend)
	return nil
}
