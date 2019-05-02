package mock

type MarshalizerStub struct {
	MarshalCalled   func(obj interface{}) ([]byte, error)
	UnmarshalCalled func(obj interface{}, buff []byte) error
}

func (ms *MarshalizerStub) Marshal(obj interface{}) ([]byte, error) {
	return ms.MarshalCalled(obj)
}

func (ms *MarshalizerStub) Unmarshal(obj interface{}, buff []byte) error {
	return ms.UnmarshalCalled(obj, buff)
}
