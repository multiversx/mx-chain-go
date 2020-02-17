package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// P2PMessageStub -
type P2PMessageStub struct {
	FromField      []byte
	DataField      []byte
	SeqNoField     []byte
	TopicIDsField  []string
	SignatureField []byte
	KeyField       []byte
	PeerField      p2p.PeerID
}

// From -
func (msg *P2PMessageStub) From() []byte {
	return msg.FromField
}

// Data -
func (msg *P2PMessageStub) Data() []byte {
	return msg.DataField
}

// SeqNo -
func (msg *P2PMessageStub) SeqNo() []byte {
	return msg.SeqNoField
}

// TopicIDs -
func (msg *P2PMessageStub) TopicIDs() []string {
	return msg.TopicIDsField
}

// Signature -
func (msg *P2PMessageStub) Signature() []byte {
	return msg.SignatureField
}

// Key -
func (msg *P2PMessageStub) Key() []byte {
	return msg.KeyField
}

// Peer -
func (msg *P2PMessageStub) Peer() p2p.PeerID {
	return msg.PeerField
}

// IsInterfaceNil returns true if there is no value under the interface
func (msg *P2PMessageStub) IsInterfaceNil() bool {
	return msg == nil
}
