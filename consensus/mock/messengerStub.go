package mock

type MessengerStub struct {
	BroadcastCalled func(topic string, buff []byte)
}

func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessengerStub) IsInterfaceNil() bool {
	if ms == nil {
		return true
	}
	return false
}
