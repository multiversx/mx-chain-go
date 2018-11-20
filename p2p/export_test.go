package p2p

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/libp2p/go-libp2p-pubsub"
)

func (t *Topic) EventBusData() []DataReceivedHandler {
	return t.eventBusDataRcvHandlers
}

func (t *Topic) Marsh() marshal.Marshalizer {
	return t.marsh
}

func (t *Topic) SetRequest(f func(hash []byte) error) {
	t.request = f
}

func (t *Topic) SetRegisterTopicValidator(f func(v pubsub.Validator) error) {
	t.registerTopicValidator = f
}

func (t *Topic) SetUnregisterTopicValidator(f func() error) {
	t.unregisterTopicValidator = f
}

var DurTimeCache = durTimeCache
var MutGloballyRegPeers = &mutGloballyRegPeers
var GloballyRegisteredPeers = globallyRegisteredPeers
