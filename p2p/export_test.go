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

func (m *Message) SetSigned(signed bool) {
	m.isSigned = signed
}

func (t *Topic) EventBusData() []DataReceived {
	return t.eventBusData
}

func (t *Topic) Marsh() marshal.Marshalizer {
	return t.marsh
}
