package p2p

import (
	"github.com/libp2p/go-libp2p-pubsub"
)

// Message is a data holder struct
type Message struct {
	from      []byte
	data      []byte
	seqNo     []byte
	topicIds  []string
	signature []byte
	key       []byte
}

// NewMessage returns a new instance of a Message object
func NewMessage(message *pubsub.Message) *Message {
	return &Message{
		from:      message.From,
		data:      message.Data,
		seqNo:     message.Seqno,
		topicIds:  message.TopicIDs,
		signature: message.Signature,
		key:       message.Key,
	}
}

// From returns the message originator's peer ID
func (m *Message) From() []byte {
	return m.from
}

// Data returns the message payload
func (m *Message) Data() []byte {
	return m.data
}

// SeqNo returns the message sequence number
func (m *Message) SeqNo() []byte {
	return m.seqNo
}

// TopicIDs returns the topic on which the message was sent
func (m *Message) TopicIDs() []string {
	return m.topicIds
}

// Signature returns the message signature
func (m *Message) Signature() []byte {
	return m.signature
}

// Key returns the message public key (if it can not be recovered from From field)
func (m *Message) Key() []byte {
	return m.key
}
