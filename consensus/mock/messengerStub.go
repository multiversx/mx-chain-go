package mock

type MessengerStub struct {
	BroadcastCalled func(topic string, buff []byte)
}

func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	ms.BroadcastCalled(topic, buff)
}
