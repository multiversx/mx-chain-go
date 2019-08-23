package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type P2PMessageMock struct {
	FromField      []byte
	DataField      []byte
	SeqNoField     []byte
	TopicIDsField  []string
	SignatureField []byte
	KeyField       []byte
	PeerField      p2p.PeerID
}

func (msg *P2PMessageMock) From() []byte {
	return msg.FromField
}

func (msg *P2PMessageMock) Data() []byte {
	return msg.DataField
}

func (msg *P2PMessageMock) SeqNo() []byte {
	return msg.SeqNo()
}

func (msg *P2PMessageMock) TopicIDs() []string {
	return msg.TopicIDsField
}

func (msg *P2PMessageMock) Signature() []byte {
	return msg.SignatureField
}

func (msg *P2PMessageMock) Key() []byte {
	return msg.KeyField
}

func (msg *P2PMessageMock) Peer() p2p.PeerID {
	return msg.PeerField
}

// IsInterfaceNil returns true if there is no value under the interface
func (msg *P2PMessageMock) IsInterfaceNil() bool {
	if msg == nil {
		return true
	}
	return false
}
