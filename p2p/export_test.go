package p2p

import "github.com/ElrondNetwork/elrond-go-sandbox/marshal"

func (m *Message) GetMarshalizer() *marshal.Marshalizer {
	return m.marsh
}

func (m *Message) SetMarshalizer(newMarsh *marshal.Marshalizer) {
	m.marsh = newMarsh
}

func (mq *MessageQueue) Clean() {
	mq.clean()
}
