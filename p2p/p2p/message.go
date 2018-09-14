package p2p

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/pkg/errors"
)

type Message struct {
	marsh   *marshal.Marshalizer
	Payload []byte

	Hops  int
	Peers []string
}

func NewMessage(peerID string, payload []byte, mrsh marshal.Marshalizer) *Message {
	if mrsh == nil {
		panic("Nil marshalizer when creating a new Message!")
	}

	return &Message{Payload: payload, Hops: 0, Peers: []string{peerID}, marsh: &mrsh}
}

func (m *Message) ToByteArray() ([]byte, error) {
	if m.marsh == nil {
		return nil, errors.New("Uninitialized marshalizer!")
	}

	return (*m.marsh).Marshal(m)
}

func CreateFromByteArray(mrsh marshal.Marshalizer, buff []byte) (*Message, error) {
	m := &Message{}
	m.marsh = &mrsh

	if mrsh == nil {
		return nil, errors.New("Uninitialized marshalizer!")
	}

	err := mrsh.Unmarshal(m, buff)

	return m, err
}

func (m *Message) AddHop(peerID string) {
	m.Hops++
	m.Peers = append(m.Peers, peerID)
}
