package memp2p

import "github.com/ElrondNetwork/elrond-go/p2p"

// Message represents a message to be sent through the in-memory network
// simulated by the Network struct.
type Message struct {
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

// NewMessage constructs a new Message instance from arguments
func NewMessage(topic string, data []byte, peerID p2p.PeerID) (*Message, error) {
	var empty []byte
	message := Message{
		from:      []byte(string(peerID)),
		data:      data,
		seqNo:     empty,
		topicIds:  []string{topic},
		signature: empty,
		key:       []byte(string(peerID)),
		peer:      peerID,
	}
	return &message, nil
}

// From returns the message originator's peer ID
func (message *Message) From() []byte {
	return message.from
}

// Data returns the message payload
func (message *Message) Data() []byte {
	return message.data
}

// SeqNo returns the message sequence number
func (message *Message) SeqNo() []byte {
	return message.seqNo
}

// TopicIDs returns the topic on which the message was sent
func (message *Message) TopicIDs() []string {
	return message.topicIds
}

// Signature returns the message signature
func (message *Message) Signature() []byte {
	return message.signature
}

// Key returns the message public key (if it can not be recovered from From field)
func (message *Message) Key() []byte {
	return message.key
}

// Peer returns the peer that originated the message
func (message *Message) Peer() p2p.PeerID {
	return message.peer
}

// IsInterfaceNil returns true if there is no value under the interface
func (message *Message) IsInterfaceNil() bool {
	if message == nil {
		return true
	}
	return false
}
