package marshallerMock

// MarshalizerStub -
type MarshalizerStub struct {
	MarshalCalled   func(obj interface{}) ([]byte, error)
	UnmarshalCalled func(obj interface{}, buff []byte) error
}

// Marshal -
func (ms *MarshalizerStub) Marshal(obj interface{}) ([]byte, error) {
	if ms.MarshalCalled != nil {
		return ms.MarshalCalled(obj)
	}
	return make([]byte, 0), nil
}

// Unmarshal -
func (ms *MarshalizerStub) Unmarshal(obj interface{}, buff []byte) error {
	if ms.UnmarshalCalled != nil {
		return ms.UnmarshalCalled(obj, buff)
	}
	return nil
}

// IsInterfaceNil -
func (ms *MarshalizerStub) IsInterfaceNil() bool {
	return ms == nil
}
