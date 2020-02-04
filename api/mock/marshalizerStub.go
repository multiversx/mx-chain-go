package mock

// MarshalizerStub -
type MarshalizerStub struct {
	MarshalCalled   func(obj interface{}) ([]byte, error)
	UnmarshalCalled func(obj interface{}, buff []byte) error
}

// Marshal -
func (ms *MarshalizerStub) Marshal(obj interface{}) ([]byte, error) {
	return ms.MarshalCalled(obj)
}

// Unmarshal -
func (ms *MarshalizerStub) Unmarshal(obj interface{}, buff []byte) error {
	return ms.UnmarshalCalled(obj, buff)
}

// IsInterfaceNil -
func (ms *MarshalizerStub) IsInterfaceNil() bool {
	return ms == nil
}
