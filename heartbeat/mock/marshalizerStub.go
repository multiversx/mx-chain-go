package mock

// MarshalizerStub -
type MarshalizerStub struct {
	MarshalHandler   func(obj interface{}) ([]byte, error)
	UnmarshalHandler func(obj interface{}, buff []byte) error
}

// Marshal -
func (ms MarshalizerStub) Marshal(obj interface{}) ([]byte, error) {
	if ms.MarshalHandler != nil {
		return ms.MarshalHandler(obj)
	}
	return nil, nil
}

// Unmarshal -
func (ms MarshalizerStub) Unmarshal(obj interface{}, buff []byte) error {
	if ms.UnmarshalHandler != nil {
		return ms.UnmarshalHandler(obj, buff)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MarshalizerStub) IsInterfaceNil() bool {
	return ms == nil
}
