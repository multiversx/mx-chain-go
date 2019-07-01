package memp2p

import "github.com/ElrondNetwork/elrond-go/p2p"

type MemP2PMessage struct {
	// sending PeerID, converted to []byte
	from []byte

	// the payload
	data []byte

	// leave empty
	seqNo []byte

	// topics set by the sender
	topicIds []string

	// leave empty
	signature []byte

	// sending PeerID, converted to []byte
	key []byte

	// sending PeerID
	peer p2p.PeerID
}

// NewMemP2PMessage constructs a new MemP2PMessage instance from arguments
func NewMemP2PMessage(topic string, data []byte, peerID p2p.PeerID) *MemP2PMessage {
	var empty []byte
	return &MemP2PMessage{
		from:      []byte(string(peerID)),
		data:      data,
		seqNo:     empty,
		topicIds:  []string{topic},
		signature: empty,
		key:       []byte(string(peerID)),
		peer:      peerID,
	}
}

// From returns the message originator's peer ID
func (message *MemP2PMessage) From() []byte {
	return message.from
}

// Data returns the message payload
func (message *MemP2PMessage) Data() []byte {
	return message.data
}

// SeqNo returns the message sequence number
func (message *MemP2PMessage) SeqNo() []byte {
	return message.seqNo
}

// TopicIDs returns the topic on which the message was sent
func (message *MemP2PMessage) TopicIDs() []string {
	return message.topicIds
}

// Signature returns the message signature
func (message *MemP2PMessage) Signature() []byte {
	return message.signature
}

// Key returns the message public key (if it can not be recovered from From field)
func (message *MemP2PMessage) Key() []byte {
	return message.key
}

// Peer returns the peer that originated the message
func (message *MemP2PMessage) Peer() p2p.PeerID {
	return message.peer
}
