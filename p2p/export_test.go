package p2p

import "github.com/ElrondNetwork/elrond-go-sandbox/marshal"

func (m *memoryMessage) GetMarshalizer() *marshal.Marshalizer {
	return m.marsh
}

func (m *memoryMessage) SetMarshalizer(newMarsh *marshal.Marshalizer) {
	m.marsh = newMarsh
}

func (mq *MessageQueue) Clean() {
	mq.clean()
}

func (m *memoryMessage) SetSigned(signed bool) {
	m.isSigned = signed
}

func (t *Topic) EventBusData() []DataReceivedHandler {
	return t.eventBusDataRcvHandlers
}

func (t *Topic) Marsh() marshal.Marshalizer {
	return t.marsh
}

func NewMemoryMessage(peerID string, payload []byte, mrsh marshal.Marshalizer) *memoryMessage {
	return newMemoryMessage(peerID, payload, mrsh)
}

func CreateMessageFromByteArray(mrsh marshal.Marshalizer, buff []byte) (*memoryMessage, error) {
	return createMessageFromByteArray(mrsh, buff)
}
