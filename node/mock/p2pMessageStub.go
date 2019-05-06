package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type P2PMessageStub struct {
	FromField      []byte
	DataField      []byte
	SeqNoField     []byte
	TopicIDsField  []string
	SignatureField []byte
	KeyField       []byte
	PeerField      p2p.PeerID
}

func (msg *P2PMessageStub) From() []byte {
	return msg.FromField
}

func (msg *P2PMessageStub) Data() []byte {
	return msg.DataField
}

func (msg *P2PMessageStub) SeqNo() []byte {
	return msg.SeqNo()
}

func (msg *P2PMessageStub) TopicIDs() []string {
	return msg.TopicIDsField
}

func (msg *P2PMessageStub) Signature() []byte {
	return msg.SignatureField
}

func (msg *P2PMessageStub) Key() []byte {
	return msg.KeyField
}

func (msg *P2PMessageStub) Peer() p2p.PeerID {
	return msg.PeerField
}
