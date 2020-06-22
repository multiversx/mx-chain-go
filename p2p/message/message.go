package message

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var _ p2p.MessageP2P = (*Message)(nil)

// Message is a data holder struct
type Message struct {
	FromField      []byte
	DataField      []byte
	SeqNoField     []byte
	TopicsField    []string
	SignatureField []byte
	KeyField       []byte
	PeerField      core.PeerID
}

// From returns the message originator's peer ID
func (m *Message) From() []byte {
	return m.FromField
}

// Data returns the message payload
func (m *Message) Data() []byte {
	return m.DataField
}

// SeqNo returns the message sequence number
func (m *Message) SeqNo() []byte {
	return m.SeqNoField
}

// Topics returns the topic on which the message was sent
func (m *Message) Topics() []string {
	return m.TopicsField
}

// Signature returns the message signature
func (m *Message) Signature() []byte {
	return m.SignatureField
}

// Key returns the message public key (if it can not be recovered from From field)
func (m *Message) Key() []byte {
	return m.KeyField
}

// Peer returns the peer that originated the message
func (m *Message) Peer() core.PeerID {
	return m.PeerField
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *Message) IsInterfaceNil() bool {
	return m == nil
}
